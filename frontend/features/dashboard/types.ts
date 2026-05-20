export type DashboardSummary = {
  date: string;
  todaySales: number;
  todayOrderCount?: number | null;
  pendingOrdersCount?: number | null;
  lowStockCount?: number | null;
  posSalesToday?: number | null;
  onlineSalesToday?: number | null;
  expenseToday?: number | null;
  netEstimateToday?: number | null;
  visibleCards: string[];
  hiddenCards: string[];
  note?: string;
};

export type RecentOrder = {
  orderId: string;
  orderNumber: string;
  customerName: string;
  totalAmount: number;
  status: string;
  paymentStatus: string;
  createdAt: string;
};

export type LowStockProduct = {
  productId: string;
  productName: string;
  sku?: string;
  availableQuantity: number;
  lowStockThreshold: number;
};
