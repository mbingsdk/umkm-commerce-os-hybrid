package testutil

import (
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
)

type SecurityFixtures struct {
	Users     UserFixtures
	Tenants   TenantFixtures
	Stores    StoreFixtures
	Catalog   CatalogFixtures
	Orders    OrderFixtures
	Inventory InventoryFixtures
}

type UserFixtures struct {
	Owner      UserFixture
	Manager    UserFixture
	Staff      UserFixture
	Cashier    UserFixture
	SuperAdmin UserFixture
}

type UserFixture struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PlatformRole string
	TenantRole   string
}

type TenantFixtures struct {
	A TenantFixture
	B TenantFixture
}

type TenantFixture struct {
	ID     uuid.UUID
	Name   string
	Slug   string
	Status string
}

type StoreFixtures struct {
	A StoreFixture
	B StoreFixture
}

type StoreFixture struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	Name     string
	Slug     string
	Status   string
}

type CatalogFixtures struct {
	CategoryA CategoryFixture
	CategoryB CategoryFixture
	ProductA  ProductFixture
	ProductB  ProductFixture
}

type CategoryFixture struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	StoreID  uuid.UUID
	Name     string
	Slug     string
}

type ProductFixture struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	StoreID  uuid.UUID
	Name     string
	Slug     string
	SKU      string
	Status   string
	Price    int64
}

type OrderFixtures struct {
	A OrderFixture
	B OrderFixture
}

type OrderFixture struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	StoreID     uuid.UUID
	OrderNumber string
	Status      string
	GrandTotal  int64
}

type InventoryFixtures struct {
	StockA StockFixture
	StockB StockFixture
}

type StockFixture struct {
	ProductID         uuid.UUID
	TenantID          uuid.UUID
	StoreID           uuid.UUID
	QuantityOnHand    int
	QuantityReserved  int
	QuantityAvailable int
	LowStockThreshold int
}

func NewSecurityFixtures() SecurityFixtures {
	tenantA := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	tenantB := uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
	storeA := uuid.MustParse("aaaaaaaa-1111-4111-8111-aaaaaaaaaaaa")
	storeB := uuid.MustParse("bbbbbbbb-2222-4222-8222-bbbbbbbbbbbb")
	productA := uuid.MustParse("aaaaaaaa-3333-4333-8333-aaaaaaaaaaaa")
	productB := uuid.MustParse("bbbbbbbb-4444-4444-8444-bbbbbbbbbbbb")

	return SecurityFixtures{
		Users: UserFixtures{
			Owner: UserFixture{
				ID:           uuid.MustParse("11111111-1111-4111-8111-111111111111"),
				Name:         "Owner Tenant A",
				Email:        "owner-a@example.test",
				PlatformRole: auth.PlatformRoleUser,
				TenantRole:   string(permission.RoleOwner),
			},
			Manager: UserFixture{
				ID:           uuid.MustParse("22222222-2222-4222-8222-222222222222"),
				Name:         "Manager Tenant A",
				Email:        "manager-a@example.test",
				PlatformRole: auth.PlatformRoleUser,
				TenantRole:   string(permission.RoleManager),
			},
			Staff: UserFixture{
				ID:           uuid.MustParse("33333333-3333-4333-8333-333333333333"),
				Name:         "Staff Tenant A",
				Email:        "staff-a@example.test",
				PlatformRole: auth.PlatformRoleUser,
				TenantRole:   string(permission.RoleStaff),
			},
			Cashier: UserFixture{
				ID:           uuid.MustParse("44444444-4444-4444-8444-444444444444"),
				Name:         "Cashier Tenant A",
				Email:        "cashier-a@example.test",
				PlatformRole: auth.PlatformRoleUser,
				TenantRole:   string(permission.RoleCashier),
			},
			SuperAdmin: UserFixture{
				ID:           uuid.MustParse("55555555-5555-4555-8555-555555555555"),
				Name:         "Platform Super Admin",
				Email:        "super-admin@example.test",
				PlatformRole: auth.PlatformRoleSuperAdmin,
				TenantRole:   "",
			},
		},
		Tenants: TenantFixtures{
			A: TenantFixture{ID: tenantA, Name: "Tenant A", Slug: "tenant-a", Status: "active"},
			B: TenantFixture{ID: tenantB, Name: "Tenant B", Slug: "tenant-b", Status: "active"},
		},
		Stores: StoreFixtures{
			A: StoreFixture{ID: storeA, TenantID: tenantA, Name: "Toko A", Slug: "toko-a", Status: "published"},
			B: StoreFixture{ID: storeB, TenantID: tenantB, Name: "Toko B", Slug: "toko-b", Status: "published"},
		},
		Catalog: CatalogFixtures{
			CategoryA: CategoryFixture{
				ID:       uuid.MustParse("aaaaaaaa-5555-4555-8555-aaaaaaaaaaaa"),
				TenantID: tenantA,
				StoreID:  storeA,
				Name:     "Bouquet",
				Slug:     "bouquet",
			},
			CategoryB: CategoryFixture{
				ID:       uuid.MustParse("bbbbbbbb-6666-4666-8666-bbbbbbbbbbbb"),
				TenantID: tenantB,
				StoreID:  storeB,
				Name:     "Snack",
				Slug:     "snack",
			},
			ProductA: ProductFixture{
				ID:       productA,
				TenantID: tenantA,
				StoreID:  storeA,
				Name:     "Bouquet Mawar",
				Slug:     "bouquet-mawar",
				SKU:      "BQT-A-001",
				Status:   "active",
				Price:    50000,
			},
			ProductB: ProductFixture{
				ID:       productB,
				TenantID: tenantB,
				StoreID:  storeB,
				Name:     "Snack Box",
				Slug:     "snack-box",
				SKU:      "SNK-B-001",
				Status:   "active",
				Price:    35000,
			},
		},
		Orders: OrderFixtures{
			A: OrderFixture{
				ID:          uuid.MustParse("aaaaaaaa-7777-4777-8777-aaaaaaaaaaaa"),
				TenantID:    tenantA,
				StoreID:     storeA,
				OrderNumber: "ORD-20260524-A",
				Status:      "pending_payment",
				GrandTotal:  50000,
			},
			B: OrderFixture{
				ID:          uuid.MustParse("bbbbbbbb-8888-4888-8888-bbbbbbbbbbbb"),
				TenantID:    tenantB,
				StoreID:     storeB,
				OrderNumber: "ORD-20260524-B",
				Status:      "pending_payment",
				GrandTotal:  35000,
			},
		},
		Inventory: InventoryFixtures{
			StockA: StockFixture{
				ProductID:         productA,
				TenantID:          tenantA,
				StoreID:           storeA,
				QuantityOnHand:    10,
				QuantityReserved:  2,
				QuantityAvailable: 8,
				LowStockThreshold: 3,
			},
			StockB: StockFixture{
				ProductID:         productB,
				TenantID:          tenantB,
				StoreID:           storeB,
				QuantityOnHand:    5,
				QuantityReserved:  0,
				QuantityAvailable: 5,
				LowStockThreshold: 2,
			},
		},
	}
}

func FixtureTime() time.Time {
	return time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
}
