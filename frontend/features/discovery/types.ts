export type DiscoveryStore = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  logoUrl?: string;
  bannerUrl?: string;
  city?: string;
  province?: string;
  storeUrl: string;
};

export type DiscoveryProduct = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  price: number;
  primaryImageUrl?: string;
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
  storeUrl: string;
  productUrl: string;
};

export type DiscoveryAggregate = {
  name?: string;
  slug?: string;
  city?: string;
  count: number;
};

export type DiscoveryHome = {
  featuredStores: DiscoveryStore[];
  featuredProducts: DiscoveryProduct[];
  latestStores: DiscoveryStore[];
  latestProducts: DiscoveryProduct[];
  popularCategories: DiscoveryAggregate[];
  popularCities: DiscoveryAggregate[];
};

export type Pagination = {
  limit: number;
  nextCursor?: string | null;
  hasMore: boolean;
};

export type DiscoveryListResult<T> = {
  items: T[];
  pagination: Pagination;
};

export type DiscoveryStoreFilters = {
  query?: string;
  city?: string;
  category?: string;
  cursor?: string;
  limit?: number;
};

export type DiscoveryProductFilters = DiscoveryStoreFilters & {
  priceMin?: string | number;
  priceMax?: string | number;
};

export type DiscoverySearchFilters = {
  query?: string;
  type?: "all" | "stores" | "products";
  city?: string;
  category?: string;
  limit?: number;
};

export type DiscoverySearchResult = {
  stores: DiscoveryStore[];
  products: DiscoveryProduct[];
};
