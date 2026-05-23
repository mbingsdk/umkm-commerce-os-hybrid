"use client";

import { useState, type FormEvent } from "react";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  AdminPageHeader,
  Field,
  StatCard,
  StatusBadge,
  formatMaybeDate,
  formatNumber
} from "@/features/admin/components/admin-shared";
import { useAdminPlans, useAdminTenantDetail, useAdminTenantMutations } from "@/features/admin/hooks/use-admin";
import { formatDateTime } from "@/lib/format/date";
import { useToastStore } from "@/lib/stores/toast.store";

type DialogMode = "status" | "plan" | null;

export function AdminTenantDetailPage({ tenantId }: { tenantId: string }) {
  const [dialogMode, setDialogMode] = useState<DialogMode>(null);
  const detailQuery = useAdminTenantDetail(tenantId);
  const plansQuery = useAdminPlans();
  const { updateStatus, updatePlan } = useAdminTenantMutations();
  const pushToast = useToastStore((state) => state.pushToast);

  if (detailQuery.isPending) {
    return <LoadingState lines={5} />;
  }

  if (detailQuery.isError) {
    return (
      <ErrorState
        title="Detail tenant gagal dimuat"
        description="Coba muat ulang sebelum melakukan perubahan status atau paket."
        onRetry={() => void detailQuery.refetch()}
      />
    );
  }

  const detail = detailQuery.data;

  function submitStatus(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    updateStatus.mutate(
      {
        tenantId,
        status: String(form.get("status") ?? ""),
        reason: String(form.get("reason") ?? "")
      },
      {
        onSuccess: () => {
          setDialogMode(null);
          pushToast({ tone: "success", title: "Status tenant diperbarui" });
        },
        onError: (error) => pushToast({ tone: "error", title: "Status gagal diperbarui", description: error.message })
      }
    );
  }

  function submitPlan(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    updatePlan.mutate(
      {
        tenantId,
        planId: String(form.get("plan_id") ?? ""),
        reason: String(form.get("reason") ?? "")
      },
      {
        onSuccess: () => {
          setDialogMode(null);
          pushToast({ tone: "success", title: "Paket tenant diperbarui" });
        },
        onError: (error) => pushToast({ tone: "error", title: "Paket gagal diperbarui", description: error.message })
      }
    );
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title={detail.tenant.name}
        description="Safe platform overview untuk tenant. Tidak menampilkan customer/order item/payment proof/cost price."
        action={
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="outline" onClick={() => setDialogMode("plan")}>
              Ubah paket
            </Button>
            <Button type="button" variant={detail.tenant.status === "suspended" ? "primary" : "danger"} onClick={() => setDialogMode("status")}>
              {detail.tenant.status === "suspended" ? "Aktifkan" : "Ubah status"}
            </Button>
          </div>
        }
      />

      <div className="flex flex-wrap gap-2">
        <StatusBadge status={detail.tenant.status} />
        {detail.plan ? <StatusBadge status={detail.plan.isActive ? "plan_active" : "plan_inactive"} /> : null}
      </div>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard label="Produk" value={formatNumber(detail.counts.products)} />
        <StatCard label="Order" value={formatNumber(detail.counts.orders)} />
        <StatCard label="User" value={formatNumber(detail.counts.users)} />
        <StatCard label="POS trx" value={formatNumber(detail.counts.posTransactions)} />
      </div>

      <div className="grid gap-6 xl:grid-cols-[1fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Tenant & store</CardTitle>
            <CardDescription>Informasi dasar yang aman untuk platform admin.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4 text-sm">
            <InfoRow label="Tenant slug" value={detail.tenant.slug} />
            <InfoRow label="Dibuat" value={formatMaybeDate(detail.tenant.createdAt)} />
            <InfoRow label="Diperbarui" value={formatMaybeDate(detail.tenant.updatedAt)} />
            <InfoRow label="Store" value={detail.primaryStore?.name ?? "—"} />
            <InfoRow label="Store slug" value={detail.primaryStore?.slug ?? "—"} />
            <InfoRow label="Kota" value={detail.primaryStore?.city ?? "—"} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Owner & paket</CardTitle>
            <CardDescription>Email owner ditampilkan sebagai safe account overview.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4 text-sm">
            <InfoRow label="Owner" value={detail.owner?.name ?? "—"} />
            <InfoRow label="Owner email" value={detail.owner?.email ?? "—"} />
            <InfoRow label="Status owner" value={detail.owner?.status ?? "—"} />
            <InfoRow label="Paket" value={detail.plan?.name ?? "—"} />
            <InfoRow label="Kode paket" value={detail.plan?.code ?? "—"} />
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Audit terbaru</CardTitle>
          <CardDescription>Cuplikan mutasi admin terkait tenant ini.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {detail.latestAudits.length === 0 ? (
            <EmptyState title="Belum ada audit" description="Perubahan super admin untuk tenant ini akan muncul di sini." />
          ) : (
            detail.latestAudits.map((audit) => (
              <div key={audit.id} className="rounded-xl border border-neutral-200 p-3 text-sm">
                <p className="font-semibold text-neutral-950">{audit.action}</p>
                <p className="mt-1 text-neutral-500">
                  {audit.actorName || audit.actorUserId || "System"} • {formatDateTime(audit.createdAt)}
                </p>
              </div>
            ))
          )}
        </CardContent>
      </Card>

      <Dialog
        open={dialogMode === "status"}
        title="Ubah status tenant"
        description="Suspended tenant tidak bisa akses dashboard dan tidak tampil publik."
        onClose={() => setDialogMode(null)}
        footer={
          <>
            <Button type="button" variant="outline" onClick={() => setDialogMode(null)}>
              Batal
            </Button>
            <Button type="submit" form="tenant-status-form" isLoading={updateStatus.isPending}>
              Simpan status
            </Button>
          </>
        }
      >
        <form id="tenant-status-form" className="space-y-4" onSubmit={submitStatus}>
          <Field label="Status">
            <select name="status" defaultValue={detail.tenant.status} className="h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm">
              <option value="trialing">Trialing</option>
              <option value="active">Active</option>
              <option value="suspended">Suspended</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </Field>
          <Field label="Reason">
            <Input name="reason" placeholder="Contoh: pelanggaran kebijakan platform" />
          </Field>
        </form>
      </Dialog>

      <Dialog
        open={dialogMode === "plan"}
        title="Ubah paket tenant"
        description="Plan mutation akan diaudit oleh backend."
        onClose={() => setDialogMode(null)}
        footer={
          <>
            <Button type="button" variant="outline" onClick={() => setDialogMode(null)}>
              Batal
            </Button>
            <Button type="submit" form="tenant-plan-form" isLoading={updatePlan.isPending}>
              Simpan paket
            </Button>
          </>
        }
      >
        <form id="tenant-plan-form" className="space-y-4" onSubmit={submitPlan}>
          <Field label="Paket">
            <select name="plan_id" defaultValue={detail.plan?.id ?? ""} className="h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm">
              <option value="">Pilih paket</option>
              {(plansQuery.data ?? []).map((plan) => (
                <option key={plan.id} value={plan.id}>
                  {plan.name} ({plan.code})
                </option>
              ))}
            </select>
          </Field>
          <Field label="Reason">
            <Input name="reason" placeholder="Contoh: upgrade manual pilot" />
          </Field>
        </form>
      </Dialog>
    </div>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between gap-4 border-b border-neutral-100 pb-3 last:border-0 last:pb-0">
      <span className="text-neutral-500">{label}</span>
      <span className="text-right font-medium text-neutral-950">{value}</span>
    </div>
  );
}
