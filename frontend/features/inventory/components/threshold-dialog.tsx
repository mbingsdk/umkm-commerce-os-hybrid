"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { thresholdSchema, type ThresholdFormValues } from "@/features/inventory/schemas/inventory.schema";
import type { InventoryStock } from "@/features/inventory/types";

type ThresholdDialogProps = {
  open: boolean;
  stock?: InventoryStock | null;
  isSubmitting?: boolean;
  error?: string;
  onClose: () => void;
  onSubmit: (value: number) => void;
};

export function ThresholdDialog({
  open,
  stock,
  isSubmitting = false,
  error,
  onClose,
  onSubmit
}: ThresholdDialogProps) {
  const form = useForm<ThresholdFormValues>({
    resolver: zodResolver(thresholdSchema),
    defaultValues: {
      lowStockThreshold: stock?.lowStockThreshold ?? 0
    }
  });

  useEffect(() => {
    if (open) {
      form.reset({ lowStockThreshold: stock?.lowStockThreshold ?? 0 });
    }
  }, [form, open, stock?.lowStockThreshold]);

  function handleSubmit(values: ThresholdFormValues) {
    if (isSubmitting) {
      return;
    }

    onSubmit(values.lowStockThreshold);
  }

  return (
    <Dialog
      open={open}
      title="Ubah batas stok menipis"
      description="Produk akan ditandai stok menipis saat stok tersedia kurang dari atau sama dengan batas ini."
      onClose={onClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={onClose}>
            Batal
          </Button>
          <Button type="submit" form="threshold-form" isLoading={isSubmitting} disabled={isSubmitting}>
            Simpan threshold
          </Button>
        </>
      }
    >
      <form id="threshold-form" className="space-y-4" onSubmit={form.handleSubmit(handleSubmit)}>
        {stock ? (
          <div className="rounded-xl bg-neutral-50 p-3 text-sm">
            <p className="font-semibold text-neutral-950">{stock.name}</p>
            <p className="mt-1 text-neutral-500">Threshold saat ini: {stock.lowStockThreshold}</p>
          </div>
        ) : null}

        <label className="block text-sm font-medium text-neutral-800">
          Batas stok menipis
          <Input
            className="mt-2"
            type="number"
            min={0}
            step={1}
            hasError={!!form.formState.errors.lowStockThreshold}
            {...form.register("lowStockThreshold", { valueAsNumber: true })}
          />
        </label>
        {form.formState.errors.lowStockThreshold ? (
          <p className="-mt-2 text-xs font-medium text-red-600">{form.formState.errors.lowStockThreshold.message}</p>
        ) : null}

        {error ? <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p> : null}
      </form>
    </Dialog>
  );
}
