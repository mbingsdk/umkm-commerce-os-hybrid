import { apiFetch, apiFetchWithMeta } from "@/lib/api/client";
import type {
  AdminAuditLog,
  AdminAuditSnippet,
  AdminAuditFilters,
  AdminFeaturedFilters,
  AdminFeaturedItem,
  AdminOwner,
  AdminPlan,
  AdminStoreSummary,
  AdminTenantCounts,
  AdminTenantDetail,
  AdminTenantFilters,
  AdminTenantListItem,
  AdminUser,
  FeaturedFormInput,
  ListResult,
  Pagination,
  PlanFormInput
} from "@/features/admin/types";

type ApiPaginationMeta = {
  pagination?: {
    limit: number;
    next_cursor?: string | null;
    has_more: boolean;
  };
};

type ApiAdminUser = {
  id: string;
  name: string;
  email: string;
  platform_role: "super_admin";
};

type ApiPlan = {
  id: string;
  code: string;
  name: string;
  description?: string;
  price_monthly: number;
  product_limit?: number | null;
  staff_limit?: number | null;
  can_use_pos: boolean;
  can_use_discovery: boolean;
  can_use_courier: boolean;
  can_use_custom_domain: boolean;
  is_active: boolean;
};

type ApiStoreSummary = {
  id: string;
  name: string;
  slug: string;
  status: string;
  city?: string;
  created_at?: string;
};

type ApiOwner = {
  id: string;
  name: string;
  email: string;
  status: string;
};

type ApiTenantCounts = {
  stores: number;
  products: number;
  orders: number;
  users: number;
  pos_transactions: number;
};

type ApiTenantListItem = {
  id: string;
  name: string;
  slug: string;
  status: string;
  plan?: ApiPlan | null;
  primary_store?: ApiStoreSummary | null;
  owner?: ApiOwner | null;
  counts: ApiTenantCounts;
  created_at: string;
};

type ApiTenantDetail = {
  tenant: {
    id: string;
    name: string;
    slug: string;
    status: string;
    created_at: string;
    updated_at: string;
  };
  plan?: ApiPlan | null;
  primary_store?: ApiStoreSummary | null;
  owner?: ApiOwner | null;
  counts: ApiTenantCounts;
  latest_audits?: ApiAuditSnippet[];
};

type ApiAuditSnippet = {
  id: string;
  actor_user_id?: string;
  actor_name?: string;
  action: string;
  target_type?: string;
  target_id?: string;
  created_at: string;
};

type ApiFeaturedItem = {
  id: string;
  item_type: "store" | "product";
  tenant_id: string;
  store_id?: string | null;
  product_id?: string | null;
  placement: AdminFeaturedItem["placement"];
  sort_order: number;
  starts_at?: string | null;
  ends_at?: string | null;
  is_active: boolean;
  target_name?: string;
  target_slug?: string;
  created_by?: string | null;
  created_at: string;
  updated_at: string;
};

type ApiAuditLog = {
  id: string;
  actor_user_id?: string;
  actor_name?: string;
  action: string;
  target_type?: string;
  target_id?: string;
  before_data?: unknown;
  after_data?: unknown;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
};

export async function getAdminMe(): Promise<AdminUser> {
  const result = await apiFetch<{ user: ApiAdminUser }>("/api/v1/admin/me", {
    tenantScoped: false
  });
  return normalizeAdminUser(result.user);
}

export async function listAdminTenants(filters: AdminTenantFilters = {}): Promise<ListResult<AdminTenantListItem>> {
  const result = await apiFetchWithMeta<ApiTenantListItem[], ApiPaginationMeta>(
    `/api/v1/admin/tenants${toQueryString(filters)}`,
    { tenantScoped: false }
  );
  return {
    items: result.data.map(normalizeTenantListItem),
    pagination: normalizePagination(result.meta)
  };
}

export async function getAdminTenantDetail(tenantId: string): Promise<AdminTenantDetail> {
  const result = await apiFetch<ApiTenantDetail>(`/api/v1/admin/tenants/${tenantId}`, {
    tenantScoped: false
  });
  return normalizeTenantDetail(result);
}

export async function updateAdminTenantStatus(tenantId: string, input: { status: string; reason?: string }) {
  await apiFetch(`/api/v1/admin/tenants/${tenantId}/status`, {
    method: "PATCH",
    body: JSON.stringify({ status: input.status, reason: input.reason ?? "" }),
    tenantScoped: false
  });
}

export async function updateAdminTenantPlan(tenantId: string, input: { planId: string; reason?: string }) {
  await apiFetch(`/api/v1/admin/tenants/${tenantId}/plan`, {
    method: "PATCH",
    body: JSON.stringify({ plan_id: input.planId, reason: input.reason ?? "" }),
    tenantScoped: false
  });
}

export async function listAdminPlans(): Promise<AdminPlan[]> {
  const result = await apiFetch<ApiPlan[]>("/api/v1/admin/plans", {
    tenantScoped: false
  });
  return result.map(normalizePlan);
}

export async function createAdminPlan(input: PlanFormInput): Promise<AdminPlan> {
  const result = await apiFetch<ApiPlan>("/api/v1/admin/plans", {
    method: "POST",
    body: JSON.stringify(toPlanPayload(input)),
    tenantScoped: false
  });
  return normalizePlan(result);
}

export async function updateAdminPlan(planId: string, input: PlanFormInput): Promise<AdminPlan> {
  const result = await apiFetch<ApiPlan>(`/api/v1/admin/plans/${planId}`, {
    method: "PATCH",
    body: JSON.stringify(toPlanPayload(input)),
    tenantScoped: false
  });
  return normalizePlan(result);
}

export async function listAdminFeaturedItems(
  filters: AdminFeaturedFilters = {}
): Promise<ListResult<AdminFeaturedItem>> {
  const result = await apiFetchWithMeta<ApiFeaturedItem[], ApiPaginationMeta>(
    `/api/v1/admin/discovery/featured${toQueryString(filters)}`,
    { tenantScoped: false }
  );
  return {
    items: result.data.map(normalizeFeaturedItem),
    pagination: normalizePagination(result.meta)
  };
}

export async function createAdminFeaturedItem(input: FeaturedFormInput): Promise<AdminFeaturedItem> {
  const result = await apiFetch<ApiFeaturedItem>("/api/v1/admin/discovery/featured", {
    method: "POST",
    body: JSON.stringify(toFeaturedPayload(input)),
    tenantScoped: false
  });
  return normalizeFeaturedItem(result);
}

export async function updateAdminFeaturedItem(featuredId: string, input: FeaturedFormInput): Promise<AdminFeaturedItem> {
  const result = await apiFetch<ApiFeaturedItem>(`/api/v1/admin/discovery/featured/${featuredId}`, {
    method: "PATCH",
    body: JSON.stringify(toFeaturedPayload(input)),
    tenantScoped: false
  });
  return normalizeFeaturedItem(result);
}

export async function deleteAdminFeaturedItem(featuredId: string): Promise<void> {
  await apiFetch(`/api/v1/admin/discovery/featured/${featuredId}`, {
    method: "DELETE",
    tenantScoped: false
  });
}

export async function listAdminAuditLogs(filters: AdminAuditFilters = {}): Promise<ListResult<AdminAuditLog>> {
  const result = await apiFetchWithMeta<ApiAuditLog[], ApiPaginationMeta>(
    `/api/v1/admin/audit-logs${toQueryString(filters)}`,
    { tenantScoped: false }
  );
  return {
    items: result.data.map(normalizeAuditLog),
    pagination: normalizePagination(result.meta)
  };
}

function normalizeAdminUser(user: ApiAdminUser): AdminUser {
  return {
    id: user.id,
    name: user.name,
    email: user.email,
    platformRole: user.platform_role
  };
}

function normalizeTenantListItem(item: ApiTenantListItem): AdminTenantListItem {
  return {
    id: item.id,
    name: item.name,
    slug: item.slug,
    status: item.status,
    plan: item.plan ? normalizePlan(item.plan) : null,
    primaryStore: item.primary_store ? normalizeStore(item.primary_store) : null,
    owner: item.owner ? normalizeOwner(item.owner) : null,
    counts: normalizeCounts(item.counts),
    createdAt: item.created_at
  };
}

function normalizeTenantDetail(item: ApiTenantDetail): AdminTenantDetail {
  return {
    tenant: {
      id: item.tenant.id,
      name: item.tenant.name,
      slug: item.tenant.slug,
      status: item.tenant.status,
      createdAt: item.tenant.created_at,
      updatedAt: item.tenant.updated_at
    },
    plan: item.plan ? normalizePlan(item.plan) : null,
    primaryStore: item.primary_store ? normalizeStore(item.primary_store) : null,
    owner: item.owner ? normalizeOwner(item.owner) : null,
    counts: normalizeCounts(item.counts),
    latestAudits: (item.latest_audits ?? []).map(normalizeAuditSnippet)
  };
}

function normalizePlan(plan: ApiPlan): AdminPlan {
  return {
    id: plan.id,
    code: plan.code,
    name: plan.name,
    description: plan.description,
    priceMonthly: plan.price_monthly,
    productLimit: plan.product_limit ?? null,
    staffLimit: plan.staff_limit ?? null,
    canUsePos: plan.can_use_pos,
    canUseDiscovery: plan.can_use_discovery,
    canUseCourier: plan.can_use_courier,
    canUseCustomDomain: plan.can_use_custom_domain,
    isActive: plan.is_active
  };
}

function normalizeStore(store: ApiStoreSummary): AdminStoreSummary {
  return {
    id: store.id,
    name: store.name,
    slug: store.slug,
    status: store.status,
    city: store.city,
    createdAt: store.created_at
  };
}

function normalizeOwner(owner: ApiOwner): AdminOwner {
  return {
    id: owner.id,
    name: owner.name,
    email: owner.email,
    status: owner.status
  };
}

function normalizeCounts(counts: ApiTenantCounts): AdminTenantCounts {
  return {
    stores: counts.stores,
    products: counts.products,
    orders: counts.orders,
    users: counts.users,
    posTransactions: counts.pos_transactions
  };
}

function normalizeAuditSnippet(log: ApiAuditSnippet): AdminAuditSnippet {
  return {
    id: log.id,
    actorUserId: log.actor_user_id,
    actorName: log.actor_name,
    action: log.action,
    targetType: log.target_type,
    targetId: log.target_id,
    createdAt: log.created_at
  };
}

function normalizeFeaturedItem(item: ApiFeaturedItem): AdminFeaturedItem {
  return {
    id: item.id,
    itemType: item.item_type,
    tenantId: item.tenant_id,
    storeId: item.store_id,
    productId: item.product_id,
    placement: item.placement,
    sortOrder: item.sort_order,
    startsAt: item.starts_at,
    endsAt: item.ends_at,
    isActive: item.is_active,
    targetName: item.target_name,
    targetSlug: item.target_slug,
    createdBy: item.created_by,
    createdAt: item.created_at,
    updatedAt: item.updated_at
  };
}

function normalizeAuditLog(log: ApiAuditLog): AdminAuditLog {
  return {
    id: log.id,
    actorUserId: log.actor_user_id,
    actorName: log.actor_name,
    action: log.action,
    targetType: log.target_type,
    targetId: log.target_id,
    beforeData: log.before_data,
    afterData: log.after_data,
    ipAddress: log.ip_address,
    userAgent: log.user_agent,
    createdAt: log.created_at
  };
}

function normalizePagination(meta?: ApiPaginationMeta): Pagination {
  return {
    limit: meta?.pagination?.limit ?? 20,
    nextCursor: meta?.pagination?.next_cursor ?? null,
    hasMore: meta?.pagination?.has_more ?? false
  };
}

function toPlanPayload(input: PlanFormInput) {
  return {
    code: input.code,
    name: input.name,
    description: input.description ?? "",
    price_monthly: input.priceMonthly,
    product_limit: input.productLimit,
    staff_limit: input.staffLimit,
    can_use_pos: input.canUsePos,
    can_use_discovery: input.canUseDiscovery,
    can_use_courier: input.canUseCourier,
    is_active: input.isActive
  };
}

function toFeaturedPayload(input: FeaturedFormInput) {
  return {
    item_type: input.itemType,
    tenant_id: input.tenantId,
    store_id: input.storeId || "",
    product_id: input.productId || "",
    placement: input.placement,
    sort_order: input.sortOrder,
    starts_at: toOptionalRFC3339(input.startsAt),
    ends_at: toOptionalRFC3339(input.endsAt),
    is_active: input.isActive
  };
}

function toOptionalRFC3339(value?: string) {
  if (!value) {
    return undefined;
  }
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? undefined : parsed.toISOString();
}

function toQueryString(filters: Record<string, unknown>) {
  const params = new URLSearchParams();
  Object.entries(filters).forEach(([key, value]) => {
    if (value == null || value === "") {
      return;
    }
    const apiKey =
      key === "planId"
        ? "plan_id"
        : key === "itemType"
          ? "item_type"
          : key === "isActive"
            ? "is_active"
            : key === "targetType"
              ? "target_type"
              : key === "targetId"
                ? "target_id"
                : key === "dateFrom"
                  ? "date_from"
                  : key === "dateTo"
                    ? "date_to"
                    : key === "query"
                      ? "q"
                      : key;
    params.set(apiKey, String(value));
  });
  return params.size > 0 ? `?${params.toString()}` : "";
}
