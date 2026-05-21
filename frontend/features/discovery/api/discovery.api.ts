import { cache } from "react";
import { ApiError, isApiErrorResponse, toApiError } from "@/lib/api/errors";
import type {
  DiscoveryAggregate,
  DiscoveryHome,
  DiscoveryListResult,
  DiscoveryProduct,
  DiscoveryProductFilters,
  DiscoverySearchFilters,
  DiscoverySearchResult,
  DiscoveryStore,
  DiscoveryStoreFilters,
  Pagination
} from "@/features/discovery/types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

type ApiSuccessResponse<T> = {
  success: true;
  message: string;
  data?: T;
  meta?: {
    pagination?: {
      limit: number;
      next_cursor?: string | null;
      has_more: boolean;
    };
  };
};

type ApiStore = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  logo_url?: string;
  banner_url?: string;
  city?: string;
  province?: string;
  store_url: string;
};

type ApiProduct = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  price: number;
  primary_image_url?: string;
  category?: {
    name: string;
    slug: string;
  };
  store: {
    id: string;
    name: string;
    slug: string;
    city?: string;
    province?: string;
  };
  store_url: string;
  product_url: string;
};

type ApiAggregate = {
  name?: string;
  slug?: string;
  city?: string;
  count: number;
};

type ApiHome = {
  featured_stores?: ApiStore[];
  featured_products?: ApiProduct[];
  latest_stores?: ApiStore[];
  latest_products?: ApiProduct[];
  popular_categories?: ApiAggregate[];
  popular_cities?: ApiAggregate[];
};

type ApiSearch = {
  stores?: ApiStore[];
  products?: ApiProduct[];
};

export const getDiscoveryHome = cache(async (): Promise<DiscoveryHome> => {
  const result = await publicDiscoveryFetch<ApiHome>("/api/v1/public/discovery/home");

  return {
    featuredStores: (result.data.featured_stores ?? []).map(normalizeStore),
    featuredProducts: (result.data.featured_products ?? []).map(normalizeProduct),
    latestStores: (result.data.latest_stores ?? []).map(normalizeStore),
    latestProducts: (result.data.latest_products ?? []).map(normalizeProduct),
    popularCategories: (result.data.popular_categories ?? []).map(normalizeAggregate),
    popularCities: (result.data.popular_cities ?? []).map(normalizeAggregate)
  };
});

export async function listDiscoveryStores(
  filters: DiscoveryStoreFilters = {}
): Promise<DiscoveryListResult<DiscoveryStore>> {
  const result = await publicDiscoveryFetch<ApiStore[]>(`/api/v1/public/discovery/stores${toQueryString(filters)}`);

  return {
    items: result.data.map(normalizeStore),
    pagination: normalizePagination(result.meta)
  };
}

export async function listDiscoveryProducts(
  filters: DiscoveryProductFilters = {}
): Promise<DiscoveryListResult<DiscoveryProduct>> {
  const result = await publicDiscoveryFetch<ApiProduct[]>(`/api/v1/public/discovery/products${toQueryString(filters)}`);

  return {
    items: result.data.map(normalizeProduct),
    pagination: normalizePagination(result.meta)
  };
}

export async function searchDiscovery(filters: DiscoverySearchFilters = {}): Promise<DiscoverySearchResult> {
  const result = await publicDiscoveryFetch<ApiSearch>(`/api/v1/public/discovery/search${toQueryString(filters)}`);

  return {
    stores: (result.data.stores ?? []).map(normalizeStore),
    products: (result.data.products ?? []).map(normalizeProduct)
  };
}

async function publicDiscoveryFetch<T>(path: string): Promise<{ data: T; meta?: ApiSuccessResponse<T>["meta"] }> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    cache: "no-store"
  });
  const body = await parseJSON(response);

  if (!response.ok || isApiErrorResponse(body)) {
    if (isApiErrorResponse(body)) {
      throw toApiError(body, response.status);
    }

    throw new ApiError("Discovery publik gagal dimuat.", response.status, "UNKNOWN_API_ERROR");
  }

  if (!isSuccessResponse<T>(body)) {
    throw new ApiError("Respons discovery tidak valid.", response.status, "INVALID_API_RESPONSE");
  }

  return {
    data: body.data as T,
    meta: body.meta
  };
}

async function parseJSON(response: Response) {
  const contentType = response.headers.get("content-type");
  if (!contentType?.includes("application/json")) {
    return null;
  }

  return response.json() as Promise<unknown>;
}

function isSuccessResponse<T>(value: unknown): value is ApiSuccessResponse<T> {
  if (!value || typeof value !== "object") {
    return false;
  }

  const candidate = value as Partial<ApiSuccessResponse<T>>;
  return candidate.success === true && typeof candidate.message === "string";
}

function toQueryString(filters: Record<string, unknown>) {
  const params = new URLSearchParams();

  Object.entries(filters).forEach(([key, value]) => {
    if (value == null || value === "") {
      return;
    }

    const apiKey =
      key === "priceMin"
        ? "price_min"
        : key === "priceMax"
          ? "price_max"
          : key === "query"
            ? "q"
            : key;

    params.set(apiKey, String(value));
  });

  return params.size > 0 ? `?${params.toString()}` : "";
}

function normalizeStore(store: ApiStore): DiscoveryStore {
  return {
    id: store.id,
    name: store.name,
    slug: store.slug,
    description: store.description,
    logoUrl: store.logo_url,
    bannerUrl: store.banner_url,
    city: store.city,
    province: store.province,
    storeUrl: store.store_url || `/s/${store.slug}`
  };
}

function normalizeProduct(product: ApiProduct): DiscoveryProduct {
  return {
    id: product.id,
    name: product.name,
    slug: product.slug,
    description: product.description,
    price: product.price,
    primaryImageUrl: product.primary_image_url,
    category: product.category,
    store: product.store,
    storeUrl: product.store_url || `/s/${product.store.slug}`,
    productUrl: product.product_url || `/s/${product.store.slug}/products/${product.slug}`
  };
}

function normalizeAggregate(item: ApiAggregate): DiscoveryAggregate {
  return {
    name: item.name,
    slug: item.slug,
    city: item.city,
    count: item.count
  };
}

function normalizePagination(meta?: ApiSuccessResponse<unknown>["meta"]): Pagination {
  return {
    limit: meta?.pagination?.limit ?? 20,
    nextCursor: meta?.pagination?.next_cursor ?? null,
    hasMore: meta?.pagination?.has_more ?? false
  };
}
