"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm } from "react-hook-form";
import { useEffect } from "react";
import { MoneyInput } from "@/components/forms/money-input";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import type { Expense } from "@/features/finance/types";
import {
  expenseCategories,
  expenseSchema,
  type ExpenseFormValues
} from "@/features/finance/schemas/expense.schema";

type ExpenseFormDialogProps = {
  open: boolean;
  expense?: Expense | null;
  isSubmitting?: boolean;
  error?: string;
  onClose: () => void;
  onSubmit: (values: ExpenseFormValues) => void;
};

export function ExpenseFormDialog({
  open,
  expense,
  isSubmitting = false,
  error,
  onClose,
  onSubmit
}: ExpenseFormDialogProps) {
  const form = useForm<ExpenseFormValues>({
    resolver: zodResolver(expenseSchema),
    defaultValues: toDefaults(expense)
  });

  useEffect(() => {
    form.reset(toDefaults(expense));
  }, [expense, form, open]);

  function handleSubmit(values: ExpenseFormValues) {
    if (isSubmitting) {
      return;
    }

    onSubmit(values);
  }

  return (
    <Dialog
      open={open}
      title={expense ? "Edit pengeluaran" : "Tambah pengeluaran"}
      description="Catat biaya operasional sederhana. Ini belum menjadi ledger akuntansi penuh."
      onClose={onClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={onClose}>
            Batal
          </Button>
          <Button type="submit" form="expense-form" isLoading={isSubmitting} disabled={isSubmitting}>
            {expense ? "Simpan perubahan" : "Tambah pengeluaran"}
          </Button>
        </>
      }
    >
      <form id="expense-form" className="space-y-4" onSubmit={form.handleSubmit(handleSubmit)}>
        <Field label="Judul" error={form.formState.errors.title?.message}>
          <Input
            placeholder="Beli plastik bouquet"
            hasError={!!form.formState.errors.title}
            {...form.register("title")}
          />
        </Field>

        <Field label="Kategori" error={form.formState.errors.category?.message}>
          <select
            className="h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            {...form.register("category")}
          >
            <option value="">Tanpa kategori</option>
            {expenseCategories.map((category) => (
              <option key={category.value} value={category.value}>
                {category.label}
              </option>
            ))}
          </select>
        </Field>

        <Field label="Nominal" error={form.formState.errors.amount?.message}>
          <Controller
            control={form.control}
            name="amount"
            render={({ field }) => (
              <MoneyInput
                value={field.value ?? null}
                onChange={(value) => field.onChange(value ?? 0)}
                hasError={!!form.formState.errors.amount}
              />
            )}
          />
        </Field>

        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Tanggal" error={form.formState.errors.expenseDate?.message}>
            <Input type="date" hasError={!!form.formState.errors.expenseDate} {...form.register("expenseDate")} />
          </Field>
          <Field label="Metode bayar" error={form.formState.errors.paymentMethod?.message}>
            <select
              className="h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
              {...form.register("paymentMethod")}
            >
              <option value="">Tidak dicatat</option>
              <option value="cash">Cash</option>
              <option value="bank_transfer">Transfer bank</option>
              <option value="qris_manual">QRIS manual</option>
              <option value="other">Lainnya</option>
            </select>
          </Field>
        </div>

        <Field label="Catatan" error={form.formState.errors.note?.message}>
          <textarea
            className="min-h-24 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-sm text-neutral-950 outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            placeholder="Opsional, jangan isi data sensitif."
            {...form.register("note")}
          />
        </Field>

        {error ? <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p> : null}
      </form>
    </Dialog>
  );
}

function Field({
  label,
  children,
  error
}: {
  label: string;
  children: React.ReactNode;
  error?: string;
}) {
  return (
    <label className="block space-y-2 text-sm font-medium text-neutral-700">
      {label}
      {children}
      {error ? <span className="block text-sm font-normal text-red-600">{error}</span> : null}
    </label>
  );
}

function toDefaults(expense?: Expense | null): ExpenseFormValues {
  return {
    category: expense?.category ?? "",
    title: expense?.title ?? "",
    amount: expense?.amount ?? 0,
    expenseDate: expense?.expenseDate ?? new Date().toISOString().slice(0, 10),
    paymentMethod: expense?.paymentMethod ?? "",
    note: expense?.note ?? ""
  };
}
