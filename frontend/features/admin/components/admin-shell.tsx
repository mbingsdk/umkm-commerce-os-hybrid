"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { useAdminMe } from "@/features/admin/hooks/use-admin";
import { useAuthStore } from "@/lib/stores/auth.store";
import { cn } from "@/lib/utils/cn";

const navItems = [
  { label: "Overview", href: "/admin" },
  { label: "Tenant", href: "/admin/tenants" },
  { label: "Paket", href: "/admin/plans" },
  { label: "Featured", href: "/admin/discovery/featured" },
  { label: "Audit log", href: "/admin/audit-logs" }
];

export function AdminShell({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const accessToken = useAuthStore((state) => state.accessToken);

  const adminQuery = useAdminMe(!!accessToken);

  if (!accessToken) {
    return (
      <main className="min-h-screen bg-neutral-50 p-6">
        <div className="mx-auto max-w-xl pt-20">
          <Card>
            <CardContent className="space-y-4">
              <Badge tone="warning">Login diperlukan</Badge>
              <h1 className="text-2xl font-semibold text-neutral-950">Masuk sebagai super admin</h1>
              <p className="text-sm leading-6 text-neutral-500">
                Console admin memakai token login yang sama, tetapi akses tetap divalidasi oleh backend super_admin guard.
              </p>
              <Link href="/login">
                <Button>Ke halaman login</Button>
              </Link>
            </CardContent>
          </Card>
        </div>
      </main>
    );
  }

  if (adminQuery.isPending) {
    return (
      <main className="min-h-screen bg-neutral-50 p-6">
        <LoadingState lines={4} />
      </main>
    );
  }

  if (adminQuery.isError) {
    return (
      <main className="min-h-screen bg-neutral-50 p-6">
        <div className="mx-auto max-w-2xl pt-20">
          <ErrorState
            title="Akses super admin ditolak"
            description="Akun ini bukan platform super_admin. UI guard hanya membantu UX; backend tetap menjadi sumber kebenaran."
            onRetry={() => void adminQuery.refetch()}
          />
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-neutral-50 text-neutral-950">
      <div className="flex min-h-screen">
        <aside className="hidden w-72 border-r border-neutral-200 bg-neutral-950 p-6 text-white lg:block">
          <p className="text-sm font-semibold text-primary-300">UMKM Commerce OS</p>
          <h1 className="mt-2 text-xl font-bold">Super Admin</h1>
          <p className="mt-2 text-xs leading-5 text-neutral-400">{adminQuery.data.email}</p>
          <nav className="mt-8 space-y-2">
            {navItems.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  "block rounded-xl px-3 py-2 text-sm font-semibold transition",
                  pathname === item.href || (item.href !== "/admin" && pathname.startsWith(item.href))
                    ? "bg-white text-neutral-950"
                    : "text-neutral-300 hover:bg-white/10 hover:text-white"
                )}
              >
                {item.label}
              </Link>
            ))}
          </nav>
          <div className="mt-8 rounded-2xl border border-white/10 bg-white/5 p-4 text-xs leading-5 text-neutral-300">
            Admin endpoint tidak memakai <code>X-Tenant-ID</code>. Semua mutasi penting tetap diaudit oleh backend.
          </div>
        </aside>

        <section className="min-w-0 flex-1">
          <div className="border-b border-neutral-200 bg-white px-4 py-3 lg:hidden">
            <div className="flex items-center justify-between gap-3">
              <div>
                <p className="text-xs font-semibold text-primary-700">Super Admin</p>
                <p className="text-sm text-neutral-500">{adminQuery.data.email}</p>
              </div>
              <Badge tone="primary">Admin</Badge>
            </div>
            <nav className="mt-3 flex gap-2 overflow-x-auto pb-1">
              {navItems.map((item) => (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    "shrink-0 rounded-full px-3 py-1.5 text-xs font-semibold",
                    pathname === item.href || (item.href !== "/admin" && pathname.startsWith(item.href))
                      ? "bg-neutral-950 text-white"
                      : "bg-neutral-100 text-neutral-700"
                  )}
                >
                  {item.label}
                </Link>
              ))}
            </nav>
          </div>
          <div className="mx-auto max-w-7xl p-4 sm:p-6 lg:p-8">{children}</div>
        </section>
      </div>
    </main>
  );
}
