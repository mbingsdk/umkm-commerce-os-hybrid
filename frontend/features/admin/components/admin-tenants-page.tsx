"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { AdminPageHeader, StatusBadge, formatNumber } from "@/features/admin/components/admin-shared";
import { useAdminPlans, useAdminTenants } from "@/features/admin/hooks/use-admin";
import type { AdminTenantListItem } from "@/features/admin/types";
import { formatDateTime } from "@/lib/format/date";

export function AdminTenantsPage() {
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState("");
  const [planId, setPlanId] = useState("");
  const [cursor, setCursor] = useState<string | undefined>();
  const filters = useMemo(
    () => ({
      query,
      status,
      planId,
      cursor,
      limit: 20
    }),
    [cursor, planId, query, status]
  );
  const tenantsQuery = useAdminTenants(filters);
  const plansQuery = useAdminPlans();

  const columns: Array<DataTableColumn<AdminTenantListItem>> = [
    {
      key: "tenant",
      header: "Tenant",
      render: (tenant) => (
        <div>
          <Link className="font-semibold text-primary-700 hover:text-primary-800" href={`/admin/tenants/${tenant.id}`}>
            {tenant.name}
          </Link>
          <p className="mt-1 text-xs text-neutral-500">{tenant.slug}</p>
        </div>
      )
    },
    {
      key: "store",
      header: "Store / owner",
      render: (tenant) => (
        <div>
          <p className="font-medium text-neutral-900">{tenant.primaryStore?.name ?? "—"}</p>
          <p className="mt-1 text-xs text-neutral-500">{tenant.owner?.email ?? "Owner belum tersedia"}</p>
        </div>
      )
    },
    { key: "status", header: "Status", render: (tenant) => <StatusBadge status={tenant.status} /> },
    { key: "plan", header: "Paket", render: (tenant) => tenant.plan?.name ?? "—" },
    {
      key: "counts",
      header: "Counts",
      render: (tenant) => (
        <p className="text-xs leading-5 text-neutral-500">
          Produk {formatNumber(tenant.counts.products)} • Order {formatNumber(tenant.counts.orders)} • User{" "}
          {formatNumber(tenant.counts.users)}
        </p>
      )
    },
    { key: "created", header: "Dibuat", render: (tenant) => formatDateTime(tenant.createdAt) }
  ];

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Tenant"
        description="Kelola tenant platform, status subscription, dan paket. Detail sensitif customer/order tidak ditampilkan di console admin."
      />

      <Card>
        <CardHeader>
          <CardTitle>Filter tenant</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-3 md:grid-cols-[1fr_180px_220px_auto]">
          <Input
            placeholder="Cari nama tenant, toko, atau owner"
            value={query}
            onChange={(event) => {
              setCursor(undefined);
              setQuery(event.target.value);
            }}
          />
          <select
            className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm"
            value={status}
            onChange={(event) => {
              setCursor(undefined);
              setStatus(event.target.value);
            }}
          >
            <option value="">Semua status</option>
            <option value="trialing">Trialing</option>
            <option value="active">Active</option>
            <option value="suspended">Suspended</option>
            <option value="cancelled">Cancelled</option>
          </select>
          <select
            className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm"
            value={planId}
            onChange={(event) => {
              setCursor(undefined);
              setPlanId(event.target.value);
            }}
          >
            <option value="">Semua paket</option>
            {(plansQuery.data ?? []).map((plan) => (
              <option key={plan.id} value={plan.id}>
                {plan.name}
              </option>
            ))}
          </select>
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              setQuery("");
              setStatus("");
              setPlanId("");
              setCursor(undefined);
            }}
          >
            Reset
          </Button>
        </CardContent>
      </Card>

      {tenantsQuery.isPending ? (
        <LoadingState lines={4} />
      ) : tenantsQuery.isError ? (
        <ErrorState
          title="Tenant gagal dimuat"
          description="Coba muat ulang daftar tenant."
          onRetry={() => void tenantsQuery.refetch()}
        />
      ) : tenantsQuery.data.items.length === 0 ? (
        <EmptyState title="Tenant tidak ditemukan" description="Coba ubah kata kunci atau filter status/paket." />
      ) : (
        <>
          <DataTable columns={columns} rows={tenantsQuery.data.items} getRowKey={(tenant) => tenant.id} />
          <div className="flex justify-end">
            <Button
              type="button"
              variant="outline"
              disabled={!tenantsQuery.data.pagination.hasMore}
              onClick={() => setCursor(tenantsQuery.data.pagination.nextCursor ?? undefined)}
            >
              Muat berikutnya
            </Button>
          </div>
        </>
      )}
    </div>
  );
}
