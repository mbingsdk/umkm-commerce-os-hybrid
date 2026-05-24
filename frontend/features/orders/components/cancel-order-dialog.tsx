"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { cancelOrderSchema, type CancelOrderFormValues } from "@/features/orders/schemas/order-actions.schema";
import { cn } from "@/lib/utils/cn";

type CancelOrderDialogProps = {
  open: boolean;
  isSubmitting?: boolean;
  error?: string;
  onClose: () => void;
  onSubmit: (values: { reason: string; note?: string }) => void;
};

export function CancelOrderDialog({ open, isSubmitting = false, error, onClose, onSubmit }: CancelOrderDialogProps) {
  const form = useForm<CancelOrderFormValues>({
    resolver: zodResolver(cancelOrderSchema),
    defaultValues: {
      reason: "",
      note: ""
    }
  });

  function handleClose() {
    form.reset();
    onClose();
  }

  function handleSubmit(values: CancelOrderFormValues) {
    if (isSubmitting) {
      return;
    }

    onSubmit({ reason: values.reason.trim(), note: values.note?.trim() || undefined });
  }

  return (
    <Dialog
      open={open}
      title="Batalkan pesanan?"
      description="Stok reserved akan dilepas oleh backend. Pastikan alasan pembatalan jelas untuk riwayat operasional."
      onClose={handleClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={handleClose}>
            Batal
          </Button>
          <Button type="submit" form="cancel-order-form" variant="danger" isLoading={isSubmitting} disabled={isSubmitting}>
            Batalkan pesanan
          </Button>
        </>
      }
    >
      <form id="cancel-order-form" className="space-y-4" onSubmit={form.handleSubmit(handleSubmit)}>
        <label className="block text-sm font-medium text-neutral-800">
          Alasan pembatalan
          <Input
            className="mt-2"
            hasError={!!form.formState.errors.reason}
            placeholder="Contoh: Customer meminta pembatalan"
            {...form.register("reason")}
          />
        </label>
        {form.formState.errors.reason ? (
          <p className="-mt-2 text-xs font-medium text-red-600">{form.formState.errors.reason.message}</p>
        ) : null}

        <label className="block text-sm font-medium text-neutral-800">
          Catatan internal
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

        {error ? <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p> : null}
      </form>
    </Dialog>
  );
}
