"use client";

import { useQuery } from "@tanstack/react-query";
import {
  getDailyReport,
  getFinanceSummary,
  getMonthlyReport,
  listExpenses
} from "@/features/finance/api/finance.api";
import type { ExpenseFilters, FinanceDateFilters, MonthlyReportFilters } from "@/features/finance/types";
import { queryKeys } from "@/lib/api/query-keys";
import { useTenantStore } from "@/lib/stores/tenant.store";

export function useFinanceSummary(filters: FinanceDateFilters, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.financeSummary(tenantId, filters),
    queryFn: () => getFinanceSummary(filters),
    enabled: enabled && !!tenantId
  });
}

export function useDailyReport(filters: FinanceDateFilters, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.financeDailyReport(tenantId, filters),
    queryFn: () => getDailyReport(filters),
    enabled: enabled && !!tenantId
  });
}

export function useMonthlyReport(filters: MonthlyReportFilters, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.financeMonthlyReport(tenantId, filters),
    queryFn: () => getMonthlyReport(filters),
    enabled: enabled && !!tenantId
  });
}

export function useExpenses(filters: ExpenseFilters, enabled = true) {
  const tenantId = useTenantStore((state) => state.selectedTenantId);

  return useQuery({
    queryKey: queryKeys.financeExpenses(tenantId, filters),
    queryFn: () => listExpenses(filters),
    enabled: enabled && !!tenantId
  });
}
