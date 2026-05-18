"use client";

import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { PermissionGate } from "@/lib/permissions/permission-gate";
import { permissions, type Permission } from "@/lib/permissions/permissions";

const navItems: Array<{
  label: string;
  href?: string;
  permission: Permission;
  ready: boolean;
}> = [
  { label: "Dashboard", href: "/dashboard", permission: permissions.tenantRead, ready: true },
  { label: "Toko", permission: permissions.storeRead, ready: false },
  { label: "Produk", permission: permissions.productRead, ready: false },
  { label: "Inventori", permission: permissions.inventoryRead, ready: false },
  { label: "Pesanan", permission: permissions.orderRead, ready: false },
  { label: "POS", permission: permissions.posCreateTransaction, ready: false },
  { label: "Keuangan", permission: permissions.financeReadSummary, ready: false }
];

export function DashboardSidebar() {
  return (
    <aside className="hidden w-64 border-r border-neutral-200 bg-white p-5 lg:block">
      <p className="text-sm font-semibold text-primary-700">UMKM Commerce OS</p>
      <Badge className="mt-3" tone="primary">
        Sprint 2C
      </Badge>
      <nav className="mt-8 space-y-2 text-sm text-neutral-600">
        {navItems.map((item) => (
          <PermissionGate key={item.label} permission={item.permission}>
            {item.ready && item.href ? (
              <Link className="block rounded-xl px-3 py-2 font-medium hover:bg-neutral-100" href={item.href}>
                {item.label}
              </Link>
            ) : (
              <div className="flex items-center justify-between rounded-xl px-3 py-2 text-neutral-400" aria-disabled="true">
                <span>{item.label}</span>
                <span className="text-xs">Segera</span>
              </div>
            )}
          </PermissionGate>
        ))}
      </nav>
    </aside>
  );
}
