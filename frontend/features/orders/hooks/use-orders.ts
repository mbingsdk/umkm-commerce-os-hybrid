"use client";

import { useQuery } from "@tanstack/react-query";
import {
  getOrderDetail,
  listOrders,
  listPaymentConfirmations
} from "@/features/orders/api/orders.api";
import type { OrderFilters } from "@/features/orders/types";
import { queryKeys } from "@/lib/api/query-keys";
import { useTenantStore } from "@/lib/stores/tenant.store";

export function useOrders(filters: OrderFilters, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.orders(tenantId, filters),
    queryFn: () => listOrders(filters),
    enabled: enabled && !!tenantId
  });
}

export function useOrderDetail(orderId: string | null, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.order(tenantId, orderId),
    queryFn: () => getOrderDetail(orderId as string),
    enabled: enabled && !!tenantId && !!orderId
  });
}

export function usePaymentConfirmations(orderId: string | null, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.paymentConfirmations(tenantId, orderId),
    queryFn: () => listPaymentConfirmations(orderId as string),
    enabled: enabled && !!tenantId && !!orderId
  });
}
