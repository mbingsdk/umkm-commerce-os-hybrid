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
    ["pos-transaction", tenantId, transactionId] as const,
  dashboardSummary: (tenantId: string | null) => ["dashboard-summary", tenantId] as const,
  dashboardRecentOrders: (tenantId: string | null, limit?: number) =>
    ["dashboard-recent-orders", tenantId, limit ?? null] as const,
  dashboardLowStock: (tenantId: string | null, limit?: number) =>
    ["dashboard-low-stock", tenantId, limit ?? null] as const,
  financeSummary: (tenantId: string | null, filters?: Record<string, unknown>) =>
    ["finance-summary", tenantId, filters ?? {}] as const,
  financeDailyReport: (tenantId: string | null, filters?: Record<string, unknown>) =>
    ["finance-daily-report", tenantId, filters ?? {}] as const,
  financeMonthlyReport: (tenantId: string | null, filters?: Record<string, unknown>) =>
    ["finance-monthly-report", tenantId, filters ?? {}] as const,
  financeExpenses: (tenantId: string | null, filters?: Record<string, unknown>) =>
    ["finance-expenses", tenantId, filters ?? {}] as const,
  courierZones: (tenantId: string | null) => ["courier-zones", tenantId] as const,
  publicCourierZones: (storeSlug: string | null) => ["public-courier-zones", storeSlug] as const,
  shipments: (tenantId: string | null, filters?: Record<string, unknown>) =>
    ["shipments", tenantId, filters ?? {}] as const,
  shipment: (tenantId: string | null, shipmentId: string | null) => ["shipment", tenantId, shipmentId] as const,
  publicTracking: (storeSlug: string | null, orderNumber?: string, phone?: string) =>
    ["public-tracking", storeSlug, orderNumber ?? "", phone ?? ""] as const
};
