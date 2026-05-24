"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useMemo } from "react";
import { useForm, useWatch } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import {
  paymentReviewSchema,
  type PaymentReviewFormValues
} from "@/features/orders/schemas/order-actions.schema";
import type { PaymentConfirmation } from "@/features/orders/types";
import { formatDateTime } from "@/lib/format/date";
import { formatRupiah } from "@/lib/format/money";

type PaymentReviewDialogProps = {
  open: boolean;
  mode: "confirm" | "reject";
  confirmations: PaymentConfirmation[];
  isSubmitting?: boolean;
  error?: string;
  onClose: () => void;
  onSubmit: (values: { paymentConfirmationId?: string; note?: string }) => void;
};

export function PaymentReviewDialog({
  open,
  mode,
  confirmations,
  isSubmitting = false,
  error,
  onClose,
  onSubmit
}: PaymentReviewDialogProps) {
  const pendingConfirmations = useMemo(
    () => confirmations.filter((confirmation) => confirmation.status === "pending"),
    [confirmations]
  );
  const form = useForm<PaymentReviewFormValues>({
    resolver: zodResolver(paymentReviewSchema),
    defaultValues: {
      paymentConfirmationId: pendingConfirmations[0]?.id ?? "",
      note: ""
    }
  });
  const selectedConfirmationId = useWatch({ control: form.control, name: "paymentConfirmationId" });
  const isConfirm = mode === "confirm";

  useEffect(() => {
    if (open) {
      form.reset({
        paymentConfirmationId: pendingConfirmations[0]?.id ?? "",
        note: ""
      });
    }
  }, [form, open, pendingConfirmations]);

  function handleClose() {
    form.reset({ paymentConfirmationId: pendingConfirmations[0]?.id ?? "", note: "" });
    onClose();
  }

  function handleSubmit(values: PaymentReviewFormValues) {
    if (isSubmitting || !values.paymentConfirmationId) {
      return;
    }

    onSubmit({
      paymentConfirmationId: values.paymentConfirmationId,
      note: values.note?.trim() || undefined
    });
  }

  return (
    <Dialog
      open={open}
      title={isConfirm ? "Konfirmasi pembayaran?" : "Tolak konfirmasi pembayaran?"}
      description={
        isConfirm
          ? "Backend akan memvalidasi order dan mengubah status pembayaran menjadi lunas."
          : "Order tetap menunggu pembayaran; customer perlu mengirim konfirmasi yang benar."
      }
      onClose={handleClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={handleClose}>
            Batal
          </Button>
          <Button
            type="submit"
            form="payment-review-form"
            variant={isConfirm ? "primary" : "danger"}
            isLoading={isSubmitting}
            disabled={isSubmitting || !selectedConfirmationId}
          >
            {isConfirm ? "Konfirmasi pembayaran" : "Tolak pembayaran"}
          </Button>
        </>
      }
    >
      <form id="payment-review-form" className="space-y-4" onSubmit={form.handleSubmit(handleSubmit)}>
        {pendingConfirmations.length === 0 ? (
          <p className="rounded-xl bg-amber-50 p-3 text-sm text-amber-800">
            Belum ada konfirmasi pembayaran pending untuk direview.
          </p>
        ) : (
          <label className="block text-sm font-medium text-neutral-800">
            Konfirmasi yang direview
            <select
              className="mt-2 h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
              {...form.register("paymentConfirmationId")}
            >
              {pendingConfirmations.map((confirmation) => (
                <option key={confirmation.id} value={confirmation.id}>
                  {confirmation.payerName} • {formatRupiah(confirmation.transferAmount)} •{" "}
                  {formatDateTime(confirmation.transferDate)}
                </option>
              ))}
            </select>
          </label>
        )}
        {form.formState.errors.paymentConfirmationId ? (
          <p className="-mt-2 text-xs font-medium text-red-600">
            {form.formState.errors.paymentConfirmationId.message}
          </p>
        ) : null}

        <label className="block text-sm font-medium text-neutral-800">
          Catatan review
          <textarea
            className="mt-2 min-h-24 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-sm text-neutral-950 shadow-sm outline-none transition placeholder:text-neutral-400 focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
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
