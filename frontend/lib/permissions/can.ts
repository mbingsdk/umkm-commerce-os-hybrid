import type { Permission } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";

export function can(permission: Permission) {
  return useTenantStore.getState().permissions.includes(permission);
}
