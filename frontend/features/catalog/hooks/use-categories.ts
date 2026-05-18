"use client";

import { useQuery } from "@tanstack/react-query";
import { listCategories } from "@/features/catalog/api/categories.api";
import { queryKeys } from "@/lib/api/query-keys";
import { useTenantStore } from "@/lib/stores/tenant.store";

export function useCategories(filters?: { isActive?: boolean }, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.categories(tenantId, filters),
    queryFn: () => listCategories(filters),
    enabled: enabled && !!tenantId
  });
}
