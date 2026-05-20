import { apiFetch, apiFetchWithMeta } from "@/lib/api/client";
import type {
  CustomerSnapshot,
  ListOrdersResult,
  OrderDetail,
  OrderFilters,
  OrderItem,
  OrderListItem,
  OrderRecord,
  Pagination,
  PaymentConfirmation,
  PaymentSummary,
  ReservationSummary,
  ShippingAddress,
  StatusLog
} from "@/features/orders/types";

type ApiPaginationMeta = {
  pagination?: {
    limit: number;
    next_cursor?: string | null;
    has_more: boolean;
  };
};

type ApiOrderListItem = {
  id: string;
  order_number: string;
  source: string;
  status: OrderListItem["status"];
  payment_status: OrderListItem["paymentStatus"];
  shipment_status?: string | null;
  grand_total: number;
  customer_name: string;
  customer_phone: string;
  item_count: number;
  created_at: string;
  updated_at: string;
};

type ApiOrderRecord = {
  id: string;
  order_number: string;
  source: string;
  status: OrderRecord["status"];
  payment_status: OrderRecord["paymentStatus"];
  shipment_status?: string | null;
  subtotal: number;
  discount_total: number;
  shipping_cost: number;
  tax_total: number;
  grand_total: number;
  confirmed_at?: string | null;
  paid_at?: string | null;
  completed_at?: string | null;
  cancelled_at?: string | null;
  created_at: string;
  updated_at: string;
};

type ApiOrderItem = {
  id: string;
  product_id?: string | null;
  product_name: string;
  sku?: string;
  quantity: number;
  unit_price: number;
  discount_total: number;
  subtotal: number;
};

type ApiCustomerSnapshot = {
  id?: string | null;
  name: string;
  phone: string;
  email?: string;
};

type ApiShippingAddress = {
  address: string;
  city?: string;
  province?: string;
  postal_code?: string;
};

type ApiStatusLog = {
  id: string;
  from_status?: string | null;
  to_status: string;
  note?: string;
  created_by?: string | null;
  created_at: string;
};

type ApiPaymentSummary = {
  payment_status: PaymentSummary["paymentStatus"];
  subtotal: number;
  discount_total: number;
  shipping_cost: number;
  tax_total: number;
  grand_total: number;
  paid_at?: string | null;
};

type ApiReservationSummary = {
  status: string;
  quantity: number;
  count: number;
};

type ApiOrderDetail = {
  order: ApiOrderRecord;
  order_items: ApiOrderItem[];
  customer_snapshot: ApiCustomerSnapshot;
  shipping_address: ApiShippingAddress;
  status_logs: ApiStatusLog[];
  payment_summary: ApiPaymentSummary;
  stock_reservations: ApiReservationSummary[];
};

type ApiPaymentConfirmation = {
  id: string;
  order_id: string;
  payer_name: string;
  bank_name: string;
  transfer_amount: number;
  transfer_date: string;
  proof_url?: string;
  note?: string;
  status: PaymentConfirmation["status"];
  reviewed_by?: string | null;
  reviewed_at?: string | null;
  review_note?: string;
  created_at: string;
};

export type UpdateOrderStatusInput = {
  status: OrderRecord["status"];
  note?: string;
};

export type CancelOrderInput = {
  reason: string;
  note?: string;
};

export type ReviewPaymentInput = {
  paymentConfirmationId?: string;
  note?: string;
};

export async function listOrders(filters: OrderFilters = {}): Promise<ListOrdersResult> {
  const searchParams = new URLSearchParams();

  if (filters.query) {
    searchParams.set("q", filters.query);
  }
  if (filters.status) {
    searchParams.set("status", filters.status);
  }
  if (filters.paymentStatus) {
    searchParams.set("payment_status", filters.paymentStatus);
  }
  if (filters.source) {
    searchParams.set("source", filters.source);
  }
  if (filters.dateFrom) {
    searchParams.set("date_from", filters.dateFrom);
  }
  if (filters.dateTo) {
    searchParams.set("date_to", filters.dateTo);
  }
  if (filters.cursor) {
    searchParams.set("cursor", filters.cursor);
  }
  if (filters.limit) {
    searchParams.set("limit", String(filters.limit));
  }

  const suffix = searchParams.size > 0 ? `?${searchParams.toString()}` : "";
  const result = await apiFetchWithMeta<ApiOrderListItem[], ApiPaginationMeta>(`/api/v1/orders/${suffix}`);

  return {
    orders: result.data.map(normalizeOrderListItem),
    pagination: normalizePagination(result.meta)
  };
}

export async function getOrderDetail(orderId: string): Promise<OrderDetail> {
  const result = await apiFetch<ApiOrderDetail>(`/api/v1/orders/${orderId}`);
  return normalizeOrderDetail(result);
}

export async function updateOrderStatus(orderId: string, input: UpdateOrderStatusInput) {
  await apiFetch(`/api/v1/orders/${orderId}/status`, {
    method: "PATCH",
    body: JSON.stringify({
      status: input.status,
      note: input.note ?? ""
    })
  });
}

export async function cancelOrder(orderId: string, input: CancelOrderInput) {
  await apiFetch(`/api/v1/orders/${orderId}/cancel`, {
    method: "POST",
    body: JSON.stringify({
      reason: input.reason,
      note: input.note ?? ""
    })
  });
}

export async function listPaymentConfirmations(orderId: string): Promise<PaymentConfirmation[]> {
  const result = await apiFetch<ApiPaymentConfirmation[]>(`/api/v1/orders/${orderId}/payment-confirmations`);
  return result.map(normalizePaymentConfirmation);
}

export async function confirmPayment(orderId: string, input: ReviewPaymentInput) {
  await apiFetch(`/api/v1/orders/${orderId}/confirm-payment`, {
    method: "POST",
    body: JSON.stringify(toReviewPayload(input))
  });
}

export async function rejectPayment(orderId: string, input: ReviewPaymentInput) {
  await apiFetch(`/api/v1/orders/${orderId}/reject-payment`, {
    method: "POST",
    body: JSON.stringify(toReviewPayload(input))
  });
}

function normalizeOrderListItem(order: ApiOrderListItem): OrderListItem {
  return {
    id: order.id,
    orderNumber: order.order_number,
    source: order.source,
    status: order.status,
    paymentStatus: order.payment_status,
    shipmentStatus: order.shipment_status,
    grandTotal: order.grand_total,
    customerName: order.customer_name,
    customerPhone: order.customer_phone,
    itemCount: order.item_count,
    createdAt: order.created_at,
    updatedAt: order.updated_at
  };
}

function normalizeOrderRecord(order: ApiOrderRecord): OrderRecord {
  return {
    id: order.id,
    orderNumber: order.order_number,
    source: order.source,
    status: order.status,
    paymentStatus: order.payment_status,
    shipmentStatus: order.shipment_status,
    subtotal: order.subtotal,
    discountTotal: order.discount_total,
    shippingCost: order.shipping_cost,
    taxTotal: order.tax_total,
    grandTotal: order.grand_total,
    confirmedAt: order.confirmed_at,
    paidAt: order.paid_at,
    completedAt: order.completed_at,
    cancelledAt: order.cancelled_at,
    createdAt: order.created_at,
    updatedAt: order.updated_at
  };
}

function normalizeOrderDetail(detail: ApiOrderDetail): OrderDetail {
  return {
    order: normalizeOrderRecord(detail.order),
    items: detail.order_items.map(normalizeOrderItem),
    customer: normalizeCustomer(detail.customer_snapshot),
    shippingAddress: normalizeShippingAddress(detail.shipping_address),
    statusLogs: detail.status_logs.map(normalizeStatusLog),
    paymentSummary: normalizePaymentSummary(detail.payment_summary),
    stockReservations: detail.stock_reservations.map(normalizeReservationSummary)
  };
}

function normalizeOrderItem(item: ApiOrderItem): OrderItem {
  return {
    id: item.id,
    productId: item.product_id,
    productName: item.product_name,
    sku: item.sku,
    quantity: item.quantity,
    unitPrice: item.unit_price,
    discountTotal: item.discount_total,
    subtotal: item.subtotal
  };
}

function normalizeCustomer(customer: ApiCustomerSnapshot): CustomerSnapshot {
  return {
    id: customer.id,
    name: customer.name,
    phone: customer.phone,
    email: customer.email
  };
}

function normalizeShippingAddress(address: ApiShippingAddress): ShippingAddress {
  return {
    address: address.address,
    city: address.city,
    province: address.province,
    postalCode: address.postal_code
  };
}

function normalizeStatusLog(log: ApiStatusLog): StatusLog {
  return {
    id: log.id,
    fromStatus: log.from_status,
    toStatus: log.to_status,
    note: log.note,
    createdBy: log.created_by,
    createdAt: log.created_at
  };
}

function normalizePaymentSummary(summary: ApiPaymentSummary): PaymentSummary {
  return {
    paymentStatus: summary.payment_status,
    subtotal: summary.subtotal,
    discountTotal: summary.discount_total,
    shippingCost: summary.shipping_cost,
    taxTotal: summary.tax_total,
    grandTotal: summary.grand_total,
    paidAt: summary.paid_at
  };
}

function normalizeReservationSummary(summary: ApiReservationSummary): ReservationSummary {
  return {
    status: summary.status,
    quantity: summary.quantity,
    count: summary.count
  };
}

function normalizePaymentConfirmation(confirmation: ApiPaymentConfirmation): PaymentConfirmation {
  return {
    id: confirmation.id,
    orderId: confirmation.order_id,
    payerName: confirmation.payer_name,
    bankName: confirmation.bank_name,
    transferAmount: confirmation.transfer_amount,
    transferDate: confirmation.transfer_date,
    proofUrl: confirmation.proof_url,
    note: confirmation.note,
    status: confirmation.status,
    reviewedBy: confirmation.reviewed_by,
    reviewedAt: confirmation.reviewed_at,
    reviewNote: confirmation.review_note,
    createdAt: confirmation.created_at
  };
}

function normalizePagination(meta?: ApiPaginationMeta): Pagination {
  return {
    limit: meta?.pagination?.limit ?? 20,
    nextCursor: meta?.pagination?.next_cursor ?? null,
    hasMore: meta?.pagination?.has_more ?? false
  };
}

function toReviewPayload(input: ReviewPaymentInput) {
  return {
    payment_confirmation_id: input.paymentConfirmationId ?? "",
    note: input.note ?? ""
  };
}
