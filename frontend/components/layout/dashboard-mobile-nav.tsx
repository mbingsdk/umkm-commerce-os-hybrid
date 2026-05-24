"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { dashboardNavItems, isDashboardNavItemActive } from "@/components/layout/dashboard-nav";
import { PermissionGate } from "@/lib/permissions/permission-gate";
import { cn } from "@/lib/utils/cn";

export function DashboardMobileNav() {
  const pathname = usePathname();

  return (
    <nav
      aria-label="Navigasi dashboard"
      className="border-b border-neutral-200 bg-white px-4 py-3 lg:hidden"
    >
      <div className="flex gap-2 overflow-x-auto pb-1">
        {dashboardNavItems.map((item) => (
          <PermissionGate key={item.label} permission={item.permission}>
            {item.ready && item.href ? (
              <Link
                href={item.href}
                className={cn(
                  "shrink-0 rounded-full px-3 py-2 text-xs font-semibold transition",
                  isDashboardNavItemActive(pathname, item.href)
                    ? "bg-primary-600 text-white"
                    : "bg-neutral-100 text-neutral-700 hover:bg-neutral-200"
                )}
              >
                {item.label}
              </Link>
            ) : (
              <span
                className="shrink-0 rounded-full bg-neutral-50 px-3 py-2 text-xs font-semibold text-neutral-400"
                aria-disabled="true"
              >
                {item.label} segera
              </span>
            )}
          </PermissionGate>
        ))}
      </div>
    </nav>
  );
}
