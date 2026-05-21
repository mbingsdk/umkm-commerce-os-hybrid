"use client";

import { useQuery } from "@tanstack/react-query";
import { listCourierZones, listPublicCourierZones } from "@/features/courier/api/courier.api";
import { queryKeys } from "@/lib/api/query-keys";
import { useTenantStore } from "@/lib/stores/tenant.store";

export function useCourierZones(enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.courierZones(tenantId),
    queryFn: listCourierZones,
    enabled: enabled && !!tenantId
  });
}

export function usePublicCourierZones(storeSlug: string | null, enabled = true) {
  return useQuery({
    queryKey: queryKeys.publicCourierZones(storeSlug),
    queryFn: () => listPublicCourierZones(storeSlug as string),
    enabled: enabled && !!storeSlug
  });
}
