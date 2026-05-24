"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm, useWatch } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  stockAdjustmentSchema,
  type StockAdjustmentFormValues
} from "@/features/inventory/schemas/inventory.schema";
import type { AdjustStockInput, InventoryStock } from "@/features/inventory/types";
import { cn } from "@/lib/utils/cn";

type StockAdjustmentDialogProps = {
  open: boolean;
  stock?: InventoryStock | null;
  isSubmitting?: boolean;
  error?: string;
  onClose: () => void;
  onSubmit: (values: AdjustStockInput) => void;
};

export function StockAdjustmentDialog({
  open,
  stock,
  isSubmitting = false,
  error,
  onClose,
  onSubmit
}: StockAdjustmentDialogProps) {
  const form = useForm<StockAdjustmentFormValues>({
    resolver: zodResolver(stockAdjustmentSchema),
    defaultValues: {
      adjustmentType: "in",
      quantity: 1,
      reason: "",
      note: "",
      confirmedStockOut: false
    }
  });
  const adjustmentType = useWatch({ control: form.control, name: "adjustmentType" });
  const quantity = useWatch({ control: form.control, name: "quantity" });
  const confirmedStockOut = useWatch({ control: form.control, name: "confirmedStockOut" });
  const isStockOut = adjustmentType === "out";
  const stockOutNeedsConfirm = isStockOut && !confirmedStockOut;
  const stockOutWouldExceedAvailable = isStockOut && !!stock && quantity > stock.quantityAvailable;
  const disabled = isSubmitting || stockOutNeedsConfirm || stockOutWouldExceedAvailable;

  function handleClose() {
    form.reset();
    onClose();
  }

  function handleSubmit(values: StockAdjustmentFormValues) {
    if (isSubmitting || stockOutNeedsConfirm || stockOutWouldExceedAvailable) {
      return;
    }

    onSubmit({
      adjustmentType: values.adjustmentType,
      quantity: values.quantity,
      reason: values.reason.trim(),
      note: values.note?.trim() || undefined
    });
  }

  return (
    <Dialog
      open={open}
      title="Sesuaikan stok"
      description="Gunakan hanya untuk koreksi manual seperti restock, opname, atau barang rusak. Semua perubahan dicatat sebagai movement."
      onClose={handleClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={handleClose}>
            Batal
          </Button>
          <Button
            type="submit"
            form="stock-adjustment-form"
            variant={isStockOut ? "danger" : "primary"}
            isLoading={isSubmitting}
            disabled={disabled}
          >
            Simpan penyesuaian
          </Button>
        </>
      }
    >
      <form id="stock-adjustment-form" className="space-y-4" onSubmit={form.handleSubmit(handleSubmit)}>
        {stock ? (
          <div className="rounded-xl bg-neutral-50 p-3 text-sm text-neutral-600">
            <p className="font-semibold text-neutral-950">{stock.name}</p>
            <p className="mt-1">
              Tersedia {stock.quantityAvailable} • Fisik {stock.quantityOnHand} • Reserved {stock.quantityReserved}
            </p>
          </div>
        ) : null}

        <label className="block text-sm font-medium text-neutral-800">
          Tipe penyesuaian
          <select
            className="mt-2 h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            {...form.register("adjustmentType", {
              onChange: () => form.setValue("confirmedStockOut", false)
            })}
          >
            <option value="in">Stok masuk</option>
            <option value="out">Stok keluar</option>
          </select>
        </label>

        <label className="block text-sm font-medium text-neutral-800">
          Jumlah
          <Input
            className="mt-2"
            type="number"
            min={1}
            step={1}
            hasError={!!form.formState.errors.quantity || stockOutWouldExceedAvailable}
            {...form.register("quantity", { valueAsNumber: true })}
          />
        </label>
        {form.formState.errors.quantity ? (
          <p className="-mt-2 text-xs font-medium text-red-600">{form.formState.errors.quantity.message}</p>
        ) : null}

        {stockOutWouldExceedAvailable ? (
          <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">
            Jumlah stok keluar tidak boleh melebihi stok tersedia saat ini.
          </p>
        ) : null}

        <label className="block text-sm font-medium text-neutral-800">
          Alasan
          <Input
            className="mt-2"
            hasError={!!form.formState.errors.reason}
            placeholder="Contoh: Restock supplier, stock opname, barang rusak"
            {...form.register("reason")}
          />
        </label>
        {form.formState.errors.reason ? (
          <p className="-mt-2 text-xs font-medium text-red-600">{form.formState.errors.reason.message}</p>
        ) : null}

        <label className="block text-sm font-medium text-neutral-800">
          Catatan
          <textarea
            className={cn(
              "mt-2 min-h-24 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-sm text-neutral-950 shadow-sm outline-none transition placeholder:text-neutral-400 focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            )}
            placeholder="Opsional"
            {...form.register("note")}
          />
        </label>
        {form.formState.errors.note ? (
          <p className="-mt-2 text-xs font-medium text-red-600">{form.formState.errors.note.message}</p>
        ) : null}

        {isStockOut ? (
          <label className="flex items-start gap-3 rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
            <input
              className="mt-1"
              type="checkbox"
              {...form.register("confirmedStockOut")}
            />
            <span>Saya paham stok keluar akan mengurangi stok fisik dan stok tersedia.</span>
          </label>
        ) : null}

        {error ? <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p> : null}
      </form>
    </Dialog>
  );
}
