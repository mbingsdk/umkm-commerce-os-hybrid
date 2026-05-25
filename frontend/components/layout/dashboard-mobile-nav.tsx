"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { dashboardNavItems, isDashboardNavItemActive } from "@/components/layout/dashboard-nav";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { useTenantStore } from "@/lib/stores/tenant.store";
import { cn } from "@/lib/utils/cn";

export function DashboardMobileNav() {
  const pathname = usePathname();
  const [menuOpen, setMenuOpen] = useState(false);
  const userPermissions = useTenantStore((state) => state.permissions);
  const visibleItems = dashboardNavItems.filter((item) => item.ready && item.href && userPermissions.includes(item.permission));
  const activeItem = visibleItems.find((item) => item.href && isDashboardNavItemActive(pathname, item.href));

  return (
    <nav
      aria-label="Navigasi dashboard"
      className="sticky top-0 z-50 border-b border-neutral-200 bg-white/95 px-4 py-3 shadow-sm backdrop-blur lg:hidden"
    >
      <div className="flex items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">Menu dashboard</p>
          <p className="truncate text-sm font-semibold text-neutral-950">{activeItem?.label ?? "Pilih area kerja"}</p>
        </div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          aria-expanded={menuOpen}
          onClick={() => setMenuOpen(true)}
        >
          Menu
        </Button>
      </div>

      <Dialog
        open={menuOpen}
        title="Menu dashboard"
        description="Pilih area kerja toko. Menu yang tidak sesuai permission tidak ditampilkan."
        onClose={() => setMenuOpen(false)}
      >
        <div className="grid gap-2">
          {visibleItems.map((item) =>
            item.href ? (
              <Link
                key={item.label}
                href={item.href}
                onClick={() => setMenuOpen(false)}
                className={cn(
                  "rounded-2xl border px-4 py-3 text-sm font-semibold transition",
                  isDashboardNavItemActive(pathname, item.href)
                    ? "border-primary-200 bg-primary-50 text-primary-700"
                    : "border-neutral-200 bg-white text-neutral-800 hover:bg-neutral-50"
                )}
              >
                {item.label}
              </Link>
            ) : null
          )}
        </div>
      </Dialog>
    </nav>
  );
}
