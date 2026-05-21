package shipment

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/order"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

var (
	shipmentTenantA      = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shipmentStoreA       = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	shipmentTenantB      = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shipmentStoreB       = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	shipmentOrderA       = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	shipmentOrderB       = uuid.MustParse("66666666-6666-6666-6666-666666666666")
	shipmentPendingOrder = uuid.MustParse("77777777-7777-7777-7777-777777777777")
	shipmentActorID      = uuid.MustParse("88888888-8888-8888-8888-888888888888")
)

func TestTenantCannotCreateShipmentForOtherTenantOrder(t *testing.T) {
	service, _, _, _ := newShipmentTestService()

	_, err := service.Create(context.Background(), shipmentTenantA, shipmentStoreA, shipmentOrderB, validCreateInput())
	assertShipmentAppErrorCode(t, err, apperror.CodeNotFound)
}

func TestInvalidOrderStatusRejected(t *testing.T) {
	service, _, _, _ := newShipmentTestService()

	_, err := service.Create(context.Background(), shipmentTenantA, shipmentStoreA, shipmentPendingOrder, validCreateInput())
	assertShipmentAppErrorCode(t, err, apperror.CodeInvalidOrderStatus)
}

func TestValidShipmentCreatesStatusLog(t *testing.T) {
	service, repo, _, outboxRepo := newShipmentTestService()

	result, err := service.Create(context.Background(), shipmentTenantA, shipmentStoreA, shipmentOrderA, validCreateInput())
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if result.ID == uuid.Nil || result.Status != StatusPending {
		t.Fatalf("create result = %#v", result)
	}
	logs := repo.logs[result.ID]
	if len(logs) != 1 || logs[0].ToStatus != StatusPending {
		t.Fatalf("shipment logs = %#v", logs)
	}
	if repo.orders[shipmentOrderA].ShipmentStatus != StatusPending {
		t.Fatalf("order shipment_status = %s, want %s", repo.orders[shipmentOrderA].ShipmentStatus, StatusPending)
	}
	if !outboxRepo.hasEvent(EventShipmentCreated) || !outboxRepo.hasEvent(EventNotificationRequested) {
		t.Fatalf("outbox events = %#v", outboxRepo.events)
	}
}

func TestInvalidShipmentTransitionRejected(t *testing.T) {
	service, _, _, _ := newShipmentTestService()
	created, err := service.Create(context.Background(), shipmentTenantA, shipmentStoreA, shipmentOrderA, validCreateInput())
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}

	_, err = service.UpdateStatus(context.Background(), shipmentTenantA, shipmentStoreA, created.ID, UpdateStatusInput{
		ActorUserID: shipmentActorID,
		Status:      StatusDelivered,
	})
	assertShipmentAppErrorCode(t, err, apperror.CodeInvalidOrderStatus)
}

func TestDeliveredShipmentUpdatesOrderState(t *testing.T) {
	service, repo, _, outboxRepo := newShipmentTestService()
	created, err := service.Create(context.Background(), shipmentTenantA, shipmentStoreA, shipmentOrderA, validCreateInput())
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}

	for _, status := range []string{StatusReadyForPickup, StatusPickedUp, StatusOnDelivery, StatusDelivered} {
		_, err = service.UpdateStatus(context.Background(), shipmentTenantA, shipmentStoreA, created.ID, UpdateStatusInput{
			ActorUserID: shipmentActorID,
			Status:      status,
		})
		if err != nil {
			t.Fatalf("UpdateStatus(%s) error = %v", status, err)
		}
	}

	if repo.orders[shipmentOrderA].Status != order.StatusDelivered {
		t.Fatalf("order status = %s, want %s", repo.orders[shipmentOrderA].Status, order.StatusDelivered)
	}
	if repo.orders[shipmentOrderA].ShipmentStatus != StatusDelivered {
		t.Fatalf("shipment_status = %s, want %s", repo.orders[shipmentOrderA].ShipmentStatus, StatusDelivered)
	}
	if len(repo.orderLogs) != 1 || repo.orderLogs[0].ToStatus != order.StatusDelivered {
		t.Fatalf("order logs = %#v", repo.orderLogs)
	}
	if !outboxRepo.hasEvent(EventOrderDelivered) {
		t.Fatalf("outbox events = %#v", outboxRepo.events)
	}
}

func TestPublicTrackingRejectsWrongPhone(t *testing.T) {
	service, _, _, _ := newShipmentTestService()

	_, err := service.PublicTracking(context.Background(), "toko-a", "ORD-TEST-A", "089999999999")
	assertShipmentAppErrorCode(t, err, apperror.CodeNotFound)
}

func TestPublicTrackingHidesInternalFields(t *testing.T) {
	service, _, _, _ := newShipmentTestService()
	createInput := validCreateInput()
	createInput.Note = "Private delivery note"
	createInput.AssignedToPhone = "081999999999"
	createInput.TrackingNumber = "TRK-PRIVATE-1"

	_, err := service.Create(context.Background(), shipmentTenantA, shipmentStoreA, shipmentOrderA, createInput)
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}

	result, err := service.PublicTracking(context.Background(), "toko-a", "ORD-TEST-A", "081111111111")
	if err != nil {
		t.Fatalf("PublicTracking error = %v", err)
	}
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal public response: %v", err)
	}
	payload := string(raw)
	for _, forbidden := range []string{"tenant_id", "store_id", "Private delivery note", "081999999999", "created_by", "updated_by"} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("public tracking leaked %q in %s", forbidden, payload)
		}
	}
	if result.Shipment == nil || result.Shipment.TrackingNumber != "TRK-PRIVATE-1" {
		t.Fatalf("public shipment = %#v", result.Shipment)
	}
}

func TestShipmentPermissionDefaultDenyRemainsIntact(t *testing.T) {
	tests := []struct {
		name       string
		role       permission.Role
		permission permission.Permission
		want       bool
	}{
		{name: "owner can create shipment", role: permission.RoleOwner, permission: permission.ShipmentCreate, want: true},
		{name: "manager can update shipment", role: permission.RoleManager, permission: permission.ShipmentUpdateStatus, want: true},
		{name: "staff can read shipment", role: permission.RoleStaff, permission: permission.ShipmentRead, want: true},
		{name: "courier admin can update shipment", role: permission.RoleCourierAdmin, permission: permission.ShipmentUpdateStatus, want: true},
		{name: "cashier cannot read shipment", role: permission.RoleCashier, permission: permission.ShipmentRead, want: false},
		{name: "inventory cannot create shipment", role: permission.RoleInventoryStaff, permission: permission.ShipmentCreate, want: false},
		{name: "driver denied until assigned-driver rule exists", role: permission.RoleDriver, permission: permission.ShipmentRead, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := permission.Allowed(string(tt.role), tt.permission); got != tt.want {
				t.Fatalf("Allowed(%q, %q) = %v, want %v", tt.role, tt.permission, got, tt.want)
			}
		})
	}
}

func validCreateInput() CreateInput {
	return CreateInput{
		ActorUserID:     shipmentActorID,
		CourierType:     CourierTypeManual,
		CourierName:     "Kurir Lokal",
		ShippingCost:    15000,
		AssignedToName:  "Udin",
		AssignedToPhone: "081222222222",
	}
}

func newShipmentTestService() (*Service, *fakeShipmentRepository, *fakeShipmentPublicStores, *fakeShipmentOutboxRepository) {
	now := time.Date(2026, 5, 21, 8, 0, 0, 0, time.UTC)
	repo := &fakeShipmentRepository{
		now:       now,
		orders:    map[uuid.UUID]order.Order{},
		items:     map[uuid.UUID][]order.Item{},
		shipments: map[uuid.UUID]Shipment{},
		logs:      map[uuid.UUID][]StatusLog{},
	}
	repo.orders[shipmentOrderA] = order.Order{
		ID:            shipmentOrderA,
		TenantID:      shipmentTenantA,
		StoreID:       shipmentStoreA,
		OrderNumber:   "ORD-TEST-A",
		Status:        order.StatusConfirmed,
		PaymentStatus: order.PaymentStatusPaid,
		Subtotal:      50000,
		ShippingCost:  15000,
		GrandTotal:    65000,
		CustomerName:  "Andi",
		CustomerPhone: "081111111111",
		InternalNote:  "Do not expose this",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	repo.orders[shipmentOrderB] = order.Order{
		ID:            shipmentOrderB,
		TenantID:      shipmentTenantB,
		StoreID:       shipmentStoreB,
		OrderNumber:   "ORD-TEST-B",
		Status:        order.StatusConfirmed,
		PaymentStatus: order.PaymentStatusPaid,
		CustomerName:  "Budi",
		CustomerPhone: "082222222222",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	repo.orders[shipmentPendingOrder] = order.Order{
		ID:            shipmentPendingOrder,
		TenantID:      shipmentTenantA,
		StoreID:       shipmentStoreA,
		OrderNumber:   "ORD-PENDING",
		Status:        order.StatusPending,
		PaymentStatus: order.PaymentStatusUnpaid,
		CustomerName:  "Cici",
		CustomerPhone: "083333333333",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	repo.items[shipmentOrderA] = []order.Item{
		{
			ID:          uuid.New(),
			TenantID:    shipmentTenantA,
			OrderID:     shipmentOrderA,
			ProductName: "Bouquet Mawar",
			Quantity:    1,
			UnitPrice:   50000,
			Subtotal:    50000,
			CreatedAt:   now,
		},
	}

	publicStores := &fakeShipmentPublicStores{stores: map[string]store.PublicContext{
		"toko-a": {
			TenantID: shipmentTenantA,
			StoreID:  shipmentStoreA,
			Store: store.Store{
				ID:       shipmentStoreA,
				TenantID: shipmentTenantA,
				Slug:     "toko-a",
				Status:   store.StatusPublished,
			},
		},
	}}
	outboxRepo := &fakeShipmentOutboxRepository{}
	service := NewService(fakeShipmentDB{}, repo, publicStores, outboxRepo)
	service.now = func() time.Time { return now }
	return service, repo, publicStores, outboxRepo
}

func assertShipmentAppErrorCode(t *testing.T, err error, code apperror.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %s, got nil", code)
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error %s, got %T: %v", code, err, err)
	}
	if appErr.Code != code {
		t.Fatalf("error code = %s, want %s", appErr.Code, code)
	}
}

type fakeShipmentDB struct{}

func (fakeShipmentDB) WithTx(_ context.Context, fn func(tx db.Tx) error) error {
	return fn(fakeShipmentDB{})
}

func (fakeShipmentDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (fakeShipmentDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (fakeShipmentDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

type fakeShipmentRepository struct {
	orders    map[uuid.UUID]order.Order
	items     map[uuid.UUID][]order.Item
	shipments map[uuid.UUID]Shipment
	logs      map[uuid.UUID][]StatusLog
	orderLogs []fakeOrderStatusLog
	now       time.Time
}

type fakeOrderStatusLog struct {
	OrderID    uuid.UUID
	FromStatus string
	ToStatus   string
	Note       string
	CreatedBy  uuid.UUID
}

func (f *fakeShipmentRepository) List(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, filters ListFilters) ([]Shipment, error) {
	items := make([]Shipment, 0)
	for _, shipment := range f.shipments {
		if shipment.TenantID != tenantID || shipment.StoreID != storeID {
			continue
		}
		if filters.Status != nil && shipment.Status != *filters.Status {
			continue
		}
		items = append(items, shipment)
	}
	return items, nil
}

func (f *fakeShipmentRepository) FindByID(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, shipmentID uuid.UUID) (*Shipment, error) {
	shipmentRecord, ok := f.shipments[shipmentID]
	if !ok || shipmentRecord.TenantID != tenantID || shipmentRecord.StoreID != storeID {
		return nil, ErrShipmentNotFound
	}
	return cloneShipment(shipmentRecord), nil
}

func (f *fakeShipmentRepository) LockByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, shipmentID uuid.UUID) (*Shipment, error) {
	return f.FindByID(ctx, q, tenantID, storeID, shipmentID)
}

func (f *fakeShipmentRepository) FindLatestByOrder(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (*Shipment, error) {
	var latest *Shipment
	for _, shipmentRecord := range f.shipments {
		if shipmentRecord.TenantID != tenantID || shipmentRecord.StoreID != storeID || shipmentRecord.OrderID != orderID {
			continue
		}
		if latest == nil || shipmentRecord.CreatedAt.After(latest.CreatedAt) {
			latest = cloneShipment(shipmentRecord)
		}
	}
	if latest == nil {
		return nil, ErrShipmentNotFound
	}
	return latest, nil
}

func (f *fakeShipmentRepository) Create(_ context.Context, _ db.Queryer, params CreateShipmentParams) (*Shipment, error) {
	orderRecord, ok := f.orders[params.OrderID]
	if !ok || orderRecord.TenantID != params.TenantID || orderRecord.StoreID != params.StoreID {
		return nil, ErrOrderNotFound
	}
	shipmentRecord := Shipment{
		ID:              uuid.New(),
		TenantID:        params.TenantID,
		StoreID:         params.StoreID,
		OrderID:         params.OrderID,
		OrderNumber:     orderRecord.OrderNumber,
		CustomerName:    orderRecord.CustomerName,
		CustomerPhone:   orderRecord.CustomerPhone,
		CourierType:     params.CourierType,
		CourierName:     params.CourierName,
		TrackingNumber:  params.TrackingNumber,
		Status:          StatusPending,
		ShippingCost:    params.ShippingCost,
		AssignedToName:  params.AssignedToName,
		AssignedToPhone: params.AssignedToPhone,
		Note:            params.Note,
		CreatedBy:       &params.CreatedBy,
		UpdatedBy:       &params.CreatedBy,
		CreatedAt:       f.now,
		UpdatedAt:       f.now,
	}
	f.shipments[shipmentRecord.ID] = shipmentRecord
	return cloneShipment(shipmentRecord), nil
}

func (f *fakeShipmentRepository) UpdateStatus(_ context.Context, _ db.Queryer, params UpdateShipmentStatusParams) (*Shipment, error) {
	shipmentRecord, ok := f.shipments[params.ShipmentID]
	if !ok || shipmentRecord.TenantID != params.TenantID || shipmentRecord.StoreID != params.StoreID {
		return nil, ErrShipmentNotFound
	}
	shipmentRecord.Status = params.Status
	shipmentRecord.UpdatedBy = &params.UpdatedBy
	shipmentRecord.UpdatedAt = f.now.Add(time.Minute)
	if params.Status == StatusPickedUp || params.Status == StatusOnDelivery || params.Status == StatusDelivered {
		shippedAt := f.now.Add(time.Minute)
		shipmentRecord.ShippedAt = &shippedAt
	}
	if params.Status == StatusDelivered {
		deliveredAt := f.now.Add(2 * time.Minute)
		shipmentRecord.DeliveredAt = &deliveredAt
	}
	f.shipments[params.ShipmentID] = shipmentRecord
	return cloneShipment(shipmentRecord), nil
}

func (f *fakeShipmentRepository) CreateStatusLog(_ context.Context, _ db.Queryer, params CreateStatusLogParams) (*StatusLog, error) {
	log := StatusLog{
		ID:         uuid.New(),
		TenantID:   params.TenantID,
		ShipmentID: params.ShipmentID,
		FromStatus: params.FromStatus,
		ToStatus:   params.ToStatus,
		Note:       params.Note,
		CreatedBy:  &params.CreatedBy,
		CreatedAt:  f.now,
	}
	f.logs[params.ShipmentID] = append(f.logs[params.ShipmentID], log)
	return &log, nil
}

func (f *fakeShipmentRepository) ListStatusLogs(_ context.Context, _ db.Queryer, tenantID uuid.UUID, shipmentID uuid.UUID) ([]StatusLog, error) {
	logs := make([]StatusLog, 0)
	for _, log := range f.logs[shipmentID] {
		if log.TenantID == tenantID {
			logs = append(logs, log)
		}
	}
	return logs, nil
}

func (f *fakeShipmentRepository) LockOrderByID(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (*order.Order, error) {
	orderRecord, ok := f.orders[orderID]
	if !ok || orderRecord.TenantID != tenantID || orderRecord.StoreID != storeID {
		return nil, ErrOrderNotFound
	}
	return cloneOrder(orderRecord), nil
}

func (f *fakeShipmentRepository) FindPublicOrderByNumber(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderNumber string) (*order.Order, error) {
	for _, orderRecord := range f.orders {
		if orderRecord.TenantID == tenantID && orderRecord.StoreID == storeID && orderRecord.OrderNumber == orderNumber {
			return cloneOrder(orderRecord), nil
		}
	}
	return nil, ErrOrderNotFound
}

func (f *fakeShipmentRepository) ListOrderItems(_ context.Context, _ db.Queryer, tenantID uuid.UUID, orderID uuid.UUID) ([]order.Item, error) {
	items := make([]order.Item, 0)
	for _, item := range f.items[orderID] {
		if item.TenantID == tenantID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (f *fakeShipmentRepository) UpdateOrderShipmentStatus(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID, status string) error {
	orderRecord, ok := f.orders[orderID]
	if !ok || orderRecord.TenantID != tenantID || orderRecord.StoreID != storeID {
		return ErrOrderNotFound
	}
	orderRecord.ShipmentStatus = status
	orderRecord.UpdatedAt = f.now.Add(time.Minute)
	f.orders[orderID] = orderRecord
	return nil
}

func (f *fakeShipmentRepository) UpdateOrderStatus(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID, status string) (*order.Order, error) {
	orderRecord, ok := f.orders[orderID]
	if !ok || orderRecord.TenantID != tenantID || orderRecord.StoreID != storeID {
		return nil, ErrOrderNotFound
	}
	orderRecord.Status = status
	orderRecord.UpdatedAt = f.now.Add(time.Minute)
	f.orders[orderID] = orderRecord
	return cloneOrder(orderRecord), nil
}

func (f *fakeShipmentRepository) CreateOrderStatusLog(_ context.Context, _ db.Queryer, tenantID uuid.UUID, orderID uuid.UUID, fromStatus string, toStatus string, note string, createdBy uuid.UUID) error {
	orderRecord, ok := f.orders[orderID]
	if !ok || orderRecord.TenantID != tenantID {
		return ErrOrderNotFound
	}
	f.orderLogs = append(f.orderLogs, fakeOrderStatusLog{
		OrderID:    orderID,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		Note:       note,
		CreatedBy:  createdBy,
	})
	return nil
}

func cloneShipment(item Shipment) *Shipment {
	clone := item
	return &clone
}

func cloneOrder(item order.Order) *order.Order {
	clone := item
	return &clone
}

type fakeShipmentPublicStores struct {
	stores map[string]store.PublicContext
}

func (f *fakeShipmentPublicStores) Resolve(_ context.Context, slug string) (store.PublicContext, error) {
	currentStore, ok := f.stores[slug]
	if !ok {
		return store.PublicContext{}, apperror.NotFound("Store not found")
	}
	return currentStore, nil
}

type fakeShipmentOutboxRepository struct {
	events []outbox.InsertEventParams
}

func (f *fakeShipmentOutboxRepository) Insert(_ context.Context, _ db.Queryer, params outbox.InsertEventParams) (*outbox.Event, error) {
	if !json.Valid(params.Payload) {
		return nil, errors.New("invalid json payload")
	}
	f.events = append(f.events, params)
	return &outbox.Event{}, nil
}

func (f *fakeShipmentOutboxRepository) hasEvent(eventType string) bool {
	for _, event := range f.events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}
