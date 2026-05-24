"use client";

import { useState, type FormEvent } from "react";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { AdminPageHeader, Field, StatusBadge, formatNumber } from "@/features/admin/components/admin-shared";
import { useAdminPlanMutations, useAdminPlans } from "@/features/admin/hooks/use-admin";
import type { AdminPlan, PlanFormInput } from "@/features/admin/types";
import { formatRupiah } from "@/lib/format/money";
import { useToastStore } from "@/lib/stores/toast.store";

export function AdminPlansPage() {
  const [editingPlan, setEditingPlan] = useState<AdminPlan | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const plansQuery = useAdminPlans();
  const { createPlan, updatePlan } = useAdminPlanMutations();
  const pushToast = useToastStore((state) => state.pushToast);

  const columns: Array<DataTableColumn<AdminPlan>> = [
    {
      key: "plan",
      header: "Paket",
      render: (plan) => (
        <div>
          <p className="font-semibold text-neutral-950">{plan.name}</p>
          <p className="mt-1 text-xs text-neutral-500">{plan.code}</p>
        </div>
      )
    },
    { key: "price", header: "Harga", render: (plan) => formatRupiah(plan.priceMonthly) },
    {
      key: "limits",
      header: "Limit",
      render: (plan) => (
        <p className="text-xs leading-5 text-neutral-500">
          Produk {limitLabel(plan.productLimit)} • Staff {limitLabel(plan.staffLimit)}
        </p>
      )
    },
    {
      key: "features",
      header: "Fitur",
      render: (plan) => (
        <p className="text-xs leading-5 text-neutral-500">
          POS {yesNo(plan.canUsePos)} • Discovery {yesNo(plan.canUseDiscovery)} • Courier {yesNo(plan.canUseCourier)}
        </p>
      )
    },
    { key: "status", header: "Status", render: (plan) => <StatusBadge status={plan.isActive ? "active" : "inactive"} /> },
    {
      key: "action",
      header: "",
      render: (plan) => (
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => {
            setEditingPlan(plan);
            setDialogOpen(true);
          }}
        >
          Edit
        </Button>
      )
    }
  ];

  function submitPlan(values: PlanFormInput) {
    if (createPlan.isPending || updatePlan.isPending) {
      return;
    }

    if (editingPlan) {
      updatePlan.mutate(
        { planId: editingPlan.id, input: values },
        {
          onSuccess: () => {
            setDialogOpen(false);
            setEditingPlan(null);
            pushToast({ tone: "success", title: "Paket diperbarui" });
          },
          onError: (error) => pushToast({ tone: "error", title: "Paket gagal disimpan", description: error.message })
        }
      );
      return;
    }

    createPlan.mutate(values, {
      onSuccess: () => {
        setDialogOpen(false);
        pushToast({ tone: "success", title: "Paket dibuat" });
      },
      onError: (error) => pushToast({ tone: "error", title: "Paket gagal disimpan", description: error.message })
    });
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Paket"
        description="Kelola paket dasar, feature flag, dan limit produk/staff. Belum termasuk billing/payment gateway."
        action={
          <Button
            type="button"
            onClick={() => {
              setEditingPlan(null);
              setDialogOpen(true);
            }}
          >
            Buat paket
          </Button>
        }
      />

      {plansQuery.isPending ? (
        <LoadingState lines={4} />
      ) : plansQuery.isError ? (
        <ErrorState
          title="Paket gagal dimuat"
          description="Coba muat ulang daftar paket."
          onRetry={() => void plansQuery.refetch()}
        />
      ) : plansQuery.data.length === 0 ? (
        <EmptyState title="Belum ada paket" description="Buat paket pertama untuk mengaktifkan limit dan feature gate." />
      ) : (
        <DataTable columns={columns} rows={plansQuery.data} getRowKey={(plan) => plan.id} />
      )}

      <PlanDialog
        open={dialogOpen}
        plan={editingPlan}
        isSubmitting={createPlan.isPending || updatePlan.isPending}
        onClose={() => {
          setDialogOpen(false);
          setEditingPlan(null);
        }}
        onSubmit={submitPlan}
      />
    </div>
  );
}

function PlanDialog({
  open,
  plan,
  isSubmitting,
  onClose,
  onSubmit
}: {
  open: boolean;
  plan: AdminPlan | null;
  isSubmitting: boolean;
  onClose: () => void;
  onSubmit: (values: PlanFormInput) => void;
}) {
  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (isSubmitting) {
      return;
    }

    const form = new FormData(event.currentTarget);
    onSubmit({
      code: String(form.get("code") ?? "").trim(),
      name: String(form.get("name") ?? "").trim(),
      description: String(form.get("description") ?? "").trim(),
      priceMonthly: Number(form.get("price_monthly") ?? 0),
      productLimit: nullableNumber(form.get("product_limit")),
      staffLimit: nullableNumber(form.get("staff_limit")),
      canUsePos: form.get("can_use_pos") === "on",
      canUseDiscovery: form.get("can_use_discovery") === "on",
      canUseCourier: form.get("can_use_courier") === "on",
      isActive: form.get("is_active") === "on"
    });
  }

  return (
    <Dialog
      open={open}
      title={plan ? "Edit paket" : "Buat paket"}
      description="Limit kosong berarti unlimited untuk MVP."
      onClose={onClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={onClose}>
            Batal
          </Button>
          <Button type="submit" form="admin-plan-form" isLoading={isSubmitting} disabled={isSubmitting}>
            Simpan
          </Button>
        </>
      }
    >
      <form key={plan?.id ?? "new"} id="admin-plan-form" className="space-y-4" onSubmit={handleSubmit}>
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Kode">
            <Input name="code" defaultValue={plan?.code ?? ""} placeholder="starter" required />
          </Field>
          <Field label="Nama">
            <Input name="name" defaultValue={plan?.name ?? ""} placeholder="Starter" required />
          </Field>
        </div>
        <Field label="Deskripsi">
          <Input name="description" defaultValue={plan?.description ?? ""} placeholder="Untuk UMKM awal" />
        </Field>
        <div className="grid gap-4 sm:grid-cols-3">
          <Field label="Harga bulanan">
            <Input name="price_monthly" type="number" min={0} defaultValue={plan?.priceMonthly ?? 0} />
          </Field>
          <Field label="Limit produk">
            <Input name="product_limit" type="number" min={0} defaultValue={plan?.productLimit ?? ""} placeholder="Unlimited" />
          </Field>
          <Field label="Limit staff">
            <Input name="staff_limit" type="number" min={0} defaultValue={plan?.staffLimit ?? ""} placeholder="Unlimited" />
          </Field>
        </div>
        <div className="grid gap-3 sm:grid-cols-2">
          <Check name="can_use_pos" label="Bisa pakai POS" defaultChecked={plan?.canUsePos ?? true} />
          <Check name="can_use_discovery" label="Bisa discovery" defaultChecked={plan?.canUseDiscovery ?? true} />
          <Check name="can_use_courier" label="Bisa courier" defaultChecked={plan?.canUseCourier ?? false} />
          <Check name="is_active" label="Paket aktif" defaultChecked={plan?.isActive ?? true} />
        </div>
      </form>
    </Dialog>
  );
}

function Check({ name, label, defaultChecked }: { name: string; label: string; defaultChecked: boolean }) {
  return (
    <label className="flex items-center gap-2 rounded-xl border border-neutral-200 p-3 text-sm text-neutral-700">
      <input name={name} type="checkbox" defaultChecked={defaultChecked} />
      {label}
    </label>
  );
}

function nullableNumber(value: FormDataEntryValue | null) {
  const raw = String(value ?? "").trim();
  if (!raw) {
    return null;
  }
  const parsed = Number(raw);
  return Number.isFinite(parsed) ? parsed : null;
}

function limitLabel(value: number | null) {
  return value == null ? "Unlimited" : formatNumber(value);
}

function yesNo(value: boolean) {
  return value ? "ya" : "tidak";
}
