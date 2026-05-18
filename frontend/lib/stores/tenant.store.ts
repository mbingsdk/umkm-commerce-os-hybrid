"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

export type TenantRole =
  | "owner"
  | "manager"
  | "staff"
  | "cashier"
  | "inventory_staff"
  | "courier_admin"
  | "driver";

export type TenantStoreSummary = {
  id: string;
  name: string;
  slug: string;
  status: string;
};

export type TenantMembership = {
  id: string;
  name: string;
  slug: string;
  role: TenantRole;
  status: string;
  store: TenantStoreSummary;
  permissions: string[];
};

type TenantState = {
  tenants: TenantMembership[];
  selectedTenantId: string | null;
  selectedStoreId: string | null;
  role: TenantRole | null;
  permissions: string[];
  setTenants: (tenants: TenantMembership[]) => void;
  upsertTenant: (tenant: TenantMembership) => void;
  selectTenant: (tenant: TenantMembership) => void;
  clearTenant: () => void;
};

export const useTenantStore = create<TenantState>()(
  persist(
    (set) => ({
      tenants: [],
      selectedTenantId: null,
      selectedStoreId: null,
      role: null,
      permissions: [],
      setTenants: (tenants) => set({ tenants }),
      upsertTenant: (tenant) =>
        set((state) => ({
          tenants: [...state.tenants.filter((item) => item.id !== tenant.id), tenant]
        })),
      selectTenant: (tenant) =>
        set({
          selectedTenantId: tenant.id,
          selectedStoreId: tenant.store.id,
          role: tenant.role,
          permissions: tenant.permissions
        }),
      clearTenant: () =>
        set({
          tenants: [],
          selectedTenantId: null,
          selectedStoreId: null,
          role: null,
          permissions: []
        })
    }),
    {
      name: "umkm-tenant"
    }
  )
);
