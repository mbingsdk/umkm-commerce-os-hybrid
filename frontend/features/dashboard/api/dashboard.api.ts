import { apiFetch } from "@/lib/api/client";
import type { DashboardSummary, LowStockProduct, RecentOrder } from "@/features/dashboard/types";

type ApiDashboardSummary = {
  date: string;
  today_sales: number;
  today_order_count?: number | null;
  pending_orders_count?: number | null;
  low_stock_count?: number | null;
  pos_sales_today?: number | null;
  online_sales_today?: number | null;
  expense_today?: number | null;
  net_estimate_today?: number | null;
  visible_cards?: string[];
  hidden_cards?: string[];
  note?: string;
};

type ApiRecentOrder = {
  order_id: string;
  order_number: string;
  customer_name: string;
  total_amount: number;
  status: string;
  payment_status: string;
  created_at: string;
};

type ApiLowStockProduct = {
  product_id: string;
  product_name: string;
  sku?: string;
  available_quantity: number;
  low_stock_threshold: number;
};

export async function getDashboardSummary(): Promise<DashboardSummary> {
  const result = await apiFetch<ApiDashboardSummary>("/api/v1/dashboard/summary");
  return normalizeSummary(result);
}

export async function getRecentOrders(limit = 5): Promise<RecentOrder[]> {
  const result = await apiFetch<ApiRecentOrder[]>(`/api/v1/dashboard/recent-orders?limit=${limit}`);
  return result.map(normalizeRecentOrder);
}

export async function getLowStock(limit = 5): Promise<LowStockProduct[]> {
  const result = await apiFetch<ApiLowStockProduct[]>(`/api/v1/dashboard/low-stock?limit=${limit}`);
  return result.map(normalizeLowStock);
}

function normalizeSummary(summary: ApiDashboardSummary): DashboardSummary {
  return {
    date: summary.date,
    todaySales: summary.today_sales,
    todayOrderCount: summary.today_order_count,
    pendingOrdersCount: summary.pending_orders_count,
    lowStockCount: summary.low_stock_count,
    posSalesToday: summary.pos_sales_today,
    onlineSalesToday: summary.online_sales_today,
    expenseToday: summary.expense_today,
    netEstimateToday: summary.net_estimate_today,
    visibleCards: summary.visible_cards ?? [],
    hiddenCards: summary.hidden_cards ?? [],
    note: summary.note
  };
}

function normalizeRecentOrder(order: ApiRecentOrder): RecentOrder {
  return {
    orderId: order.order_id,
    orderNumber: order.order_number,
    customerName: order.customer_name,
    totalAmount: order.total_amount,
    status: order.status,
    paymentStatus: order.payment_status,
    createdAt: order.created_at
  };
}

function normalizeLowStock(item: ApiLowStockProduct): LowStockProduct {
  return {
    productId: item.product_id,
    productName: item.product_name,
    sku: item.sku,
    availableQuantity: item.available_quantity,
    lowStockThreshold: item.low_stock_threshold
  };
}
