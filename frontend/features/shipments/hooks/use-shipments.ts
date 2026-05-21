"use client";

import { useQuery } from "@tanstack/react-query";
import {
  getPublicOrderTracking,
  getShipmentDetail,
  listShipments
} from "@/features/shipments/api/shipments.api";
import type { ShipmentFilters } from "@/features/shipments/types";
import { queryKeys } from "@/lib/api/query-keys";
import { useTenantStore } from "@/lib/stores/tenant.store";

export function useShipments(filters: ShipmentFilters, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.shipments(tenantId, filters),
    queryFn: () => listShipments(filters),
    enabled: enabled && !!tenantId
  });
}

export function useShipmentDetail(shipmentId: string | null, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.shipment(tenantId, shipmentId),
    queryFn: () => getShipmentDetail(shipmentId as string),
    enabled: enabled && !!tenantId && !!shipmentId
  });
}

export function usePublicOrderTracking(
  storeSlug: string | null,
  orderNumber: string,
  phone: string,
  enabled = true
) {
  return useQuery({
    queryKey: queryKeys.publicTracking(storeSlug, orderNumber, phone),
    queryFn: () => getPublicOrderTracking(storeSlug as string, orderNumber, phone),
    enabled: enabled && !!storeSlug && !!orderNumber && !!phone,
    retry: false
  });
}
