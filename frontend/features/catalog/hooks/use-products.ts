"use client";

import { useQuery } from "@tanstack/react-query";
import {
  getProductDetail,
  listProducts,
  type ProductFilters
} from "@/features/catalog/api/products.api";
import { queryKeys } from "@/lib/api/query-keys";
import { useTenantStore } from "@/lib/stores/tenant.store";

export function useProducts(filters?: ProductFilters, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.products(tenantId, filters),
    queryFn: () => listProducts(filters),
    enabled: enabled && !!tenantId
  });
}

export function useProduct(productId: string | null, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.product(tenantId, productId),
    queryFn: () => getProductDetail(productId as string),
    enabled: enabled && !!tenantId && !!productId
  });
}
