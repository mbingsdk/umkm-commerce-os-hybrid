import type { Permission } from "@/lib/permissions/permissions";
import { permissions } from "@/lib/permissions/permissions";
import type { TenantRole } from "@/lib/stores/tenant.store";

const rolePermissions: Record<TenantRole, Permission[]> = {
  owner: [
    permissions.tenantRead,
    permissions.tenantCreate,
    permissions.storeRead,
    permissions.storeUpdate,
    permissions.storePublish,
    permissions.storeUpdateBusinessHours,
    permissions.productRead,
    permissions.inventoryRead,
    permissions.orderRead,
    permissions.posCreateTransaction,
    permissions.financeReadSummary
  ],
  manager: [
    permissions.tenantRead,
    permissions.storeRead,
    permissions.storeUpdateBusinessHours,
    permissions.productRead,
    permissions.inventoryRead,
    permissions.orderRead,
    permissions.posCreateTransaction,
    permissions.financeReadSummary
  ],
  staff: [permissions.tenantRead, permissions.storeRead, permissions.productRead, permissions.inventoryRead, permissions.orderRead],
  cashier: [permissions.tenantRead, permissions.storeRead, permissions.posCreateTransaction],
  inventory_staff: [permissions.tenantRead, permissions.storeRead, permissions.inventoryRead],
  courier_admin: [permissions.tenantRead, permissions.storeRead, permissions.orderRead],
  driver: []
};

export function permissionsForRole(role: TenantRole) {
  return rolePermissions[role] ?? [];
}
