"use client";

import { useQuery } from "@tanstack/react-query";
import { listProductMovements, listStocks } from "@/features/inventory/api/inventory.api";
import type { MovementFilters, StockFilters } from "@/features/inventory/types";
import { queryKeys } from "@/lib/api/query-keys";
import { useTenantStore } from "@/lib/stores/tenant.store";

export function useInventoryStocks(filters: StockFilters, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.inventoryStocks(tenantId, filters),
    queryFn: () => listStocks(filters),
    enabled: enabled && !!tenantId
  });
}

export function useProductMovements(productId: string | null, filters: MovementFilters, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.inventoryMovements(tenantId, productId, filters),
    queryFn: () => listProductMovements(productId as string, filters),
    enabled: enabled && !!tenantId && !!productId
  });
}
