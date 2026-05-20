export type POSPaymentMethod = "cash" | "qris_manual";

export type Pagination = {
  limit: number;
  nextCursor?: string | null;
  hasMore: boolean;
};

export type POSSession = {
  id: string;
  sessionNumber: string;
  cashierId: string;
  openingCash: number;
  closingCash?: number | null;
  expectedCash?: number | null;
  difference?: number | null;
  status: "open" | "closed" | "cancelled" | string;
  openedAt: string;
  closedAt?: string | null;
};

export type POSProduct = {
  productId: string;
  name: string;
  sku?: string;
  barcode?: string;
  price: number;
  image?: string;
  availableStock: number;
  categoryId?: string | null;
  categoryName?: string;
};

export type POSCartLine = {
  productId: string;
  name: string;
  sku?: string;
  price: number;
  availableStock: number;
  quantity: number;
};

export type POSTransaction = {
  id: string;
  transactionNumber: string;
  sessionId: string;
  cashierId: string;
  subtotal: number;
  discountTotal: number;
  taxTotal: number;
  grandTotal: number;
  paymentMethod: POSPaymentMethod | string;
  amountPaid: number;
  changeAmount: number;
  status: string;
  createdAt: string;
};

export type POSTransactionItem = {
  id: string;
  productId?: string | null;
  productName: string;
  sku?: string;
  quantity: number;
  unitPrice: number;
  discountTotal: number;
  subtotal: number;
};

export type POSTransactionDetail = {
  transaction: POSTransaction;
  items: POSTransactionItem[];
};

export type POSProductFilters = {
  query?: string;
  barcode?: string;
  limit?: number;
};

export type POSTransactionFilters = {
  dateFrom?: string;
  dateTo?: string;
  paymentMethod?: POSPaymentMethod | "";
  cashierId?: string;
  cursor?: string | null;
  limit?: number;
};

export type ListPOSTransactionsResult = {
  transactions: POSTransaction[];
  pagination: Pagination;
};

export type OpenPOSSessionInput = {
  openingCashAmount: number;
  note?: string;
};

export type ClosePOSSessionInput = {
  closingCashAmount: number;
  note?: string;
};

export type CreatePOSTransactionInput = {
  sessionId: string;
  items: Array<{
    productId: string;
    quantity: number;
  }>;
  paymentMethod: POSPaymentMethod;
  amountPaid: number;
  note?: string;
  idempotencyKey: string;
};
