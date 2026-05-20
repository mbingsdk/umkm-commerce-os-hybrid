import { apiFetch, apiFetchWithMeta } from "@/lib/api/client";
import type {
  AdjustStockInput,
  AdjustStockResult,
  InventoryStock,
  ListMovementsResult,
  ListStocksResult,
  MovementFilters,
  Pagination,
  StockFilters,
  StockMovement,
  UpdateThresholdInput,
  UpdateThresholdResult
} from "@/features/inventory/types";

type ApiPaginationMeta = {
  pagination?: {
    limit: number;
    next_cursor?: string | null;
    has_more: boolean;
  };
};

type ApiInventoryStock = {
  product_id: string;
  name?: string;
  product_name?: string;
  sku?: string;
  category_id?: string | null;
  category_name?: string;
  primary_image_url?: string;
  quantity_on_hand: number;
  quantity_reserved: number;
  quantity_available: number;
  low_stock_threshold: number;
  is_low_stock: boolean;
  is_out_of_stock: boolean;
  updated_at: string;
};

type ApiStockMovement = {
  id: string;
  product_id: string;
  movement_type: string;
  quantity: number;
  balance_after: number;
  reference_type?: string;
  reference_id?: string | null;
  reason?: string;
  note?: string;
  created_by?: {
    id: string;
    name?: string;
  } | null;
  created_at: string;
};

type ApiAdjustStockResult = {
  product_id: string;
  quantity_on_hand: number;
  quantity_reserved: number;
  quantity_available: number;
  low_stock_threshold: number;
  movement_id: string;
};

type ApiUpdateThresholdResult = {
  product_id: string;
  low_stock_threshold: number;
};

export async function listStocks(filters: StockFilters = {}): Promise<ListStocksResult> {
  const searchParams = stockSearchParams(filters);
  const suffix = searchParams.size > 0 ? `?${searchParams.toString()}` : "";
  const result = await apiFetchWithMeta<ApiInventoryStock[], ApiPaginationMeta>(`/api/v1/inventory/stocks${suffix}`);

  return {
    stocks: result.data.map(normalizeStock),
    pagination: normalizePagination(result.meta)
  };
}

export async function listProductMovements(
  productId: string,
  filters: MovementFilters = {}
): Promise<ListMovementsResult> {
  const searchParams = new URLSearchParams();
  if (filters.cursor) {
    searchParams.set("cursor", filters.cursor);
  }
  if (filters.limit) {
    searchParams.set("limit", String(filters.limit));
  }

  const suffix = searchParams.size > 0 ? `?${searchParams.toString()}` : "";
  const result = await apiFetchWithMeta<ApiStockMovement[], ApiPaginationMeta>(
    `/api/v1/inventory/products/${productId}/movements${suffix}`
  );

  return {
    movements: result.data.map(normalizeMovement),
    pagination: normalizePagination(result.meta)
  };
}

export async function adjustProductStock(productId: string, input: AdjustStockInput): Promise<AdjustStockResult> {
  const result = await apiFetch<ApiAdjustStockResult>(`/api/v1/inventory/products/${productId}/adjust`, {
    method: "POST",
    body: JSON.stringify({
      adjustment_type: input.adjustmentType,
      quantity: input.quantity,
      reason: input.reason,
      note: input.note ?? ""
    })
  });

  return normalizeAdjustResult(result);
}

export async function updateStockThreshold(
  productId: string,
  input: UpdateThresholdInput
): Promise<UpdateThresholdResult> {
  const result = await apiFetch<ApiUpdateThresholdResult>(`/api/v1/inventory/products/${productId}/threshold`, {
    method: "PATCH",
    body: JSON.stringify({
      low_stock_threshold: input.lowStockThreshold
    })
  });

  return {
    productId: result.product_id,
    lowStockThreshold: result.low_stock_threshold
  };
}

function stockSearchParams(filters: StockFilters) {
  const searchParams = new URLSearchParams();

  if (filters.query) {
    searchParams.set("q", filters.query);
  }
  if (filters.lowStock) {
    searchParams.set("low_stock", "true");
  }
  if (filters.outOfStock) {
    searchParams.set("out_of_stock", "true");
  }
  if (filters.categoryId) {
    searchParams.set("category_id", filters.categoryId);
  }
  if (filters.cursor) {
    searchParams.set("cursor", filters.cursor);
  }
  if (filters.limit) {
    searchParams.set("limit", String(filters.limit));
  }

  return searchParams;
}

function normalizeStock(stock: ApiInventoryStock): InventoryStock {
  return {
    productId: stock.product_id,
    name: stock.name ?? stock.product_name ?? "Produk tanpa nama",
    sku: stock.sku,
    categoryId: stock.category_id,
    categoryName: stock.category_name,
    primaryImageUrl: stock.primary_image_url,
    quantityOnHand: stock.quantity_on_hand,
    quantityReserved: stock.quantity_reserved,
    quantityAvailable: stock.quantity_available,
    lowStockThreshold: stock.low_stock_threshold,
    isLowStock: stock.is_low_stock,
    isOutOfStock: stock.is_out_of_stock,
    updatedAt: stock.updated_at
  };
}

function normalizeMovement(movement: ApiStockMovement): StockMovement {
  return {
    id: movement.id,
    productId: movement.product_id,
    movementType: movement.movement_type,
    quantity: movement.quantity,
    balanceAfter: movement.balance_after,
    referenceType: movement.reference_type,
    referenceId: movement.reference_id,
    reason: movement.reason,
    note: movement.note,
    createdBy: movement.created_by,
    createdAt: movement.created_at
  };
}

function normalizeAdjustResult(result: ApiAdjustStockResult): AdjustStockResult {
  return {
    productId: result.product_id,
    quantityOnHand: result.quantity_on_hand,
    quantityReserved: result.quantity_reserved,
    quantityAvailable: result.quantity_available,
    lowStockThreshold: result.low_stock_threshold,
    movementId: result.movement_id
  };
}

function normalizePagination(meta?: ApiPaginationMeta): Pagination {
  return {
    limit: meta?.pagination?.limit ?? 20,
    nextCursor: meta?.pagination?.next_cursor ?? null,
    hasMore: meta?.pagination?.has_more ?? false
  };
}
