export type OrderStatus =
  | "pending"
  | "confirmed"
  | "processing"
  | "ready_to_ship"
  | "shipped"
  | "delivered"
  | "completed"
  | "cancelled"
  | "returned"
  | "refunded";

export type PaymentStatus = "unpaid" | "waiting_confirmation" | "paid" | "failed" | "refunded";

export type OrderSource =
  | "storefront"
  | "marketplace_discovery"
  | "pos"
  | "whatsapp_manual"
  | "admin_manual"
  | "marketplace_sync"
  | "reseller"
  | "api_partner";

export type OrderListItem = {
  id: string;
  orderNumber: string;
  source: OrderSource | string;
  status: OrderStatus;
  paymentStatus: PaymentStatus;
  shipmentStatus?: string | null;
  grandTotal: number;
  customerName: string;
  customerPhone: string;
  itemCount: number;
  createdAt: string;
  updatedAt: string;
};

export type OrderRecord = {
  id: string;
  orderNumber: string;
  source: OrderSource | string;
  status: OrderStatus;
  paymentStatus: PaymentStatus;
  shipmentStatus?: string | null;
  subtotal: number;
  discountTotal: number;
  shippingCost: number;
  taxTotal: number;
  grandTotal: number;
  confirmedAt?: string | null;
  paidAt?: string | null;
  completedAt?: string | null;
  cancelledAt?: string | null;
  createdAt: string;
  updatedAt: string;
};

export type OrderItem = {
  id: string;
  productId?: string | null;
  productName: string;
  sku?: string;
  quantity: number;
  unitPrice: number;
  discountTotal: number;
  subtotal: number;
};

export type CustomerSnapshot = {
  id?: string | null;
  name: string;
  phone: string;
  email?: string;
};

export type ShippingAddress = {
  address: string;
  city?: string;
  province?: string;
  postalCode?: string;
};

export type StatusLog = {
  id: string;
  fromStatus?: OrderStatus | string | null;
  toStatus: OrderStatus | string;
  note?: string;
  createdBy?: string | null;
  createdAt: string;
};

export type PaymentSummary = {
  paymentStatus: PaymentStatus;
  subtotal: number;
  discountTotal: number;
  shippingCost: number;
  taxTotal: number;
  grandTotal: number;
  paidAt?: string | null;
};

export type ReservationSummary = {
  status: string;
  quantity: number;
  count: number;
};

export type OrderDetail = {
  order: OrderRecord;
  items: OrderItem[];
  customer: CustomerSnapshot;
  shippingAddress: ShippingAddress;
  statusLogs: StatusLog[];
  paymentSummary: PaymentSummary;
  stockReservations: ReservationSummary[];
};

export type Pagination = {
  limit: number;
  nextCursor?: string | null;
  hasMore: boolean;
};

export type ListOrdersResult = {
  orders: OrderListItem[];
  pagination: Pagination;
};

export type OrderFilters = {
  query?: string;
  status?: OrderStatus | "";
  paymentStatus?: PaymentStatus | "";
  source?: OrderSource | "";
  dateFrom?: string;
  dateTo?: string;
  cursor?: string | null;
  limit?: number;
};

export type PaymentConfirmationStatus = "pending" | "confirmed" | "rejected";

export type PaymentConfirmation = {
  id: string;
  orderId: string;
  payerName: string;
  bankName: string;
  transferAmount: number;
  transferDate: string;
  proofUrl?: string;
  note?: string;
  status: PaymentConfirmationStatus;
  reviewedBy?: string | null;
  reviewedAt?: string | null;
  reviewNote?: string;
  createdAt: string;
};
