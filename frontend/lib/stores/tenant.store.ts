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

type TenantState = {
  selectedTenantId: string | null;
  selectedStoreId: string | null;
  role: TenantRole | null;
  permissions: string[];
  selectTenant: (payload: {
    tenantId: string;
    storeId: string;
    role: TenantRole;
    permissions?: string[];
  }) => void;
  clearTenant: () => void;
};

export const useTenantStore = create<TenantState>()(
  persist(
    (set) => ({
      selectedTenantId: null,
      selectedStoreId: null,
      role: null,
      permissions: [],
      selectTenant: ({ tenantId, storeId, role, permissions = [] }) =>
        set({
          selectedTenantId: tenantId,
          selectedStoreId: storeId,
          role,
          permissions
        }),
      clearTenant: () =>
        set({
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
