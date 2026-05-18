"use client";

import { useTenantStore } from "@/lib/stores/tenant.store";

export function TenantSelector() {
  const tenants = useTenantStore((state) => state.tenants);
  const selectedTenantId = useTenantStore((state) => state.selectedTenantId);
  const selectTenant = useTenantStore((state) => state.selectTenant);

  if (tenants.length === 0) {
    return (
      <button
        type="button"
        className="h-10 rounded-xl border border-neutral-200 bg-neutral-50 px-3 text-sm text-neutral-500"
        disabled
      >
        Belum ada tenant
      </button>
    );
  }

  return (
    <label className="flex items-center gap-2 text-sm text-neutral-600">
      <span className="sr-only">Pilih tenant</span>
      <select
        className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-900 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
        value={selectedTenantId ?? ""}
        onChange={(event) => {
          const tenant = tenants.find((item) => item.id === event.target.value);
          if (tenant) {
            selectTenant(tenant);
          }
        }}
      >
        {tenants.map((tenant) => (
          <option key={tenant.id} value={tenant.id}>
            {tenant.name}
          </option>
        ))}
      </select>
    </label>
  );
}
