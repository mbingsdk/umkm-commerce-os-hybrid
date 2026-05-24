package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/password"
)

const (
	demoPassword = "demopassword"
)

type demoSeed struct {
	passwordHash string
}

type demoProduct struct {
	Name           string
	Slug           string
	Description    string
	SKU            string
	CategorySlug   string
	Price          int64
	CompareAtPrice *int64
	CostPrice      int64
	InitialStock   int
	LowStock       int
	IsDiscoverable bool
	WeightGram     int
}

type demoZone struct {
	Name        string
	Description string
	Rate        int64
	SortOrder   int
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	_ = godotenv.Load(".env")
	_ = godotenv.Load(filepath.Join("backend", ".env"))

	if !envBool("DEMO_SEED_ENABLED", true) {
		log.Println("demo seed skipped because DEMO_SEED_ENABLED=false")
		return nil
	}

	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if appEnv == "production" && !envBool("DEMO_SEED_ALLOW_PRODUCTION", false) {
		return errors.New("refusing to run demo seed in production; set DEMO_SEED_ALLOW_PRODUCTION=true only for an approved disposable environment")
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	database, err := db.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer database.Close()

	hasher := password.NewBcryptHasher()
	hash, err := hasher.Hash(envString("DEMO_SEED_PASSWORD", demoPassword))
	if err != nil {
		return fmt.Errorf("hash demo password: %w", err)
	}

	seed := demoSeed{passwordHash: hash}
	if err := database.WithTx(ctx, func(tx db.Tx) error {
		return seed.run(ctx, tx)
	}); err != nil {
		return err
	}

	log.Println("demo seed completed for Toko Bunga Ayu; data is non-production")
	return nil
}

func (s demoSeed) run(ctx context.Context, tx db.Tx) error {
	plans, err := s.upsertPlans(ctx, tx)
	if err != nil {
		return err
	}

	ownerID, err := s.upsertUser(ctx, tx, userParams{
		Name:         "Owner Demo Toko Bunga Ayu",
		Email:        "owner.demo@umkm.test",
		Phone:        "081234567890",
		PlatformRole: "user",
	})
	if err != nil {
		return err
	}

	staffID, err := s.upsertUser(ctx, tx, userParams{
		Name:         "Staff Demo Toko Bunga Ayu",
		Email:        "staff.demo@umkm.test",
		Phone:        "081234567891",
		PlatformRole: "user",
	})
	if err != nil {
		return err
	}

	cashierID, err := s.upsertUser(ctx, tx, userParams{
		Name:         "Kasir Demo Toko Bunga Ayu",
		Email:        "cashier.demo@umkm.test",
		Phone:        "081234567892",
		PlatformRole: "user",
	})
	if err != nil {
		return err
	}

	if envBool("DEMO_SEED_SUPER_ADMIN_ENABLED", false) {
		if _, err := s.upsertUser(ctx, tx, userParams{
			Name:         "Super Admin Demo",
			Email:        "superadmin.demo@umkm.test",
			Phone:        "081234567899",
			PlatformRole: "super_admin",
		}); err != nil {
			return err
		}
	}

	tenantID, err := s.upsertTenant(ctx, tx, plans["growth"])
	if err != nil {
		return err
	}

	storeID, err := s.upsertStore(ctx, tx, tenantID)
	if err != nil {
		return err
	}

	members := []struct {
		userID string
		role   string
	}{
		{ownerID, "owner"},
		{staffID, "staff"},
		{cashierID, "cashier"},
	}
	for _, member := range members {
		if err := s.upsertMembership(ctx, tx, member.userID, tenantID, member.role); err != nil {
			return err
		}
	}

	categories, err := s.upsertCategories(ctx, tx, tenantID, storeID)
	if err != nil {
		return err
	}

	if err := s.upsertProducts(ctx, tx, tenantID, storeID, ownerID, categories); err != nil {
		return err
	}

	if err := s.upsertCourierZones(ctx, tx, tenantID, storeID); err != nil {
		return err
	}

	return nil
}

func (s demoSeed) upsertPlans(ctx context.Context, tx db.Tx) (map[string]string, error) {
	type plan struct {
		Code            string
		Name            string
		Description     string
		PriceMonthly    int64
		ProductLimit    any
		StaffLimit      any
		CanUsePOS       bool
		CanUseDiscovery bool
		CanUseCourier   bool
	}

	plans := []plan{
		{
			Code:            "starter",
			Name:            "Starter",
			Description:     "Paket demo Starter untuk validasi awal UMKM.",
			PriceMonthly:    99000,
			ProductLimit:    100,
			StaffLimit:      2,
			CanUsePOS:       true,
			CanUseDiscovery: true,
			CanUseCourier:   false,
		},
		{
			Code:            "growth",
			Name:            "Growth",
			Description:     "Paket demo Growth untuk toko pilot dengan POS, discovery, dan courier.",
			PriceMonthly:    199000,
			ProductLimit:    1000,
			StaffLimit:      5,
			CanUsePOS:       true,
			CanUseDiscovery: true,
			CanUseCourier:   true,
		},
		{
			Code:            "business",
			Name:            "Business",
			Description:     "Paket demo Business dengan limit lebih besar.",
			PriceMonthly:    499000,
			ProductLimit:    10000,
			StaffLimit:      20,
			CanUsePOS:       true,
			CanUseDiscovery: true,
			CanUseCourier:   true,
		},
	}

	result := make(map[string]string, len(plans))
	const query = `
		INSERT INTO plans (
			code,
			name,
			description,
			price_monthly,
			product_limit,
			staff_limit,
			can_use_pos,
			can_use_discovery,
			can_use_courier,
			is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true)
		ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			price_monthly = EXCLUDED.price_monthly,
			product_limit = EXCLUDED.product_limit,
			staff_limit = EXCLUDED.staff_limit,
			can_use_pos = EXCLUDED.can_use_pos,
			can_use_discovery = EXCLUDED.can_use_discovery,
			can_use_courier = EXCLUDED.can_use_courier,
			is_active = true,
			updated_at = now()
		RETURNING id
	`
	for _, plan := range plans {
		var id string
		if err := tx.QueryRow(
			ctx,
			query,
			plan.Code,
			plan.Name,
			plan.Description,
			plan.PriceMonthly,
			plan.ProductLimit,
			plan.StaffLimit,
			plan.CanUsePOS,
			plan.CanUseDiscovery,
			plan.CanUseCourier,
		).Scan(&id); err != nil {
			return nil, fmt.Errorf("upsert plan %s: %w", plan.Code, err)
		}
		result[plan.Code] = id
	}

	return result, nil
}

type userParams struct {
	Name         string
	Email        string
	Phone        string
	PlatformRole string
}

func (s demoSeed) upsertUser(ctx context.Context, tx db.Tx, params userParams) (string, error) {
	const query = `
		INSERT INTO users (
			name,
			email,
			phone,
			password_hash,
			platform_role,
			status,
			email_verified_at
		)
		VALUES ($1, $2, $3, $4, $5, 'active', now())
		ON CONFLICT (email) DO UPDATE SET
			name = EXCLUDED.name,
			phone = EXCLUDED.phone,
			password_hash = EXCLUDED.password_hash,
			platform_role = EXCLUDED.platform_role,
			status = 'active',
			email_verified_at = COALESCE(users.email_verified_at, now()),
			updated_at = now(),
			deleted_at = NULL
		RETURNING id
	`

	var id string
	if err := tx.QueryRow(
		ctx,
		query,
		params.Name,
		params.Email,
		params.Phone,
		s.passwordHash,
		params.PlatformRole,
	).Scan(&id); err != nil {
		return "", fmt.Errorf("upsert user %s: %w", params.Email, err)
	}

	return id, nil
}

func (s demoSeed) upsertTenant(ctx context.Context, tx db.Tx, planID string) (string, error) {
	const query = `
		INSERT INTO tenants (
			plan_id,
			name,
			slug,
			status,
			trial_ends_at,
			deleted_at
		)
		VALUES ($1, 'Toko Bunga Ayu', 'toko-bunga-ayu', 'trialing', now() + interval '30 days', NULL)
		ON CONFLICT (slug) DO UPDATE SET
			plan_id = EXCLUDED.plan_id,
			name = EXCLUDED.name,
			status = 'trialing',
			trial_ends_at = COALESCE(tenants.trial_ends_at, EXCLUDED.trial_ends_at),
			deleted_at = NULL,
			updated_at = now()
		RETURNING id
	`

	var id string
	if err := tx.QueryRow(ctx, query, planID).Scan(&id); err != nil {
		return "", fmt.Errorf("upsert tenant: %w", err)
	}

	return id, nil
}

func (s demoSeed) upsertStore(ctx context.Context, tx db.Tx, tenantID string) (string, error) {
	const query = `
		INSERT INTO stores (
			tenant_id,
			name,
			slug,
			description,
			phone,
			whatsapp,
			email,
			address,
			city,
			province,
			postal_code,
			status,
			is_discoverable,
			published_at,
			deleted_at
		)
		VALUES (
			$1,
			'Toko Bunga Ayu',
			'toko-bunga-ayu',
			'Demo toko bunga non-produksi di Makassar untuk pilot UMKM Commerce OS Hybrid.',
			'081234567890',
			'081234567890',
			'halo@tokobungaayu.test',
			'Jl. Pengayoman No. 12, Panakkukang',
			'Makassar',
			'Sulawesi Selatan',
			'90231',
			'published',
			true,
			now(),
			NULL
		)
		ON CONFLICT (tenant_id, slug) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			phone = EXCLUDED.phone,
			whatsapp = EXCLUDED.whatsapp,
			email = EXCLUDED.email,
			address = EXCLUDED.address,
			city = EXCLUDED.city,
			province = EXCLUDED.province,
			postal_code = EXCLUDED.postal_code,
			status = 'published',
			is_discoverable = true,
			published_at = COALESCE(stores.published_at, now()),
			deleted_at = NULL,
			updated_at = now()
		RETURNING id
	`

	var id string
	if err := tx.QueryRow(ctx, query, tenantID).Scan(&id); err != nil {
		return "", fmt.Errorf("upsert store: %w", err)
	}

	return id, nil
}

func (s demoSeed) upsertMembership(ctx context.Context, tx db.Tx, userID string, tenantID string, role string) error {
	const query = `
		INSERT INTO user_tenants (
			user_id,
			tenant_id,
			role,
			status,
			joined_at
		)
		VALUES ($1, $2, $3, 'active', now())
		ON CONFLICT (user_id, tenant_id) DO UPDATE SET
			role = EXCLUDED.role,
			status = 'active',
			joined_at = COALESCE(user_tenants.joined_at, now()),
			updated_at = now()
	`

	if _, err := tx.Exec(ctx, query, userID, tenantID, role); err != nil {
		return fmt.Errorf("upsert membership %s: %w", role, err)
	}

	return nil
}

func (s demoSeed) upsertCategories(ctx context.Context, tx db.Tx, tenantID string, storeID string) (map[string]string, error) {
	categories := []struct {
		Name        string
		Slug        string
		Description string
		SortOrder   int
	}{
		{"Bouquet", "bouquet", "Rangkaian bunga demo untuk hadiah harian.", 1},
		{"Hampers", "hampers", "Paket hampers demo untuk wisuda dan perayaan.", 2},
	}

	result := make(map[string]string, len(categories))
	const query = `
		INSERT INTO categories (
			tenant_id,
			store_id,
			name,
			slug,
			description,
			sort_order,
			is_active,
			deleted_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, true, NULL)
		ON CONFLICT (store_id, slug) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			sort_order = EXCLUDED.sort_order,
			is_active = true,
			deleted_at = NULL,
			updated_at = now()
		RETURNING id
	`

	for _, category := range categories {
		var id string
		if err := tx.QueryRow(
			ctx,
			query,
			tenantID,
			storeID,
			category.Name,
			category.Slug,
			category.Description,
			category.SortOrder,
		).Scan(&id); err != nil {
			return nil, fmt.Errorf("upsert category %s: %w", category.Slug, err)
		}
		result[category.Slug] = id
	}

	return result, nil
}

func (s demoSeed) upsertProducts(ctx context.Context, tx db.Tx, tenantID string, storeID string, ownerID string, categories map[string]string) error {
	compareMawar := int64(175000)
	products := []demoProduct{
		{
			Name:           "Bouquet Mawar Merah",
			Slug:           "bouquet-mawar-merah",
			Description:    "Bouquet mawar merah demo untuk hadiah ulang tahun dan anniversary.",
			SKU:            "BQT-MWR-001",
			CategorySlug:   "bouquet",
			Price:          150000,
			CompareAtPrice: &compareMawar,
			CostPrice:      90000,
			InitialStock:   12,
			LowStock:       3,
			IsDiscoverable: true,
			WeightGram:     700,
		},
		{
			Name:           "Bouquet Money",
			Slug:           "bouquet-money",
			Description:    "Bouquet uang demo. Nominal uang asli tidak termasuk dalam harga demo.",
			SKU:            "BQT-MNY-001",
			CategorySlug:   "bouquet",
			Price:          250000,
			CostPrice:      150000,
			InitialStock:   8,
			LowStock:       2,
			IsDiscoverable: true,
			WeightGram:     600,
		},
		{
			Name:           "Hampers Wisuda",
			Slug:           "hampers-wisuda",
			Description:    "Hampers wisuda demo berisi bunga, kartu ucapan, dan kemasan premium.",
			SKU:            "HMP-WSD-001",
			CategorySlug:   "hampers",
			Price:          185000,
			CostPrice:      110000,
			InitialStock:   10,
			LowStock:       3,
			IsDiscoverable: true,
			WeightGram:     1200,
		},
		{
			Name:           "Bunga Matahari Mini",
			Slug:           "bunga-matahari-mini",
			Description:    "Rangkaian bunga matahari mini demo untuk hadiah cepat.",
			SKU:            "BQT-SUN-001",
			CategorySlug:   "bouquet",
			Price:          95000,
			CostPrice:      55000,
			InitialStock:   15,
			LowStock:       4,
			IsDiscoverable: false,
			WeightGram:     500,
		},
	}

	for _, product := range products {
		categoryID, ok := categories[product.CategorySlug]
		if !ok {
			return fmt.Errorf("missing category %s for product %s", product.CategorySlug, product.Slug)
		}

		productID, err := s.upsertProduct(ctx, tx, tenantID, storeID, categoryID, product)
		if err != nil {
			return err
		}

		if err := s.ensureInitialStock(ctx, tx, tenantID, storeID, productID, ownerID, product); err != nil {
			return err
		}
	}

	return nil
}

func (s demoSeed) upsertProduct(ctx context.Context, tx db.Tx, tenantID string, storeID string, categoryID string, product demoProduct) (string, error) {
	const query = `
		INSERT INTO products (
			tenant_id,
			store_id,
			category_id,
			name,
			slug,
			description,
			sku,
			price,
			compare_at_price,
			cost_price,
			weight_gram,
			status,
			is_discoverable,
			track_inventory,
			allow_backorder,
			deleted_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'active', $12, true, false, NULL)
		ON CONFLICT (store_id, slug) DO UPDATE SET
			category_id = EXCLUDED.category_id,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			sku = EXCLUDED.sku,
			price = EXCLUDED.price,
			compare_at_price = EXCLUDED.compare_at_price,
			cost_price = EXCLUDED.cost_price,
			weight_gram = EXCLUDED.weight_gram,
			status = 'active',
			is_discoverable = EXCLUDED.is_discoverable,
			track_inventory = true,
			allow_backorder = false,
			deleted_at = NULL,
			updated_at = now()
		RETURNING id
	`

	var id string
	if err := tx.QueryRow(
		ctx,
		query,
		tenantID,
		storeID,
		categoryID,
		product.Name,
		product.Slug,
		product.Description,
		product.SKU,
		product.Price,
		product.CompareAtPrice,
		product.CostPrice,
		product.WeightGram,
		product.IsDiscoverable,
	).Scan(&id); err != nil {
		return "", fmt.Errorf("upsert product %s: %w", product.Slug, err)
	}

	return id, nil
}

func (s demoSeed) ensureInitialStock(ctx context.Context, tx db.Tx, tenantID string, storeID string, productID string, ownerID string, product demoProduct) error {
	const snapshotQuery = `
		INSERT INTO product_stock_snapshots (
			tenant_id,
			store_id,
			product_id,
			quantity_on_hand,
			quantity_reserved,
			quantity_available,
			low_stock_threshold
		)
		VALUES ($1, $2, $3, $4, 0, $4, $5)
		ON CONFLICT (product_id) DO NOTHING
	`
	if _, err := tx.Exec(ctx, snapshotQuery, tenantID, storeID, productID, product.InitialStock, product.LowStock); err != nil {
		return fmt.Errorf("ensure stock snapshot %s: %w", product.Slug, err)
	}

	const movementQuery = `
		INSERT INTO stock_movements (
			tenant_id,
			store_id,
			product_id,
			movement_type,
			quantity,
			balance_after,
			reference_type,
			reference_id,
			note,
			created_by
		)
		SELECT $1, $2, $3, 'initial', $4, $4, 'demo_seed', $3, 'Initial stock from non-production demo seed.', $5
		WHERE NOT EXISTS (
			SELECT 1
			FROM stock_movements
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND product_id = $3
			  AND movement_type = 'initial'
			  AND reference_type = 'demo_seed'
			  AND reference_id = $3
		)
	`
	if _, err := tx.Exec(ctx, movementQuery, tenantID, storeID, productID, product.InitialStock, ownerID); err != nil {
		return fmt.Errorf("ensure initial stock movement %s: %w", product.Slug, err)
	}

	return nil
}

func (s demoSeed) upsertCourierZones(ctx context.Context, tx db.Tx, tenantID string, storeID string) error {
	zones := []demoZone{
		{
			Name:        "Makassar Kota",
			Description: "Pengiriman demo area Makassar kota.",
			Rate:        15000,
			SortOrder:   1,
		},
		{
			Name:        "Sekitar Makassar",
			Description: "Pengiriman demo area sekitar Makassar.",
			Rate:        25000,
			SortOrder:   2,
		},
	}

	const query = `
		WITH existing AS (
			SELECT id
			FROM courier_zones
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND name = $3
			  AND deleted_at IS NULL
			LIMIT 1
		),
		updated AS (
			UPDATE courier_zones
			SET description = $4,
			    rate = $5,
			    is_active = true,
			    sort_order = $6,
			    updated_at = now()
			WHERE id IN (SELECT id FROM existing)
			RETURNING id
		),
		inserted AS (
			INSERT INTO courier_zones (
				tenant_id,
				store_id,
				name,
				description,
				rate,
				is_active,
				sort_order
			)
			SELECT $1, $2, $3, $4, $5, true, $6
			WHERE NOT EXISTS (SELECT 1 FROM existing)
			RETURNING id
		)
		SELECT id FROM updated
		UNION ALL
		SELECT id FROM inserted
	`

	for _, zone := range zones {
		var id string
		if err := tx.QueryRow(ctx, query, tenantID, storeID, zone.Name, zone.Description, zone.Rate, zone.SortOrder).Scan(&id); err != nil {
			return fmt.Errorf("upsert courier zone %s: %w", zone.Name, err)
		}
	}

	return nil
}

func envString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
