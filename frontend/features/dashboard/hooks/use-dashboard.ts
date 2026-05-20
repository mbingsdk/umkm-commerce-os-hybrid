"use client";

import { useQuery } from "@tanstack/react-query";
import { getDashboardSummary, getLowStock, getRecentOrders } from "@/features/dashboard/api/dashboard.api";
import { queryKeys } from "@/lib/api/query-keys";
import { useTenantStore } from "@/lib/stores/tenant.store";

export function useDashboardSummary(enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.dashboardSummary(tenantId),
    queryFn: getDashboardSummary,
    enabled: enabled && !!tenantId
  });
}

export function useDashboardRecentOrders(limit = 5, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.dashboardRecentOrders(tenantId, limit),
    queryFn: () => getRecentOrders(limit),
    enabled: enabled && !!tenantId
  });
}

export function useDashboardLowStock(limit = 5, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.dashboardLowStock(tenantId, limit),
    queryFn: () => getLowStock(limit),
    enabled: enabled && !!tenantId
  });
}
