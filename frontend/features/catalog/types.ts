export type Category = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  parentId?: string | null;
  sortOrder: number;
  isActive: boolean;
};

export type ProductStatus = "draft" | "active" | "inactive" | "archived";

export type ProductStock = {
  quantityOnHand: number;
  quantityReserved: number;
  quantityAvailable: number;
  lowStockThreshold: number;
};

export type ProductImage = {
  id: string;
  url: string;
  altText?: string;
  isPrimary: boolean;
  sortOrder: number;
};

export type ProductListItem = {
  id: string;
  categoryId?: string | null;
  name: string;
  slug: string;
  sku?: string;
  price: number;
  compareAtPrice?: number;
  status: ProductStatus;
  isDiscoverable: boolean;
  primaryImageUrl?: string;
  stock: ProductStock;
};

export type ProductDetail = ProductListItem & {
  categoryId?: string | null;
  description?: string;
  barcode?: string;
  costPrice?: number;
  weightGram: number;
  status: ProductStatus;
  trackInventory: boolean;
  allowBackorder: boolean;
  images: ProductImage[];
};

export type UploadAsset = {
  url: string;
  mimeType: string;
  size: number;
};
