import { apiFetch } from "@/lib/api/client";
import { permissionsForRole } from "@/lib/permissions/roles";
import type { TenantMembership, TenantRole } from "@/lib/stores/tenant.store";

export type CreateStoreInput = {
  tenant_name: string;
  tenant_slug: string;
  store: {
    name: string;
    slug: string;
    description?: string;
    whatsapp?: string;
    city?: string;
    province?: string;
  };
};

type CreateStoreResponse = {
  tenant: {
    id: string;
    name: string;
    slug: string;
    status: string;
  };
  store: {
    id: string;
    name: string;
    slug: string;
    status: string;
  };
  membership: {
    role: TenantRole;
    permissions: string[];
  };
};

export async function createStoreOnboarding(input: CreateStoreInput): Promise<TenantMembership> {
  const response = await apiFetch<CreateStoreResponse>("/api/v1/onboarding/create-store", {
    method: "POST",
    body: JSON.stringify(input),
    tenantScoped: false
  });

  return {
    id: response.tenant.id,
    name: response.tenant.name,
    slug: response.tenant.slug,
    role: response.membership.role,
    status: response.tenant.status,
    store: response.store,
    permissions:
      response.membership.permissions.length > 0
        ? response.membership.permissions
        : permissionsForRole(response.membership.role)
  };
}
