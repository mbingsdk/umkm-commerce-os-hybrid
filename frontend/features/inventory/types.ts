export type Pagination = {
  limit: number;
  nextCursor?: string | null;
  hasMore: boolean;
};

export type StockFilters = {
  query?: string;
  lowStock?: boolean;
  outOfStock?: boolean;
  categoryId?: string;
  cursor?: string | null;
  limit?: number;
};

export type MovementFilters = {
  cursor?: string | null;
  limit?: number;
};

export type InventoryStock = {
  productId: string;
  name: string;
  sku?: string;
  categoryId?: string | null;
  categoryName?: string;
  primaryImageUrl?: string;
  quantityOnHand: number;
  quantityReserved: number;
  quantityAvailable: number;
  lowStockThreshold: number;
  isLowStock: boolean;
  isOutOfStock: boolean;
  updatedAt: string;
};

export type StockMovementActor = {
  id: string;
  name?: string;
};

export type StockMovement = {
  id: string;
  productId: string;
  movementType: string;
  quantity: number;
  balanceAfter: number;
  referenceType?: string;
  referenceId?: string | null;
  reason?: string;
  note?: string;
  createdBy?: StockMovementActor | null;
  createdAt: string;
};

export type ListStocksResult = {
  stocks: InventoryStock[];
  pagination: Pagination;
};

export type ListMovementsResult = {
  movements: StockMovement[];
  pagination: Pagination;
};

export type AdjustStockInput = {
  adjustmentType: "in" | "out";
  quantity: number;
  reason: string;
  note?: string;
};

export type AdjustStockResult = {
  productId: string;
  quantityOnHand: number;
  quantityReserved: number;
  quantityAvailable: number;
  lowStockThreshold: number;
  movementId: string;
};

export type UpdateThresholdInput = {
  lowStockThreshold: number;
};

export type UpdateThresholdResult = {
  productId: string;
  lowStockThreshold: number;
};
