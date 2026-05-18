import { apiFetch } from "@/lib/api/client";

type ApiStore = {
  id: string;
  tenant_id: string;
  name: string;
  slug: string;
  description?: string;
  logo_url?: string;
  banner_url?: string;
  phone?: string;
  whatsapp?: string;
  email?: string;
  address?: string;
  city?: string;
  province?: string;
  postal_code?: string;
  status: string;
  is_discoverable: boolean;
  published_at?: string;
};

type ApiBusinessHoursResponse = {
  items: Array<{
    day_of_week: number;
    open_time?: string;
    close_time?: string;
    is_closed: boolean;
  }>;
};

export type Store = {
  id: string;
  tenantId: string;
  name: string;
  slug: string;
  description?: string;
  logoUrl?: string;
  bannerUrl?: string;
  phone?: string;
  whatsapp?: string;
  email?: string;
  address?: string;
  city?: string;
  province?: string;
  postalCode?: string;
  status: string;
  isDiscoverable: boolean;
  publishedAt?: string;
};

export type UpdateCurrentStoreInput = {
  name: string;
  description?: string;
  phone?: string;
  whatsapp?: string;
  email?: string;
  address?: string;
  city?: string;
  province?: string;
  postal_code?: string;
  is_discoverable: boolean;
};

export type UpdateBusinessHoursInput = {
  items: Array<{
    day_of_week: number;
    open_time?: string;
    close_time?: string;
    is_closed: boolean;
  }>;
};

export async function getCurrentStore(): Promise<Store> {
  const store = await apiFetch<ApiStore>("/api/v1/stores/current");
  return normalizeStore(store);
}

export async function updateCurrentStore(input: UpdateCurrentStoreInput): Promise<Store> {
  const store = await apiFetch<ApiStore>("/api/v1/stores/current", {
    method: "PATCH",
    body: JSON.stringify(input)
  });

  return normalizeStore(store);
}

export async function publishCurrentStore(): Promise<Store> {
  const store = await apiFetch<ApiStore>("/api/v1/stores/current/publish", {
    method: "POST"
  });

  return normalizeStore(store);
}

export async function unpublishCurrentStore(): Promise<Store> {
  const store = await apiFetch<ApiStore>("/api/v1/stores/current/unpublish", {
    method: "POST"
  });

  return normalizeStore(store);
}

export async function updateBusinessHours(input: UpdateBusinessHoursInput) {
  return apiFetch<ApiBusinessHoursResponse>("/api/v1/stores/current/business-hours", {
    method: "PUT",
    body: JSON.stringify(input)
  });
}

function normalizeStore(store: ApiStore): Store {
  return {
    id: store.id,
    tenantId: store.tenant_id,
    name: store.name,
    slug: store.slug,
    description: store.description,
    logoUrl: store.logo_url,
    bannerUrl: store.banner_url,
    phone: store.phone,
    whatsapp: store.whatsapp,
    email: store.email,
    address: store.address,
    city: store.city,
    province: store.province,
    postalCode: store.postal_code,
    status: store.status,
    isDiscoverable: store.is_discoverable,
    publishedAt: store.published_at
  };
}
