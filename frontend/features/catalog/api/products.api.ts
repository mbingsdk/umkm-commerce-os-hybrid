import { apiFetch } from "@/lib/api/client";
import type {
  ProductDetail,
  ProductImage,
  ProductListItem,
  ProductStatus,
  UploadAsset
} from "@/features/catalog/types";

type ApiStock = {
  quantity_on_hand: number;
  quantity_reserved: number;
  quantity_available: number;
  low_stock_threshold?: number;
};

type ApiImage = {
  id: string;
  url: string;
  alt_text?: string;
  is_primary: boolean;
  sort_order: number;
};

type ApiProductListItem = {
  id: string;
  category_id?: string | null;
  name: string;
  slug: string;
  sku?: string;
  price: number;
  compare_at_price?: number;
  status: ProductStatus;
  is_discoverable: boolean;
  primary_image_url?: string;
  stock: ApiStock;
};

type ApiProductDetail = ApiProductListItem & {
  description?: string;
  barcode?: string;
  cost_price?: number;
  weight_gram: number;
  track_inventory: boolean;
  allow_backorder: boolean;
  images: ApiImage[];
};

type ApiProductSummary = {
  id: string;
  name: string;
  slug: string;
  status: ProductStatus;
  stock: ApiStock;
};

type ApiUploadAsset = {
  url: string;
  mime_type: string;
  size: number;
};

export type ProductFilters = {
  query?: string;
  status?: ProductStatus | "";
  categoryId?: string;
};

export type CreateProductInput = {
  categoryId?: string;
  name: string;
  slug: string;
  description?: string;
  sku?: string;
  barcode?: string;
  price: number;
  compareAtPrice?: number | null;
  costPrice?: number | null;
  weightGram: number;
  status: ProductStatus;
  isDiscoverable: boolean;
  trackInventory: boolean;
  allowBackorder: boolean;
  initialStock: number;
};

export type UpdateProductInput = Omit<CreateProductInput, "initialStock">;

export type AttachProductImageInput = {
  file: File;
  altText?: string;
  isPrimary?: boolean;
  sortOrder?: number;
};

export async function listProducts(filters?: ProductFilters): Promise<ProductListItem[]> {
  const searchParams = new URLSearchParams();

  if (filters?.query) {
    searchParams.set("q", filters.query);
  }
  if (filters?.status) {
    searchParams.set("status", filters.status);
  }
  if (filters?.categoryId) {
    searchParams.set("category_id", filters.categoryId);
  }

  const suffix = searchParams.size > 0 ? `?${searchParams.toString()}` : "";
  const products = await apiFetch<ApiProductListItem[]>(`/api/v1/products${suffix}`);

  return products.map(normalizeProductListItem);
}

export async function getProductDetail(productId: string): Promise<ProductDetail> {
  const product = await apiFetch<ApiProductDetail>(`/api/v1/products/${productId}`);
  return normalizeProductDetail(product);
}

export async function createProduct(input: CreateProductInput): Promise<{ id: string }> {
  const product = await apiFetch<ApiProductSummary>("/api/v1/products", {
    method: "POST",
    body: JSON.stringify(toCreatePayload(input))
  });

  return { id: product.id };
}

export async function updateProduct(productId: string, input: UpdateProductInput) {
  await apiFetch<ApiProductSummary>(`/api/v1/products/${productId}`, {
    method: "PATCH",
    body: JSON.stringify(toUpdatePayload(input))
  });
}

export async function deleteProduct(productId: string) {
  await apiFetch<void>(`/api/v1/products/${productId}`, {
    method: "DELETE"
  });
}

export async function uploadProductImage(file: File): Promise<UploadAsset> {
  const body = new FormData();
  body.append("file", file);
  body.append("folder", "products");

  const asset = await apiFetch<ApiUploadAsset>("/api/v1/uploads", {
    method: "POST",
    body
  });

  return {
    url: asset.url,
    mimeType: asset.mime_type,
    size: asset.size
  };
}

export async function attachProductImage(productId: string, input: AttachProductImageInput): Promise<ProductImage> {
  const body = new FormData();
  body.append("file", input.file);
  body.append("alt_text", input.altText ?? "");
  body.append("is_primary", String(input.isPrimary ?? false));
  body.append("sort_order", String(input.sortOrder ?? 0));

  const image = await apiFetch<ApiImage>(`/api/v1/products/${productId}/images`, {
    method: "POST",
    body
  });

  return normalizeImage(image);
}

export async function deleteProductImage(productId: string, imageId: string) {
  await apiFetch<void>(`/api/v1/products/${productId}/images/${imageId}`, {
    method: "DELETE"
  });
}

function normalizeProductListItem(product: ApiProductListItem): ProductListItem {
  return {
    id: product.id,
    categoryId: product.category_id,
    name: product.name,
    slug: product.slug,
    sku: product.sku,
    price: product.price,
    compareAtPrice: product.compare_at_price,
    status: product.status,
    isDiscoverable: product.is_discoverable,
    primaryImageUrl: product.primary_image_url,
    stock: normalizeStock(product.stock)
  };
}

function normalizeProductDetail(product: ApiProductDetail): ProductDetail {
  return {
    ...normalizeProductListItem(product),
    description: product.description,
    barcode: product.barcode,
    costPrice: product.cost_price,
    weightGram: product.weight_gram,
    trackInventory: product.track_inventory,
    allowBackorder: product.allow_backorder,
    images: product.images.map(normalizeImage)
  };
}

function normalizeStock(stock: ApiStock) {
  return {
    quantityOnHand: stock.quantity_on_hand,
    quantityReserved: stock.quantity_reserved,
    quantityAvailable: stock.quantity_available,
    lowStockThreshold: stock.low_stock_threshold ?? 0
  };
}

function normalizeImage(image: ApiImage): ProductImage {
  return {
    id: image.id,
    url: image.url,
    altText: image.alt_text,
    isPrimary: image.is_primary,
    sortOrder: image.sort_order
  };
}

function toCreatePayload(input: CreateProductInput) {
  return {
    category_id: input.categoryId || null,
    name: input.name,
    slug: input.slug,
    description: input.description ?? "",
    sku: input.sku ?? "",
    barcode: input.barcode ?? "",
    price: input.price,
    compare_at_price: input.compareAtPrice ?? null,
    cost_price: input.costPrice ?? null,
    weight_gram: input.weightGram,
    status: input.status,
    is_discoverable: input.isDiscoverable,
    track_inventory: input.trackInventory,
    allow_backorder: input.allowBackorder,
    initial_stock: input.initialStock
  };
}

function toUpdatePayload(input: UpdateProductInput) {
  return {
    category_id: input.categoryId || null,
    name: input.name,
    slug: input.slug,
    description: input.description ?? "",
    sku: input.sku ?? "",
    barcode: input.barcode ?? "",
    price: input.price,
    compare_at_price: input.compareAtPrice ?? null,
    cost_price: input.costPrice ?? null,
    weight_gram: input.weightGram,
    status: input.status,
    is_discoverable: input.isDiscoverable,
    track_inventory: input.trackInventory,
    allow_backorder: input.allowBackorder
  };
}
