"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { dashboardNavItems, isDashboardNavItemActive } from "@/components/layout/dashboard-nav";
import { PermissionGate } from "@/lib/permissions/permission-gate";
import { cn } from "@/lib/utils/cn";

export function DashboardSidebar() {
  const pathname = usePathname();

  return (
    <aside className="hidden w-64 border-r border-neutral-200 bg-white p-5 lg:block">
      <p className="text-sm font-semibold text-primary-700">UMKM Commerce OS</p>
      <nav className="mt-8 space-y-2 text-sm text-neutral-600">
        {dashboardNavItems.map((item) => (
          <PermissionGate key={item.label} permission={item.permission}>
            {item.ready && item.href ? (
              <Link
                className={cn(
                  "block rounded-xl px-3 py-2 font-medium transition hover:bg-neutral-100",
                  isDashboardNavItemActive(pathname, item.href) && "bg-primary-50 text-primary-700"
                )}
                href={item.href}
              >
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
