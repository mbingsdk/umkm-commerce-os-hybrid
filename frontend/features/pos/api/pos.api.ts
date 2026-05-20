import { apiFetch, apiFetchWithMeta } from "@/lib/api/client";
import type {
  ClosePOSSessionInput,
  CreatePOSTransactionInput,
  ListPOSTransactionsResult,
  OpenPOSSessionInput,
  Pagination,
  POSProduct,
  POSProductFilters,
  POSSession,
  POSTransaction,
  POSTransactionDetail,
  POSTransactionFilters,
  POSTransactionItem
} from "@/features/pos/types";

type ApiPaginationMeta = {
  pagination?: {
    limit: number;
    next_cursor?: string | null;
    has_more: boolean;
  };
};

type ApiSession = {
  id: string;
  session_number: string;
  cashier_id: string;
  opening_cash: number;
  closing_cash?: number | null;
  expected_cash?: number | null;
  difference?: number | null;
  status: POSSession["status"];
  opened_at: string;
  closed_at?: string | null;
};

type ApiProduct = {
  product_id: string;
  name: string;
  sku?: string;
  barcode?: string;
  price: number;
  image?: string;
  available_stock: number;
  category_id?: string | null;
  category_name?: string;
};

type ApiTransaction = {
  id: string;
  transaction_number: string;
  session_id: string;
  cashier_id: string;
  subtotal: number;
  discount_total: number;
  tax_total: number;
  grand_total: number;
  payment_method: POSTransaction["paymentMethod"];
  amount_paid: number;
  change_amount: number;
  status: string;
  created_at: string;
};

type ApiTransactionItem = {
  id: string;
  product_id?: string | null;
  product_name: string;
  sku?: string;
  quantity: number;
  unit_price: number;
  discount_total: number;
  subtotal: number;
};

type ApiTransactionDetail = {
  transaction: ApiTransaction;
  items: ApiTransactionItem[];
};

export async function openPOSSession(input: OpenPOSSessionInput): Promise<POSSession> {
  const session = await apiFetch<ApiSession>("/api/v1/pos/sessions/open", {
    method: "POST",
    body: JSON.stringify({
      opening_cash_amount: input.openingCashAmount,
      note: input.note ?? ""
    })
  });

  return normalizeSession(session);
}

export async function getCurrentPOSSession(): Promise<POSSession> {
  const session = await apiFetch<ApiSession>("/api/v1/pos/sessions/current");
  return normalizeSession(session);
}

export async function closePOSSession(sessionId: string, input: ClosePOSSessionInput): Promise<POSSession> {
  const session = await apiFetch<ApiSession>(`/api/v1/pos/sessions/${sessionId}/close`, {
    method: "POST",
    body: JSON.stringify({
      closing_cash_amount: input.closingCashAmount,
      note: input.note ?? ""
    })
  });

  return normalizeSession(session);
}

export async function listPOSProducts(filters: POSProductFilters = {}): Promise<POSProduct[]> {
  const searchParams = new URLSearchParams();
  if (filters.query) {
    searchParams.set("q", filters.query);
  }
  if (filters.barcode) {
    searchParams.set("barcode", filters.barcode);
  }
  if (filters.limit) {
    searchParams.set("limit", String(filters.limit));
  }

  const suffix = searchParams.size > 0 ? `?${searchParams.toString()}` : "";
  const products = await apiFetch<ApiProduct[]>(`/api/v1/pos/products${suffix}`);
  return products.map(normalizeProduct);
}

export async function createPOSTransaction(input: CreatePOSTransactionInput): Promise<POSTransaction> {
  const transaction = await apiFetch<ApiTransaction>("/api/v1/pos/transactions", {
    method: "POST",
    headers: {
      "Idempotency-Key": input.idempotencyKey
    },
    body: JSON.stringify({
      session_id: input.sessionId,
      items: input.items.map((item) => ({
        product_id: item.productId,
        quantity: item.quantity
      })),
      payment_method: input.paymentMethod,
      amount_paid: input.amountPaid,
      note: input.note ?? ""
    })
  });

  return normalizeTransaction(transaction);
}

export async function listPOSTransactions(filters: POSTransactionFilters = {}): Promise<ListPOSTransactionsResult> {
  const searchParams = new URLSearchParams();
  if (filters.dateFrom) {
    searchParams.set("date_from", filters.dateFrom);
  }
  if (filters.dateTo) {
    searchParams.set("date_to", filters.dateTo);
  }
  if (filters.paymentMethod) {
    searchParams.set("payment_method", filters.paymentMethod);
  }
  if (filters.cashierId) {
    searchParams.set("cashier_id", filters.cashierId);
  }
  if (filters.cursor) {
    searchParams.set("cursor", filters.cursor);
  }
  if (filters.limit) {
    searchParams.set("limit", String(filters.limit));
  }

  const suffix = searchParams.size > 0 ? `?${searchParams.toString()}` : "";
  const result = await apiFetchWithMeta<ApiTransaction[], ApiPaginationMeta>(`/api/v1/pos/transactions${suffix}`);

  return {
    transactions: result.data.map(normalizeTransaction),
    pagination: normalizePagination(result.meta)
  };
}

export async function getPOSTransactionDetail(transactionId: string): Promise<POSTransactionDetail> {
  const detail = await apiFetch<ApiTransactionDetail>(`/api/v1/pos/transactions/${transactionId}`);

  return {
    transaction: normalizeTransaction(detail.transaction),
    items: detail.items.map(normalizeTransactionItem)
  };
}

function normalizeSession(session: ApiSession): POSSession {
  return {
    id: session.id,
    sessionNumber: session.session_number,
    cashierId: session.cashier_id,
    openingCash: session.opening_cash,
    closingCash: session.closing_cash,
    expectedCash: session.expected_cash,
    difference: session.difference,
    status: session.status,
    openedAt: session.opened_at,
    closedAt: session.closed_at
  };
}

function normalizeProduct(product: ApiProduct): POSProduct {
  return {
    productId: product.product_id,
    name: product.name,
    sku: product.sku,
    barcode: product.barcode,
    price: product.price,
    image: product.image,
    availableStock: product.available_stock,
    categoryId: product.category_id,
    categoryName: product.category_name
  };
}

function normalizeTransaction(transaction: ApiTransaction): POSTransaction {
  return {
    id: transaction.id,
    transactionNumber: transaction.transaction_number,
    sessionId: transaction.session_id,
    cashierId: transaction.cashier_id,
    subtotal: transaction.subtotal,
    discountTotal: transaction.discount_total,
    taxTotal: transaction.tax_total,
    grandTotal: transaction.grand_total,
    paymentMethod: transaction.payment_method,
    amountPaid: transaction.amount_paid,
    changeAmount: transaction.change_amount,
    status: transaction.status,
    createdAt: transaction.created_at
  };
}

function normalizeTransactionItem(item: ApiTransactionItem): POSTransactionItem {
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

function normalizePagination(meta?: ApiPaginationMeta): Pagination {
  return {
    limit: meta?.pagination?.limit ?? 20,
    nextCursor: meta?.pagination?.next_cursor ?? null,
    hasMore: meta?.pagination?.has_more ?? false
  };
}
