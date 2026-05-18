export const queryKeys = {
  me: ["me"] as const,
  tenants: ["tenants"] as const,
  currentStore: (tenantId: string | null) => ["current-store", tenantId] as const,
  categories: (tenantId: string | null, filters?: { isActive?: boolean }) =>
    ["categories", tenantId, filters ?? {}] as const,
  products: (
    tenantId: string | null,
    filters?: { query?: string; status?: string; categoryId?: string }
  ) => ["products", tenantId, filters ?? {}] as const,
  product: (tenantId: string | null, productId: string | null) => ["product", tenantId, productId] as const
};
