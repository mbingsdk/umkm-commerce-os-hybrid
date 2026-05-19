package checkout

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/catalog/product"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/inventory"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/order"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/idempotency"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

const (
	defaultReservationTTL = 24 * time.Hour
	idempotencyLockTTL    = 5 * time.Minute
)

type txRunner interface {
	WithTx(ctx context.Context, fn func(tx db.Tx) error) error
}

type storeResolver interface {
	Resolve(ctx context.Context, slug string) (store.PublicContext, error)
}

type idempotencyStore interface {
	Begin(
		ctx context.Context,
		q db.Queryer,
		tenantID uuid.UUID,
		scope string,
		key string,
		requestHash string,
		lockedUntil time.Time,
	) (*idempotency.State, error)
	SaveCompletedResponse(
		ctx context.Context,
		q db.Queryer,
		tenantID uuid.UUID,
		scope string,
		key string,
		statusCode int,
		responseBody json.RawMessage,
	) error
}

type checkoutStore interface {
	ListProductsForCheckout(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productIDs []uuid.UUID) ([]ProductForCheckout, error)
	LockStockSnapshots(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productIDs []uuid.UUID) ([]StockSnapshot, error)
	FindOrCreateCustomer(ctx context.Context, q db.Queryer, params FindOrCreateCustomerParams) (*CustomerRecord, error)
	CreateCustomerAddress(ctx context.Context, q db.Queryer, params CreateAddressParams) (*AddressRecord, error)
	CreateOrder(ctx context.Context, q db.Queryer, params CreateOrderParams) (*OrderRecord, error)
	CreateOrderItem(ctx context.Context, q db.Queryer, params CreateOrderItemParams) error
	CreateStockReservation(ctx context.Context, q db.Queryer, params CreateReservationParams) error
	UpdateStockSnapshot(ctx context.Context, q db.Queryer, params UpdateSnapshotParams) error
	CreateStockMovement(ctx context.Context, q db.Queryer, params CreateStockMovementParams) error
	CreateOrderStatusLog(ctx context.Context, q db.Queryer, tenantID uuid.UUID, orderID uuid.UUID, toStatus string, note string) error
	UpdateCustomerStats(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, customerID uuid.UUID, orderTotal int64) error
}

type outboxStore interface {
	Insert(ctx context.Context, q db.Queryer, params outbox.InsertEventParams) (*outbox.Event, error)
}

type Service struct {
	tx             txRunner
	stores         storeResolver
	repository     checkoutStore
	idempotency    idempotencyStore
	outbox         outboxStore
	now            func() time.Time
	newUUID        func() uuid.UUID
	reservationTTL time.Duration
}

func NewService(
	tx txRunner,
	stores storeResolver,
	repository checkoutStore,
	idempotency idempotencyStore,
	outbox outboxStore,
) *Service {
	return &Service{
		tx:             tx,
		stores:         stores,
		repository:     repository,
		idempotency:    idempotency,
		outbox:         outbox,
		now:            time.Now,
		newUUID:        uuid.New,
		reservationTTL: defaultReservationTTL,
	}
}

type Command struct {
	StoreSlug      string
	IdempotencyKey string
	Method         string
	Path           string
	RawBody        []byte
	Request        CheckoutRequest
}

type normalizedItem struct {
	ProductID uuid.UUID
	Quantity  int
}

type normalizedRequest struct {
	Items          []normalizedItem
	CustomerName   string
	CustomerPhone  string
	CustomerEmail  string
	AddressLabel   string
	RecipientName  string
	RecipientPhone string
	Address        string
	City           string
	Province       string
	PostalCode     string
	PaymentMethod  string
	CustomerNote   string
}

func (s *Service) Checkout(ctx context.Context, cmd Command) (CheckoutResult, error) {
	idempotencyKey := strings.TrimSpace(cmd.IdempotencyKey)
	if idempotencyKey == "" {
		return CheckoutResult{}, apperror.Validation("Validation failed", []map[string]string{
			{"field": "Idempotency-Key", "message": "header is required"},
		})
	}

	currentStore, err := s.stores.Resolve(ctx, cmd.StoreSlug)
	if err != nil {
		return CheckoutResult{}, err
	}

	requestHash, err := idempotency.RequestHash(cmd.Method, cmd.Path, cmd.RawBody)
	if err != nil {
		return CheckoutResult{}, apperror.Validation("Invalid JSON payload", []map[string]string{
			{"field": "body", "message": err.Error()},
		})
	}

	var result CheckoutResult
	err = s.tx.WithTx(ctx, func(tx db.Tx) error {
		state, err := s.idempotency.Begin(
			ctx,
			tx,
			currentStore.TenantID,
			idempotency.ScopeCheckout,
			idempotencyKey,
			requestHash,
			s.now().UTC().Add(idempotencyLockTTL),
		)
		if err != nil {
			return err
		}
		if state.CanReplay {
			var response CheckoutResponse
			if err := json.Unmarshal(state.ResponseBody, &response); err != nil {
				return apperror.Internal(err)
			}
			result = CheckoutResult{
				Response:   response,
				StatusCode: state.StatusCode,
			}
			if result.StatusCode == 0 {
				result.StatusCode = http.StatusCreated
			}
			return nil
		}
		if state.IsProcessing && !state.Created {
			return apperror.Conflict("Checkout request is still processing")
		}

		normalized, err := normalizeRequest(cmd.Request)
		if err != nil {
			return err
		}

		response, err := s.createCheckout(ctx, tx, currentStore, normalized)
		if err != nil {
			return err
		}

		responseBody, err := json.Marshal(response)
		if err != nil {
			return apperror.Internal(err)
		}
		if err := s.idempotency.SaveCompletedResponse(
			ctx,
			tx,
			currentStore.TenantID,
			idempotency.ScopeCheckout,
			idempotencyKey,
			http.StatusCreated,
			responseBody,
		); err != nil {
			return err
		}

		result = CheckoutResult{
			Response:   response,
			StatusCode: http.StatusCreated,
		}
		return nil
	})
	if err != nil {
		return CheckoutResult{}, err
	}

	return result, nil
}

func (s *Service) createCheckout(
	ctx context.Context,
	tx db.Tx,
	currentStore store.PublicContext,
	request normalizedRequest,
) (CheckoutResponse, error) {
	productIDs := productIDsFromItems(request.Items)

	snapshots, err := s.repository.LockStockSnapshots(ctx, tx, currentStore.TenantID, currentStore.StoreID, productIDs)
	if err != nil {
		return CheckoutResponse{}, apperror.Internal(err)
	}
	snapshotByProduct := make(map[uuid.UUID]StockSnapshot, len(snapshots))
	for _, snapshot := range snapshots {
		snapshotByProduct[snapshot.ProductID] = snapshot
	}

	products, err := s.repository.ListProductsForCheckout(ctx, tx, currentStore.TenantID, currentStore.StoreID, productIDs)
	if err != nil {
		return CheckoutResponse{}, apperror.Internal(err)
	}
	productByID := make(map[uuid.UUID]ProductForCheckout, len(products))
	for _, item := range products {
		productByID[item.ID] = item
	}

	var subtotal int64
	for _, item := range request.Items {
		productRecord, ok := productByID[item.ProductID]
		if !ok ||
			productRecord.TenantID != currentStore.TenantID ||
			productRecord.StoreID != currentStore.StoreID ||
			productRecord.Status != product.StatusActive {
			return CheckoutResponse{}, apperror.NotFound("Product not found")
		}

		lineSubtotal, err := multiplyMoney(productRecord.Price, item.Quantity)
		if err != nil {
			return CheckoutResponse{}, err
		}
		if subtotal > math.MaxInt64-lineSubtotal {
			return CheckoutResponse{}, apperror.Validation("Validation failed", []map[string]string{
				{"field": "items", "message": "order total is too large"},
			})
		}
		subtotal += lineSubtotal

		if productRecord.TrackInventory {
			snapshot, ok := snapshotByProduct[item.ProductID]
			if !ok || snapshot.TenantID != currentStore.TenantID || snapshot.StoreID != currentStore.StoreID {
				return CheckoutResponse{}, apperror.InsufficientStock("Insufficient stock", []map[string]any{
					{"product_id": item.ProductID.String(), "available": 0, "requested": item.Quantity},
				})
			}
			if snapshot.QuantityAvailable < item.Quantity {
				return CheckoutResponse{}, apperror.InsufficientStock("Insufficient stock", []map[string]any{
					{"product_id": item.ProductID.String(), "available": snapshot.QuantityAvailable, "requested": item.Quantity},
				})
			}
		}
	}

	shippingCost := int64(0)
	discountTotal := int64(0)
	taxTotal := int64(0)
	grandTotal := subtotal + shippingCost + taxTotal - discountTotal

	customerRecord, err := s.repository.FindOrCreateCustomer(ctx, tx, FindOrCreateCustomerParams{
		TenantID: currentStore.TenantID,
		StoreID:  currentStore.StoreID,
		Name:     request.CustomerName,
		Phone:    request.CustomerPhone,
		Email:    request.CustomerEmail,
	})
	if err != nil {
		return CheckoutResponse{}, apperror.Internal(err)
	}

	if _, err := s.repository.CreateCustomerAddress(ctx, tx, CreateAddressParams{
		TenantID:       currentStore.TenantID,
		CustomerID:     customerRecord.ID,
		Label:          request.AddressLabel,
		RecipientName:  request.RecipientName,
		RecipientPhone: request.RecipientPhone,
		Address:        request.Address,
		City:           request.City,
		Province:       request.Province,
		PostalCode:     request.PostalCode,
	}); err != nil {
		return CheckoutResponse{}, apperror.Internal(err)
	}

	now := s.now().UTC()
	orderNumber := generateOrderNumber(now, s.newUUID())
	orderRecord, err := s.repository.CreateOrder(ctx, tx, CreateOrderParams{
		TenantID:           currentStore.TenantID,
		StoreID:            currentStore.StoreID,
		CustomerID:         customerRecord.ID,
		OrderNumber:        orderNumber,
		Source:             order.SourceStorefront,
		Status:             order.StatusPending,
		PaymentStatus:      order.PaymentStatusUnpaid,
		Subtotal:           subtotal,
		DiscountTotal:      discountTotal,
		ShippingCost:       shippingCost,
		TaxTotal:           taxTotal,
		GrandTotal:         grandTotal,
		CustomerName:       request.CustomerName,
		CustomerPhone:      request.CustomerPhone,
		CustomerEmail:      request.CustomerEmail,
		ShippingAddress:    request.Address,
		ShippingCity:       request.City,
		ShippingProvince:   request.Province,
		ShippingPostalCode: request.PostalCode,
		CustomerNote:       request.CustomerNote,
	})
	if err != nil {
		return CheckoutResponse{}, apperror.Internal(err)
	}

	reservationExpiresAt := now.Add(s.reservationTTL)
	stockEventItems := make([]map[string]any, 0, len(request.Items))
	for _, item := range request.Items {
		productRecord := productByID[item.ProductID]
		lineSubtotal, err := multiplyMoney(productRecord.Price, item.Quantity)
		if err != nil {
			return CheckoutResponse{}, err
		}
		if err := s.repository.CreateOrderItem(ctx, tx, CreateOrderItemParams{
			TenantID:      currentStore.TenantID,
			OrderID:       orderRecord.ID,
			ProductID:     productRecord.ID,
			ProductName:   productRecord.Name,
			SKU:           productRecord.SKU,
			Quantity:      item.Quantity,
			UnitPrice:     productRecord.Price,
			DiscountTotal: 0,
			Subtotal:      lineSubtotal,
		}); err != nil {
			return CheckoutResponse{}, apperror.Internal(err)
		}

		if !productRecord.TrackInventory {
			continue
		}

		snapshot := snapshotByProduct[item.ProductID]
		newReserved := snapshot.QuantityReserved + item.Quantity
		newAvailable := snapshot.QuantityAvailable - item.Quantity
		if err := s.repository.CreateStockReservation(ctx, tx, CreateReservationParams{
			TenantID:  currentStore.TenantID,
			StoreID:   currentStore.StoreID,
			ProductID: productRecord.ID,
			OrderID:   orderRecord.ID,
			Quantity:  item.Quantity,
			ExpiresAt: &reservationExpiresAt,
		}); err != nil {
			return CheckoutResponse{}, apperror.Internal(err)
		}
		if err := s.repository.UpdateStockSnapshot(ctx, tx, UpdateSnapshotParams{
			TenantID:          currentStore.TenantID,
			StoreID:           currentStore.StoreID,
			ProductID:         productRecord.ID,
			QuantityReserved:  newReserved,
			QuantityAvailable: newAvailable,
		}); err != nil {
			return CheckoutResponse{}, apperror.Internal(err)
		}
		if err := s.repository.CreateStockMovement(ctx, tx, CreateStockMovementParams{
			TenantID:      currentStore.TenantID,
			StoreID:       currentStore.StoreID,
			ProductID:     productRecord.ID,
			MovementType:  inventory.MovementTypeReserved,
			Quantity:      item.Quantity,
			BalanceAfter:  newAvailable,
			ReferenceType: "order",
			ReferenceID:   orderRecord.ID,
			Note:          "Stock reserved for storefront checkout",
		}); err != nil {
			return CheckoutResponse{}, apperror.Internal(err)
		}

		stockEventItems = append(stockEventItems, map[string]any{
			"product_id": productRecord.ID.String(),
			"quantity":   item.Quantity,
		})
	}

	if err := s.repository.CreateOrderStatusLog(
		ctx,
		tx,
		currentStore.TenantID,
		orderRecord.ID,
		order.StatusPending,
		"Order created from public storefront checkout",
	); err != nil {
		return CheckoutResponse{}, apperror.Internal(err)
	}

	if err := s.repository.UpdateCustomerStats(ctx, tx, currentStore.TenantID, currentStore.StoreID, customerRecord.ID, grandTotal); err != nil {
		return CheckoutResponse{}, apperror.Internal(err)
	}

	if err := s.insertOutboxEvents(ctx, tx, currentStore, orderRecord, stockEventItems); err != nil {
		return CheckoutResponse{}, err
	}

	return CheckoutResponse{
		OrderID:       orderRecord.ID,
		OrderNumber:   orderRecord.OrderNumber,
		Status:        orderRecord.Status,
		PaymentStatus: orderRecord.PaymentStatus,
		Totals: CheckoutTotalsResponse{
			Subtotal:      orderRecord.Subtotal,
			DiscountTotal: orderRecord.DiscountTotal,
			ShippingCost:  orderRecord.ShippingCost,
			TaxTotal:      orderRecord.TaxTotal,
			GrandTotal:    orderRecord.GrandTotal,
		},
		PaymentInstruction: CheckoutPaymentInstructions{
			Method:  request.PaymentMethod,
			Message: "Pesanan berhasil dibuat. Instruksi pembayaran manual akan dikonfirmasi oleh penjual.",
		},
	}, nil
}

func (s *Service) insertOutboxEvents(
	ctx context.Context,
	tx db.Tx,
	currentStore store.PublicContext,
	orderRecord *OrderRecord,
	stockItems []map[string]any,
) error {
	orderPayload, err := json.Marshal(map[string]any{
		"tenant_id":    currentStore.TenantID.String(),
		"store_id":     currentStore.StoreID.String(),
		"order_id":     orderRecord.ID.String(),
		"order_number": orderRecord.OrderNumber,
		"grand_total":  orderRecord.GrandTotal,
	})
	if err != nil {
		return apperror.Internal(err)
	}

	stockPayload, err := json.Marshal(map[string]any{
		"tenant_id": currentStore.TenantID.String(),
		"store_id":  currentStore.StoreID.String(),
		"order_id":  orderRecord.ID.String(),
		"items":     stockItems,
	})
	if err != nil {
		return apperror.Internal(err)
	}

	notificationPayload, err := json.Marshal(map[string]any{
		"tenant_id": currentStore.TenantID.String(),
		"store_id":  currentStore.StoreID.String(),
		"order_id":  orderRecord.ID.String(),
		"type":      "order_created_manual_payment",
	})
	if err != nil {
		return apperror.Internal(err)
	}

	for _, event := range []outbox.InsertEventParams{
		{
			TenantID:      currentStore.TenantID,
			EventType:     EventOrderCreated,
			AggregateType: AggregateOrder,
			AggregateID:   orderRecord.ID,
			Payload:       orderPayload,
		},
		{
			TenantID:      currentStore.TenantID,
			EventType:     EventStockReserved,
			AggregateType: AggregateOrder,
			AggregateID:   orderRecord.ID,
			Payload:       stockPayload,
		},
		{
			TenantID:      currentStore.TenantID,
			EventType:     EventNotificationRequested,
			AggregateType: AggregateOrder,
			AggregateID:   orderRecord.ID,
			Payload:       notificationPayload,
		},
	} {
		if _, err := s.outbox.Insert(ctx, tx, event); err != nil {
			return apperror.Internal(err)
		}
	}

	return nil
}

func normalizeRequest(request CheckoutRequest) (normalizedRequest, error) {
	details := make([]map[string]string, 0)
	if len(request.Items) == 0 {
		details = append(details, map[string]string{"field": "items", "message": "items must not be empty"})
	}

	quantityByProduct := make(map[uuid.UUID]int)
	for idx, item := range request.Items {
		fieldPrefix := fmt.Sprintf("items[%d]", idx)
		if item.ProductID == uuid.Nil {
			details = append(details, map[string]string{"field": fieldPrefix + ".product_id", "message": "product_id is required"})
			continue
		}
		if item.Quantity <= 0 {
			details = append(details, map[string]string{"field": fieldPrefix + ".quantity", "message": "quantity must be greater than zero"})
			continue
		}
		quantityByProduct[item.ProductID] += item.Quantity
	}

	customerName := strings.TrimSpace(request.Customer.Name)
	customerPhone := strings.TrimSpace(request.Customer.Phone)
	customerEmail := strings.TrimSpace(request.Customer.Email)
	if customerName == "" {
		details = append(details, map[string]string{"field": "customer.name", "message": "name is required"})
	}
	if customerPhone == "" {
		details = append(details, map[string]string{"field": "customer.phone", "message": "phone is required"})
	}

	address := strings.TrimSpace(request.ShippingAddress.Address)
	if address == "" {
		details = append(details, map[string]string{"field": "shipping_address.address", "message": "address is required"})
	}

	paymentMethod := strings.TrimSpace(request.PaymentMethod)
	if paymentMethod == "" {
		paymentMethod = PaymentMethodManualTransfer
	}
	if paymentMethod != PaymentMethodManualTransfer {
		details = append(details, map[string]string{"field": "payment_method", "message": "only manual_transfer is supported for MVP checkout"})
	}

	if len(details) > 0 {
		return normalizedRequest{}, apperror.Validation("Validation failed", details)
	}

	recipientName := strings.TrimSpace(request.ShippingAddress.RecipientName)
	if recipientName == "" {
		recipientName = customerName
	}
	recipientPhone := strings.TrimSpace(request.ShippingAddress.RecipientPhone)
	if recipientPhone == "" {
		recipientPhone = customerPhone
	}

	items := make([]normalizedItem, 0, len(quantityByProduct))
	for productID, quantity := range quantityByProduct {
		items = append(items, normalizedItem{
			ProductID: productID,
			Quantity:  quantity,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return strings.Compare(items[i].ProductID.String(), items[j].ProductID.String()) < 0
	})

	return normalizedRequest{
		Items:          items,
		CustomerName:   customerName,
		CustomerPhone:  customerPhone,
		CustomerEmail:  customerEmail,
		AddressLabel:   strings.TrimSpace(request.ShippingAddress.Label),
		RecipientName:  recipientName,
		RecipientPhone: recipientPhone,
		Address:        address,
		City:           strings.TrimSpace(request.ShippingAddress.City),
		Province:       strings.TrimSpace(request.ShippingAddress.Province),
		PostalCode:     strings.TrimSpace(request.ShippingAddress.PostalCode),
		PaymentMethod:  paymentMethod,
		CustomerNote:   strings.TrimSpace(request.CustomerNote),
	}, nil
}

func productIDsFromItems(items []normalizedItem) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ProductID)
	}
	return ids
}

func multiplyMoney(unitPrice int64, quantity int) (int64, error) {
	if unitPrice < 0 {
		return 0, apperror.Validation("Validation failed", []map[string]string{
			{"field": "items", "message": "product price is invalid"},
		})
	}
	if quantity <= 0 {
		return 0, apperror.Validation("Validation failed", []map[string]string{
			{"field": "items.quantity", "message": "quantity must be greater than zero"},
		})
	}
	if unitPrice > math.MaxInt64/int64(quantity) {
		return 0, apperror.Validation("Validation failed", []map[string]string{
			{"field": "items", "message": "line total is too large"},
		})
	}
	return unitPrice * int64(quantity), nil
}

func generateOrderNumber(now time.Time, id uuid.UUID) string {
	compactID := strings.ReplaceAll(id.String(), "-", "")
	if len(compactID) > 12 {
		compactID = compactID[:12]
	}
	return fmt.Sprintf("ORD-%s-%s", now.Format("20060102"), strings.ToUpper(compactID))
}
