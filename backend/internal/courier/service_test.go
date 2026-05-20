package courier

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/audit"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

var (
	courierTenantA = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	courierStoreA  = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	courierTenantB = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	courierStoreB  = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	courierZoneA   = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	courierZoneB   = uuid.MustParse("66666666-6666-6666-6666-666666666666")
	courierZoneC   = uuid.MustParse("77777777-7777-7777-7777-777777777777")
	courierActorID = uuid.MustParse("88888888-8888-8888-8888-888888888888")
)

func TestCreateZoneValidation(t *testing.T) {
	service, repo, auditRepo := newCourierTestService()

	_, err := service.CreateZone(context.Background(), courierTenantA, courierStoreA, CreateZoneInput{
		ActorUserID: courierActorID,
		Name:        "",
		Rate:        -1,
	})
	assertCourierAppErrorCode(t, err, apperror.CodeValidation)
	if len(repo.zones) != 3 {
		t.Fatalf("zone count = %d, want unchanged 3", len(repo.zones))
	}
	if len(auditRepo.entries) != 0 {
		t.Fatalf("audit entries = %d, want 0", len(auditRepo.entries))
	}
}

func TestListZonesTenantScoped(t *testing.T) {
	service, _, _ := newCourierTestService()

	zones, err := service.ListZones(context.Background(), courierTenantA, courierStoreA, ListZoneFilters{})
	if err != nil {
		t.Fatalf("ListZones error = %v", err)
	}
	if len(zones) != 2 {
		t.Fatalf("zones len = %d, want 2 tenant A zones", len(zones))
	}
	for _, zone := range zones {
		if zone.ID == courierZoneB.String() {
			t.Fatalf("tenant B zone leaked into tenant A list: %#v", zones)
		}
	}
}

func TestTenantCannotUpdateOtherTenantZone(t *testing.T) {
	service, _, auditRepo := newCourierTestService()
	name := "Zona dari tenant A"

	_, err := service.UpdateZone(context.Background(), courierTenantA, courierStoreA, courierZoneB, UpdateZoneInput{
		ActorUserID: courierActorID,
		Name:        &name,
	})
	assertCourierAppErrorCode(t, err, apperror.CodeNotFound)
	if len(auditRepo.entries) != 0 {
		t.Fatalf("audit entries = %d, want 0", len(auditRepo.entries))
	}
}

func TestPublicEndpointHidesInactiveZones(t *testing.T) {
	service, _, _ := newCourierTestService()

	zones, err := service.ListPublicZones(context.Background(), "toko-a")
	if err != nil {
		t.Fatalf("ListPublicZones error = %v", err)
	}
	if len(zones) != 1 {
		t.Fatalf("public zones len = %d, want only 1 active zone", len(zones))
	}
	if zones[0].ID != courierZoneA.String() {
		t.Fatalf("public zone = %s, want active zone %s", zones[0].ID, courierZoneA)
	}
}

func TestPublicEndpointRejectsUnpublishedStore(t *testing.T) {
	service, _, _ := newCourierTestService()

	_, err := service.ListPublicZones(context.Background(), "draft-store")
	assertCourierAppErrorCode(t, err, apperror.CodeNotFound)
}

func TestDeleteIsTenantScoped(t *testing.T) {
	service, repo, auditRepo := newCourierTestService()

	_, err := service.DeleteZone(context.Background(), courierTenantA, courierStoreA, courierZoneB, courierActorID)
	assertCourierAppErrorCode(t, err, apperror.CodeNotFound)
	if repo.zones[courierZoneB].DeletedAt != nil {
		t.Fatalf("tenant B zone was deleted by tenant A")
	}
	if len(auditRepo.entries) != 0 {
		t.Fatalf("audit entries = %d, want 0", len(auditRepo.entries))
	}
}

func TestCreateUpdateDeleteZoneCreatesAuditLog(t *testing.T) {
	service, repo, auditRepo := newCourierTestService()
	falseValue := false

	created, err := service.CreateZone(context.Background(), courierTenantA, courierStoreA, CreateZoneInput{
		ActorUserID: courierActorID,
		Name:        "Gratis radius dekat",
		Rate:        0,
		IsActive:    &falseValue,
	})
	if err != nil {
		t.Fatalf("CreateZone error = %v", err)
	}
	if created.ID == "" || created.IsActive {
		t.Fatalf("created zone = %#v", created)
	}

	createdID := uuid.MustParse(created.ID)
	newRate := int64(15000)
	_, err = service.UpdateZone(context.Background(), courierTenantA, courierStoreA, createdID, UpdateZoneInput{
		ActorUserID: courierActorID,
		Rate:        &newRate,
	})
	if err != nil {
		t.Fatalf("UpdateZone error = %v", err)
	}

	_, err = service.DeleteZone(context.Background(), courierTenantA, courierStoreA, createdID, courierActorID)
	if err != nil {
		t.Fatalf("DeleteZone error = %v", err)
	}
	if repo.zones[createdID].DeletedAt == nil {
		t.Fatalf("created zone was not soft deleted")
	}
	if len(auditRepo.entries) != 3 {
		t.Fatalf("audit entries = %d, want 3", len(auditRepo.entries))
	}
	if auditRepo.entries[0].Action != AuditActionCourierZoneCreated ||
		auditRepo.entries[1].Action != AuditActionCourierZoneUpdated ||
		auditRepo.entries[2].Action != AuditActionCourierZoneDeleted {
		t.Fatalf("audit entries = %#v", auditRepo.entries)
	}
}

func TestCourierPermissionDefaultDenyRemainsIntact(t *testing.T) {
	tests := []struct {
		name       string
		role       permission.Role
		permission permission.Permission
		want       bool
	}{
		{name: "owner can create zone", role: permission.RoleOwner, permission: permission.CourierCreateZone, want: true},
		{name: "manager can delete zone", role: permission.RoleManager, permission: permission.CourierDeleteZone, want: true},
		{name: "staff can only read zone", role: permission.RoleStaff, permission: permission.CourierReadZone, want: true},
		{name: "staff cannot create zone", role: permission.RoleStaff, permission: permission.CourierCreateZone, want: false},
		{name: "courier admin cannot delete zone", role: permission.RoleCourierAdmin, permission: permission.CourierDeleteZone, want: false},
		{name: "cashier denied zone read", role: permission.RoleCashier, permission: permission.CourierReadZone, want: false},
		{name: "driver denied zone read", role: permission.RoleDriver, permission: permission.CourierReadZone, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := permission.Allowed(string(tt.role), tt.permission); got != tt.want {
				t.Fatalf("Allowed(%q, %q) = %v, want %v", tt.role, tt.permission, got, tt.want)
			}
		})
	}
}

func newCourierTestService() (*Service, *fakeCourierRepository, *fakeCourierAuditRepository) {
	now := time.Date(2026, 5, 21, 8, 0, 0, 0, time.UTC)
	repo := &fakeCourierRepository{
		now: now,
		zones: map[uuid.UUID]Zone{
			courierZoneA: {
				ID:          courierZoneA,
				TenantID:    courierTenantA,
				StoreID:     courierStoreA,
				Name:        "Dalam kota",
				Description: "Area kota",
				Rate:        10000,
				IsActive:    true,
				SortOrder:   1,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			courierZoneC: {
				ID:        courierZoneC,
				TenantID:  courierTenantA,
				StoreID:   courierStoreA,
				Name:      "Luar kota",
				Rate:      25000,
				IsActive:  false,
				SortOrder: 2,
				CreatedAt: now,
				UpdatedAt: now,
			},
			courierZoneB: {
				ID:        courierZoneB,
				TenantID:  courierTenantB,
				StoreID:   courierStoreB,
				Name:      "Tenant B",
				Rate:      30000,
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}
	publicStores := &fakeCourierPublicStores{
		stores: map[string]store.PublicContext{
			"toko-a": {
				TenantID: courierTenantA,
				StoreID:  courierStoreA,
				Store: store.Store{
					ID:       courierStoreA,
					TenantID: courierTenantA,
					Slug:     "toko-a",
					Status:   store.StatusPublished,
				},
			},
		},
	}
	auditRepo := &fakeCourierAuditRepository{}
	return NewService(fakeCourierDB{}, repo, publicStores, auditRepo), repo, auditRepo
}

func assertCourierAppErrorCode(t *testing.T, err error, code apperror.Code) {
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

type fakeCourierDB struct{}

func (fakeCourierDB) WithTx(_ context.Context, fn func(tx db.Tx) error) error {
	return fn(fakeCourierDB{})
}

func (fakeCourierDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (fakeCourierDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (fakeCourierDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

type fakeCourierRepository struct {
	zones map[uuid.UUID]Zone
	now   time.Time
}

func (f *fakeCourierRepository) ListZones(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, filters ListZoneFilters) ([]Zone, error) {
	zones := make([]Zone, 0)
	for _, zone := range f.zones {
		if zone.TenantID != tenantID || zone.StoreID != storeID || zone.DeletedAt != nil {
			continue
		}
		if filters.IsActive != nil && zone.IsActive != *filters.IsActive {
			continue
		}
		zones = append(zones, zone)
	}
	return zones, nil
}

func (f *fakeCourierRepository) ListPublicActiveZones(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID) ([]Zone, error) {
	zones := make([]Zone, 0)
	for _, zone := range f.zones {
		if zone.TenantID == tenantID && zone.StoreID == storeID && zone.IsActive && zone.DeletedAt == nil {
			zones = append(zones, zone)
		}
	}
	return zones, nil
}

func (f *fakeCourierRepository) FindZoneByID(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, zoneID uuid.UUID) (*Zone, error) {
	zone, ok := f.zones[zoneID]
	if !ok || zone.TenantID != tenantID || zone.StoreID != storeID || zone.DeletedAt != nil {
		return nil, ErrZoneNotFound
	}
	return cloneCourierZone(zone), nil
}

func (f *fakeCourierRepository) CreateZone(_ context.Context, _ db.Queryer, params CreateZoneParams) (*Zone, error) {
	zone := Zone{
		ID:          uuid.New(),
		TenantID:    params.TenantID,
		StoreID:     params.StoreID,
		Name:        params.Name,
		Description: params.Description,
		Rate:        params.Rate,
		IsActive:    params.IsActive,
		SortOrder:   params.SortOrder,
		CreatedAt:   f.now,
		UpdatedAt:   f.now,
	}
	f.zones[zone.ID] = zone
	return cloneCourierZone(zone), nil
}

func (f *fakeCourierRepository) UpdateZone(_ context.Context, _ db.Queryer, params UpdateZoneParams) (*Zone, error) {
	zone, ok := f.zones[params.ZoneID]
	if !ok || zone.TenantID != params.TenantID || zone.StoreID != params.StoreID || zone.DeletedAt != nil {
		return nil, ErrZoneNotFound
	}
	zone.Name = params.Name
	zone.Description = params.Description
	zone.Rate = params.Rate
	zone.IsActive = params.IsActive
	zone.SortOrder = params.SortOrder
	zone.UpdatedAt = f.now.Add(time.Minute)
	f.zones[params.ZoneID] = zone
	return cloneCourierZone(zone), nil
}

func (f *fakeCourierRepository) SoftDeleteZone(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, zoneID uuid.UUID) (*Zone, error) {
	zone, ok := f.zones[zoneID]
	if !ok || zone.TenantID != tenantID || zone.StoreID != storeID || zone.DeletedAt != nil {
		return nil, ErrZoneNotFound
	}
	deletedAt := f.now.Add(time.Minute)
	zone.DeletedAt = &deletedAt
	zone.UpdatedAt = deletedAt
	f.zones[zoneID] = zone
	return cloneCourierZone(zone), nil
}

func cloneCourierZone(zone Zone) *Zone {
	clone := zone
	return &clone
}

type fakeCourierPublicStores struct {
	stores map[string]store.PublicContext
}

func (f *fakeCourierPublicStores) Resolve(_ context.Context, slug string) (store.PublicContext, error) {
	currentStore, ok := f.stores[slug]
	if !ok {
		return store.PublicContext{}, apperror.NotFound("Store not found")
	}
	return currentStore, nil
}

type fakeCourierAuditRepository struct {
	entries []audit.Entry
}

func (f *fakeCourierAuditRepository) Create(_ context.Context, _ db.Queryer, entry audit.Entry) error {
	f.entries = append(f.entries, entry)
	return nil
}
