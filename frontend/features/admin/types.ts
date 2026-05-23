export type Pagination = {
  limit: number;
  nextCursor: string | null;
  hasMore: boolean;
};

export type AdminUser = {
  id: string;
  name: string;
  email: string;
  platformRole: "super_admin";
};

export type AdminPlan = {
  id: string;
  code: string;
  name: string;
  description?: string;
  priceMonthly: number;
  productLimit: number | null;
  staffLimit: number | null;
  canUsePos: boolean;
  canUseDiscovery: boolean;
  canUseCourier: boolean;
  canUseCustomDomain: boolean;
  isActive: boolean;
};

export type AdminStoreSummary = {
  id: string;
  name: string;
  slug: string;
  status: string;
  city?: string;
  createdAt?: string;
};

export type AdminOwner = {
  id: string;
  name: string;
  email: string;
  status: string;
};

export type AdminTenantCounts = {
  stores: number;
  products: number;
  orders: number;
  users: number;
  posTransactions: number;
};

export type AdminTenantListItem = {
  id: string;
  name: string;
  slug: string;
  status: string;
  plan?: AdminPlan | null;
  primaryStore?: AdminStoreSummary | null;
  owner?: AdminOwner | null;
  counts: AdminTenantCounts;
  createdAt: string;
};

export type AdminTenantDetail = {
  tenant: {
    id: string;
    name: string;
    slug: string;
    status: string;
    createdAt: string;
    updatedAt: string;
  };
  plan?: AdminPlan | null;
  primaryStore?: AdminStoreSummary | null;
  owner?: AdminOwner | null;
  counts: AdminTenantCounts;
  latestAudits: AdminAuditSnippet[];
};

export type AdminAuditSnippet = {
  id: string;
  actorUserId?: string;
  actorName?: string;
  action: string;
  targetType?: string;
  targetId?: string;
  createdAt: string;
};

export type AdminFeaturedItem = {
  id: string;
  itemType: "store" | "product";
  tenantId: string;
  storeId?: string | null;
  productId?: string | null;
  placement: "home" | "stores" | "products" | "category" | "city";
  sortOrder: number;
  startsAt?: string | null;
  endsAt?: string | null;
  isActive: boolean;
  targetName?: string;
  targetSlug?: string;
  createdBy?: string | null;
  createdAt: string;
  updatedAt: string;
};

export type AdminAuditLog = {
  id: string;
  actorUserId?: string;
  actorName?: string;
  action: string;
  targetType?: string;
  targetId?: string;
  beforeData?: unknown;
  afterData?: unknown;
  ipAddress?: string;
  userAgent?: string;
  createdAt: string;
};

export type ListResult<T> = {
  items: T[];
  pagination: Pagination;
};

export type AdminTenantFilters = {
  query?: string;
  status?: string;
  planId?: string;
  cursor?: string;
  limit?: number;
};

export type AdminFeaturedFilters = {
  itemType?: string;
  placement?: string;
  isActive?: string;
  cursor?: string;
  limit?: number;
};

export type AdminAuditFilters = {
  action?: string;
  targetType?: string;
  targetId?: string;
  dateFrom?: string;
  dateTo?: string;
  cursor?: string;
  limit?: number;
};

export type PlanFormInput = {
  code: string;
  name: string;
  description?: string;
  priceMonthly: number;
  productLimit: number | null;
  staffLimit: number | null;
  canUsePos: boolean;
  canUseDiscovery: boolean;
  canUseCourier: boolean;
  isActive: boolean;
};

export type FeaturedFormInput = {
  itemType: "store" | "product";
  tenantId: string;
  storeId?: string;
  productId?: string;
  placement: AdminFeaturedItem["placement"];
  sortOrder: number;
  startsAt?: string;
  endsAt?: string;
  isActive: boolean;
};
