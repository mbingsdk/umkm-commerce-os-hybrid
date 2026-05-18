"use client";

import { LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TenantSelector } from "@/features/auth/components/tenant-selector";
import type { Store } from "@/features/settings/api/store.api";
import { formatDate } from "@/lib/format/date";
import { useAuthStore } from "@/lib/stores/auth.store";

type DashboardTopbarProps = {
  currentStore?: Store;
  onLogout: () => void;
  isLoggingOut?: boolean;
};

export function DashboardTopbar({ currentStore, onLogout, isLoggingOut = false }: DashboardTopbarProps) {
  const user = useAuthStore((state) => state.user);

  return (
    <header className="border-b border-neutral-200 bg-white px-4 py-4 sm:px-6 lg:px-8">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">{formatDate(new Date())}</p>
          <p className="mt-1 text-sm text-neutral-600">
            {currentStore ? `Toko aktif: ${currentStore.name}` : "Pilih tenant untuk mulai bekerja."}
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          <TenantSelector />
          <div className="hidden text-right sm:block">
            <p className="text-sm font-semibold text-neutral-900">{user?.name ?? "Pengguna"}</p>
            <p className="text-xs text-neutral-500">{user?.email ?? "Belum termuat"}</p>
          </div>
          <Button variant="outline" size="sm" onClick={onLogout} isLoading={isLoggingOut}>
            <LogOut className="h-4 w-4" />
            Keluar
          </Button>
        </div>
      </div>
    </header>
  );
}
