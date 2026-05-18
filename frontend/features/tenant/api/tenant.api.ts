import { apiFetch } from "@/lib/api/client";
import { permissionsForRole } from "@/lib/permissions/roles";
import type { TenantMembership, TenantRole } from "@/lib/stores/tenant.store";

type ApiTenant = {
  id: string;
  name: string;
  slug: string;
  role: TenantRole;
  status: string;
  store: {
    id: string;
    name: string;
    slug: string;
    status: string;
  };
};

export async function listTenants(): Promise<TenantMembership[]> {
  const tenants = await apiFetch<ApiTenant[]>("/api/v1/tenants", {
    tenantScoped: false
  });

  return tenants.map((tenant) => ({
    ...tenant,
    permissions: permissionsForRole(tenant.role)
  }));
}
