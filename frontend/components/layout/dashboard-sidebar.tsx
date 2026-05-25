"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { dashboardNavItems, isDashboardNavItemActive } from "@/components/layout/dashboard-nav";
import { PermissionGate } from "@/lib/permissions/permission-gate";
import { cn } from "@/lib/utils/cn";

export function DashboardSidebar() {
  const pathname = usePathname();

  return (
    <aside className="sticky top-0 hidden h-screen w-64 shrink-0 overflow-y-auto border-r border-neutral-200 bg-white p-5 lg:block">
      <p className="text-sm font-semibold text-primary-700">UMKM Commerce OS</p>
      <nav className="mt-8 space-y-2 text-sm text-neutral-600">
        {dashboardNavItems.map((item) => {
          if (!item.ready || !item.href) {
            return null;
          }

          return (
            <PermissionGate key={item.label} permission={item.permission}>
              <Link
                className={cn(
                  "block rounded-xl px-3 py-2 font-medium transition hover:bg-neutral-100",
                  isDashboardNavItemActive(pathname, item.href) && "bg-primary-50 text-primary-700"
                )}
                href={item.href}
              >
                {item.label}
              </Link>
            </PermissionGate>
          );
        })}
      </nav>
    </aside>
  );
}
