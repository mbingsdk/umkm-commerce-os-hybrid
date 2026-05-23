"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createAdminFeaturedItem,
  createAdminPlan,
  deleteAdminFeaturedItem,
  getAdminMe,
  getAdminTenantDetail,
  listAdminAuditLogs,
  listAdminFeaturedItems,
  listAdminPlans,
  listAdminTenants,
  updateAdminFeaturedItem,
  updateAdminPlan,
  updateAdminTenantPlan,
  updateAdminTenantStatus
} from "@/features/admin/api/admin.api";
import type {
  AdminAuditFilters,
  AdminFeaturedFilters,
  AdminTenantFilters,
  FeaturedFormInput,
  PlanFormInput
} from "@/features/admin/types";
import { queryKeys } from "@/lib/api/query-keys";

export function useAdminMe(enabled = true) {
  return useQuery({
    queryKey: queryKeys.adminMe,
    queryFn: getAdminMe,
    enabled
  });
}

export function useAdminTenants(filters: AdminTenantFilters, enabled = true) {
  return useQuery({
    queryKey: queryKeys.adminTenants(filters),
    queryFn: () => listAdminTenants(filters),
    enabled
  });
}

export function useAdminTenantDetail(tenantId: string | null, enabled = true) {
  return useQuery({
    queryKey: queryKeys.adminTenant(tenantId),
    queryFn: () => getAdminTenantDetail(tenantId as string),
    enabled: enabled && !!tenantId
  });
}

export function useAdminPlans(enabled = true) {
  return useQuery({
    queryKey: queryKeys.adminPlans,
    queryFn: listAdminPlans,
    enabled
  });
}

export function useAdminFeaturedItems(filters: AdminFeaturedFilters, enabled = true) {
  return useQuery({
    queryKey: queryKeys.adminFeatured(filters),
    queryFn: () => listAdminFeaturedItems(filters),
    enabled
  });
}

export function useAdminAuditLogs(filters: AdminAuditFilters, enabled = true) {
  return useQuery({
    queryKey: queryKeys.adminAuditLogs(filters),
    queryFn: () => listAdminAuditLogs(filters),
    enabled
  });
}

export function useAdminTenantMutations() {
  const queryClient = useQueryClient();
  return {
    updateStatus: useMutation({
      mutationFn: ({ tenantId, status, reason }: { tenantId: string; status: string; reason?: string }) =>
        updateAdminTenantStatus(tenantId, { status, reason }),
      onSuccess: async (_, variables) => {
        await queryClient.invalidateQueries({ queryKey: ["admin-tenants"] });
        await queryClient.invalidateQueries({ queryKey: queryKeys.adminTenant(variables.tenantId) });
      }
    }),
    updatePlan: useMutation({
      mutationFn: ({ tenantId, planId, reason }: { tenantId: string; planId: string; reason?: string }) =>
        updateAdminTenantPlan(tenantId, { planId, reason }),
      onSuccess: async (_, variables) => {
        await queryClient.invalidateQueries({ queryKey: ["admin-tenants"] });
        await queryClient.invalidateQueries({ queryKey: queryKeys.adminTenant(variables.tenantId) });
      }
    })
  };
}

export function useAdminPlanMutations() {
  const queryClient = useQueryClient();
  return {
    createPlan: useMutation({
      mutationFn: (input: PlanFormInput) => createAdminPlan(input),
      onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.adminPlans })
    }),
    updatePlan: useMutation({
      mutationFn: ({ planId, input }: { planId: string; input: PlanFormInput }) => updateAdminPlan(planId, input),
      onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.adminPlans })
    })
  };
}

export function useAdminFeaturedMutations() {
  const queryClient = useQueryClient();
  return {
    createFeatured: useMutation({
      mutationFn: (input: FeaturedFormInput) => createAdminFeaturedItem(input),
      onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin-featured"] })
    }),
    updateFeatured: useMutation({
      mutationFn: ({ featuredId, input }: { featuredId: string; input: FeaturedFormInput }) =>
        updateAdminFeaturedItem(featuredId, input),
      onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin-featured"] })
    }),
    deleteFeatured: useMutation({
      mutationFn: (featuredId: string) => deleteAdminFeaturedItem(featuredId),
      onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin-featured"] })
    })
  };
}
