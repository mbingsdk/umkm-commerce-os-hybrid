"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useState } from "react";
import { useForm, type UseFormRegisterReturn } from "react-hook-form";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { AdminPageHeader, Field, StatusBadge, formatNumber } from "@/features/admin/components/admin-shared";
import { useAdminPlanMutations, useAdminPlans } from "@/features/admin/hooks/use-admin";
import { adminPlanSchema, type AdminPlanFormValues } from "@/features/admin/schemas/admin-plan.schema";
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
  const form = useForm<AdminPlanFormValues>({
    resolver: zodResolver(adminPlanSchema),
    defaultValues: toPlanFormValues(plan)
  });

  useEffect(() => {
    if (open) {
      form.reset(toPlanFormValues(plan));
    }
  }, [form, open, plan]);

  function handleClose() {
    form.reset(toPlanFormValues(plan));
    onClose();
  }

  function handleSubmit(values: AdminPlanFormValues) {
    if (isSubmitting) {
      return;
    }

    onSubmit({
      code: values.code.trim(),
      name: values.name.trim(),
      description: values.description?.trim() || "",
      priceMonthly: values.priceMonthly,
      productLimit: nullableNumber(values.productLimit),
      staffLimit: nullableNumber(values.staffLimit),
      canUsePos: values.canUsePos,
      canUseDiscovery: values.canUseDiscovery,
      canUseCourier: values.canUseCourier,
      isActive: values.isActive
    });
  }

  return (
    <Dialog
      open={open}
      title={plan ? "Edit paket" : "Buat paket"}
      description="Limit kosong berarti unlimited untuk MVP."
      onClose={handleClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={handleClose}>
            Batal
          </Button>
          <Button type="submit" form="admin-plan-form" isLoading={isSubmitting} disabled={isSubmitting}>
            Simpan
          </Button>
        </>
      }
    >
      <form key={plan?.id ?? "new"} id="admin-plan-form" className="space-y-4" onSubmit={form.handleSubmit(handleSubmit)}>
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Kode">
            <Input hasError={!!form.formState.errors.code} placeholder="starter" {...form.register("code")} />
            {form.formState.errors.code ? (
              <span className="block text-xs font-medium text-red-600">{form.formState.errors.code.message}</span>
            ) : null}
          </Field>
          <Field label="Nama">
            <Input hasError={!!form.formState.errors.name} placeholder="Starter" {...form.register("name")} />
            {form.formState.errors.name ? (
              <span className="block text-xs font-medium text-red-600">{form.formState.errors.name.message}</span>
            ) : null}
          </Field>
        </div>
        <Field label="Deskripsi">
          <Input
            hasError={!!form.formState.errors.description}
            placeholder="Untuk UMKM awal"
            {...form.register("description")}
          />
          {form.formState.errors.description ? (
            <span className="block text-xs font-medium text-red-600">{form.formState.errors.description.message}</span>
          ) : null}
        </Field>
        <div className="grid gap-4 sm:grid-cols-3">
          <Field label="Harga bulanan">
            <Input
              hasError={!!form.formState.errors.priceMonthly}
              type="number"
              min={0}
              {...form.register("priceMonthly", { valueAsNumber: true })}
            />
            {form.formState.errors.priceMonthly ? (
              <span className="block text-xs font-medium text-red-600">{form.formState.errors.priceMonthly.message}</span>
            ) : null}
          </Field>
          <Field label="Limit produk">
            <Input
              hasError={!!form.formState.errors.productLimit}
              type="number"
              min={0}
              placeholder="Unlimited"
              {...form.register("productLimit")}
            />
            {form.formState.errors.productLimit ? (
              <span className="block text-xs font-medium text-red-600">{form.formState.errors.productLimit.message}</span>
            ) : null}
          </Field>
          <Field label="Limit staff">
            <Input
              hasError={!!form.formState.errors.staffLimit}
              type="number"
              min={0}
              placeholder="Unlimited"
              {...form.register("staffLimit")}
            />
            {form.formState.errors.staffLimit ? (
              <span className="block text-xs font-medium text-red-600">{form.formState.errors.staffLimit.message}</span>
            ) : null}
          </Field>
        </div>
        <div className="grid gap-3 sm:grid-cols-2">
          <Check label="Bisa pakai POS" registration={form.register("canUsePos")} />
          <Check label="Bisa discovery" registration={form.register("canUseDiscovery")} />
          <Check label="Bisa courier" registration={form.register("canUseCourier")} />
          <Check label="Paket aktif" registration={form.register("isActive")} />
        </div>
      </form>
    </Dialog>
  );
}

function Check({
  label,
  registration
}: {
  label: string;
  registration: UseFormRegisterReturn;
}) {
  return (
    <label className="flex items-center gap-2 rounded-xl border border-neutral-200 p-3 text-sm text-neutral-700">
      <input type="checkbox" {...registration} />
      {label}
    </label>
  );
}

function nullableNumber(value: string) {
  const raw = value.trim();
  if (!raw) {
    return null;
  }
  const parsed = Number(raw);
  return Number.isFinite(parsed) ? parsed : null;
}

function toPlanFormValues(plan: AdminPlan | null): AdminPlanFormValues {
  return {
    code: plan?.code ?? "",
    name: plan?.name ?? "",
    description: plan?.description ?? "",
    priceMonthly: plan?.priceMonthly ?? 0,
    productLimit: plan?.productLimit == null ? "" : String(plan.productLimit),
    staffLimit: plan?.staffLimit == null ? "" : String(plan.staffLimit),
    canUsePos: plan?.canUsePos ?? true,
    canUseDiscovery: plan?.canUseDiscovery ?? true,
    canUseCourier: plan?.canUseCourier ?? false,
    isActive: plan?.isActive ?? true
  };
}

function limitLabel(value: number | null) {
  return value == null ? "Unlimited" : formatNumber(value);
}

function yesNo(value: boolean) {
  return value ? "ya" : "tidak";
}
