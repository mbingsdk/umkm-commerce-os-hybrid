"use client";

import Link from "next/link";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { AdminPageHeader, StatCard, StatusBadge, formatNumber } from "@/features/admin/components/admin-shared";
import { useAdminPlans, useAdminTenants } from "@/features/admin/hooks/use-admin";
import { formatDateTime } from "@/lib/format/date";

export function AdminDashboard() {
  const tenantsQuery = useAdminTenants({ limit: 100 });
  const plansQuery = useAdminPlans();

  if (tenantsQuery.isPending || plansQuery.isPending) {
    return <LoadingState lines={5} />;
  }

  if (tenantsQuery.isError || plansQuery.isError) {
    return (
      <ErrorState
        title="Admin dashboard belum bisa dimuat"
        description="Coba muat ulang ringkasan platform."
        onRetry={() => {
          void tenantsQuery.refetch();
          void plansQuery.refetch();
        }}
      />
    );
  }

  const tenants = tenantsQuery.data.items;
  const activeCount = tenants.filter((tenant) => tenant.status === "active" || tenant.status === "trialing").length;
  const suspendedCount = tenants.filter((tenant) => tenant.status === "suspended").length;
  const latestTenants = tenants.slice(0, 5);

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Admin overview"
        description="Ringkasan ringan untuk memantau tenant, paket, dan area operasional platform."
      />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Tenant dimuat" value={formatNumber(tenants.length)} helper="Mengikuti batas data dari API admin." />
        <StatCard label="Active / trialing" value={formatNumber(activeCount)} helper="Boleh akses dashboard tenant." />
        <StatCard label="Suspended" value={formatNumber(suspendedCount)} helper="Tidak tampil publik." />
        <StatCard label="Paket aktif" value={formatNumber(plansQuery.data.filter((plan) => plan.isActive).length)} helper="Dikelola super admin." />
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
        <Card>
          <CardHeader>
            <CardTitle>Tenant signup terbaru</CardTitle>
            <CardDescription>Tenant terbaru berdasarkan daftar admin.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {latestTenants.length === 0 ? (
              <EmptyState title="Belum ada tenant" description="Tenant baru akan muncul di sini setelah onboarding." />
            ) : (
              latestTenants.map((tenant) => (
                <Link
                  key={tenant.id}
                  href={`/admin/tenants/${tenant.id}`}
                  className="block rounded-2xl border border-neutral-200 p-4 transition hover:bg-neutral-50"
                >
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                    <div>
                      <p className="font-semibold text-neutral-950">{tenant.name}</p>
                      <p className="mt-1 text-sm text-neutral-500">
                        {tenant.primaryStore?.name ?? "Store belum tersedia"} • {tenant.owner?.email ?? "owner belum ada"}
                      </p>
                      <p className="mt-1 text-xs text-neutral-400">{formatDateTime(tenant.createdAt)}</p>
                    </div>
                    <StatusBadge status={tenant.status} />
                  </div>
                </Link>
              ))
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Quick links</CardTitle>
            <CardDescription>Aksi admin yang paling sering dipakai.</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-3">
            <LinkButton href="/admin/tenants">Kelola tenant</LinkButton>
            <LinkButton href="/admin/plans">Kelola paket</LinkButton>
            <LinkButton href="/admin/discovery/featured">Atur featured discovery</LinkButton>
            <LinkButton href="/admin/audit-logs">Buka audit log</LinkButton>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function LinkButton({ href, children }: { href: string; children: string }) {
  return (
    <Link href={href}>
      <Button className="w-full justify-start" variant="outline">
        {children}
      </Button>
    </Link>
  );
}
