export type StockStatus = "in_stock" | "low_stock" | "out_of_stock";

export type PublicBusinessHour = {
  dayOfWeek: number;
  openTime?: string;
  closeTime?: string;
  isClosed: boolean;
};

export type PublicStore = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  logoUrl?: string;
  bannerUrl?: string;
  phone?: string;
  whatsapp?: string;
  city?: string;
  province?: string;
  businessHours: PublicBusinessHour[];
  seo?: {
    title?: string;
    description?: string;
  };
};

export type PublicCategory = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  imageUrl?: string;
};

export type PublicProductListItem = {
  id: string;
  name: string;
  slug: string;
  price: number;
  compareAtPrice?: number;
  primaryImageUrl?: string;
  stockStatus: StockStatus;
};

export type PublicProductImage = {
  url: string;
  altText?: string;
};

export type PublicProductCategory = {
  id: string;
  name: string;
  slug: string;
};

export type PublicProductDetail = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  price: number;
  compareAtPrice?: number;
  weightGram: number;
  images: PublicProductImage[];
  category?: PublicProductCategory;
  stock: {
    stockStatus: StockStatus;
    quantityAvailable: number;
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

export type PublicProductFilters = {
  query?: string;
  categorySlug?: string;
  inStock?: boolean;
  limit?: number;
  cursor?: string;
};

export type PublicProductListResult = {
  items: PublicProductListItem[];
  pagination?: {
    limit: number;
    nextCursor?: string | null;
    hasMore: boolean;
  };
};
