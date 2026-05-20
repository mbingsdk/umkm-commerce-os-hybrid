"use client";

import { useQuery } from "@tanstack/react-query";
import {
  getCurrentPOSSession,
  getPOSTransactionDetail,
  listPOSProducts,
  listPOSTransactions
} from "@/features/pos/api/pos.api";
import type { POSProductFilters, POSTransactionFilters } from "@/features/pos/types";
import { queryKeys } from "@/lib/api/query-keys";
import { useTenantStore } from "@/lib/stores/tenant.store";

export function useCurrentPOSSession(enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.posCurrentSession(tenantId),
    queryFn: getCurrentPOSSession,
    enabled: enabled && !!tenantId,
    retry: false
  });
}

export function usePOSProducts(filters: POSProductFilters, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.posProducts(tenantId, filters),
    queryFn: () => listPOSProducts(filters),
    enabled: enabled && !!tenantId
  });
}

export function usePOSTransactions(filters: POSTransactionFilters, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.posTransactions(tenantId, filters),
    queryFn: () => listPOSTransactions(filters),
    enabled: enabled && !!tenantId
  });
}

export function usePOSTransactionDetail(transactionId: string | null, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.posTransaction(tenantId, transactionId),
    queryFn: () => getPOSTransactionDetail(transactionId as string),
    enabled: enabled && !!tenantId && !!transactionId
  });
}
