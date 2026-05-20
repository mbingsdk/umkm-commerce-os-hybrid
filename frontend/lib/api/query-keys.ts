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
  product: (tenantId: string | null, productId: string | null) => ["product", tenantId, productId] as const,
  orders: (tenantId: string | null, filters?: Record<string, unknown>) => ["orders", tenantId, filters ?? {}] as const,
  order: (tenantId: string | null, orderId: string | null) => ["order", tenantId, orderId] as const,
  paymentConfirmations: (tenantId: string | null, orderId: string | null) =>
    ["payment-confirmations", tenantId, orderId] as const,
  inventoryStocks: (tenantId: string | null, filters?: Record<string, unknown>) =>
    ["inventory-stocks", tenantId, filters ?? {}] as const,
  inventoryMovements: (tenantId: string | null, productId: string | null, filters?: Record<string, unknown>) =>
    ["inventory-movements", tenantId, productId, filters ?? {}] as const,
  posCurrentSession: (tenantId: string | null) => ["pos-current-session", tenantId] as const,
  posProducts: (tenantId: string | null, filters?: Record<string, unknown>) =>
    ["pos-products", tenantId, filters ?? {}] as const,
  posTransactions: (tenantId: string | null, filters?: Record<string, unknown>) =>
    ["pos-transactions", tenantId, filters ?? {}] as const,
  posTransaction: (tenantId: string | null, transactionId: string | null) =>
    ["pos-transaction", tenantId, transactionId] as const
};
