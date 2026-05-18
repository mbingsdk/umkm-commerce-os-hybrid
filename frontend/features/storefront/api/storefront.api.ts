import { cache } from "react";
import { ApiError, isApiErrorResponse, toApiError } from "@/lib/api/errors";
import type {
  PublicBusinessHour,
  PublicCategory,
  PublicProductDetail,
  PublicProductFilters,
  PublicProductListItem,
  PublicProductListResult,
  PublicStore,
  StockStatus
} from "@/features/storefront/types";

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

type ApiBusinessHour = {
  day_of_week: number;
  open_time?: string;
  close_time?: string;
  is_closed: boolean;
};

type ApiPublicStore = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  logo_url?: string;
  banner_url?: string;
  phone?: string;
  whatsapp?: string;
  city?: string;
  province?: string;
  business_hours?: ApiBusinessHour[];
  seo?: {
    title?: string;
    description?: string;
  };
};

type ApiPublicCategory = {
  id: string;
  name: string;
  slug: string;
  image_url?: string;
};

type ApiPublicProductListItem = {
  id: string;
  name: string;
  slug: string;
  price: number;
  compare_at_price?: number;
  primary_image_url?: string;
  stock_status: StockStatus;
};

type ApiPublicProductDetail = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  price: number;
  compare_at_price?: number;
  weight_gram: number;
  images?: Array<{
    url: string;
    alt_text?: string;
  }>;
  category?: {
    id: string;
    name: string;
    slug: string;
  };
  stock: {
    stock_status: StockStatus;
    quantity_available: number;
  };
  store: {
    name: string;
    slug: string;
    city?: string;
  };
  seo?: {
    title?: string;
    description?: string;
  };
};

export const getPublicStoreBySlug = cache(async (storeSlug: string): Promise<PublicStore> => {
  const response = await publicApiFetch<ApiPublicStore>(`/api/v1/public/stores/${encodeURIComponent(storeSlug)}`);
  return normalizeStore(response.data);
});

export const listPublicCategories = cache(async (storeSlug: string): Promise<PublicCategory[]> => {
  const response = await publicApiFetch<ApiPublicCategory[]>(
    `/api/v1/public/stores/${encodeURIComponent(storeSlug)}/categories`
  );

  return response.data.map(normalizeCategory);
});

export async function listPublicProducts(
  storeSlug: string,
  filters: PublicProductFilters = {}
): Promise<PublicProductListResult> {
  const searchParams = new URLSearchParams();

  if (filters.query) {
    searchParams.set("q", filters.query);
  }
  if (filters.categorySlug) {
    searchParams.set("category", filters.categorySlug);
  }
  if (filters.inStock !== undefined) {
    searchParams.set("in_stock", String(filters.inStock));
  }
  if (filters.limit) {
    searchParams.set("limit", String(filters.limit));
  }
  if (filters.cursor) {
    searchParams.set("cursor", filters.cursor);
  }

  const suffix = searchParams.size > 0 ? `?${searchParams.toString()}` : "";
  const response = await publicApiFetch<ApiPublicProductListItem[]>(
    `/api/v1/public/stores/${encodeURIComponent(storeSlug)}/products${suffix}`
  );

  return {
    items: response.data.map(normalizeProductListItem),
    pagination: response.meta?.pagination
      ? {
          limit: response.meta.pagination.limit,
          nextCursor: response.meta.pagination.next_cursor,
          hasMore: response.meta.pagination.has_more
        }
      : undefined
  };
}

export const getPublicProductDetail = cache(
  async (storeSlug: string, productSlug: string): Promise<PublicProductDetail> => {
    const response = await publicApiFetch<ApiPublicProductDetail>(
      `/api/v1/public/stores/${encodeURIComponent(storeSlug)}/products/${encodeURIComponent(productSlug)}`
    );

    return normalizeProductDetail(response.data);
  }
);

export function isPublicNotFoundError(error: unknown): error is ApiError {
  return error instanceof ApiError && error.status === 404;
}

async function publicApiFetch<T>(path: string): Promise<{ data: T; meta?: ApiSuccessResponse<T>["meta"] }> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    cache: "no-store"
  });
  const body = await parseJSON(response);

  if (!response.ok || isApiErrorResponse(body)) {
    if (isApiErrorResponse(body)) {
      throw toApiError(body, response.status);
    }

    throw new ApiError("Permintaan publik gagal diproses.", response.status, "UNKNOWN_API_ERROR");
  }

  if (!isApiSuccessResponse<T>(body)) {
    throw new ApiError("Respons API publik tidak valid.", response.status, "INVALID_API_RESPONSE");
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

function isApiSuccessResponse<T>(value: unknown): value is ApiSuccessResponse<T> {
  if (!value || typeof value !== "object") {
    return false;
  }

  const candidate = value as Partial<ApiSuccessResponse<T>>;
  return candidate.success === true && typeof candidate.message === "string";
}

function normalizeStore(store: ApiPublicStore): PublicStore {
  return {
    id: store.id,
    name: store.name,
    slug: store.slug,
    description: store.description,
    logoUrl: store.logo_url,
    bannerUrl: store.banner_url,
    phone: store.phone,
    whatsapp: store.whatsapp,
    city: store.city,
    province: store.province,
    businessHours: (store.business_hours ?? []).map(normalizeBusinessHour),
    seo: store.seo
  };
}

function normalizeBusinessHour(hour: ApiBusinessHour): PublicBusinessHour {
  return {
    dayOfWeek: hour.day_of_week,
    openTime: hour.open_time,
    closeTime: hour.close_time,
    isClosed: hour.is_closed
  };
}

function normalizeCategory(category: ApiPublicCategory): PublicCategory {
  return {
    id: category.id,
    name: category.name,
    slug: category.slug,
    imageUrl: category.image_url
  };
}

function normalizeProductListItem(product: ApiPublicProductListItem): PublicProductListItem {
  return {
    id: product.id,
    name: product.name,
    slug: product.slug,
    price: product.price,
    compareAtPrice: product.compare_at_price,
    primaryImageUrl: product.primary_image_url,
    stockStatus: product.stock_status
  };
}

function normalizeProductDetail(product: ApiPublicProductDetail): PublicProductDetail {
  return {
    id: product.id,
    name: product.name,
    slug: product.slug,
    description: product.description,
    price: product.price,
    compareAtPrice: product.compare_at_price,
    weightGram: product.weight_gram,
    images: (product.images ?? []).map((image) => ({
      url: image.url,
      altText: image.alt_text
    })),
    category: product.category,
    stock: {
      stockStatus: product.stock.stock_status,
      quantityAvailable: product.stock.quantity_available
    },
    store: product.store,
    seo: product.seo
  };
}
