"use client";

import { useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
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
  const [selectedId, setSelectedId] = useState("");
  const [note, setNote] = useState("");
  const isConfirm = mode === "confirm";
  const selectedConfirmationId = pendingConfirmations.some((confirmation) => confirmation.id === selectedId)
    ? selectedId
    : pendingConfirmations[0]?.id ?? "";

  function handleClose() {
    setSelectedId("");
    setNote("");
    onClose();
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
            type="button"
            variant={isConfirm ? "primary" : "danger"}
            isLoading={isSubmitting}
            disabled={isSubmitting || !selectedConfirmationId}
            onClick={() => {
              if (isSubmitting || !selectedConfirmationId) {
                return;
              }

              onSubmit({ paymentConfirmationId: selectedConfirmationId || undefined, note: note.trim() || undefined });
            }}
          >
            {isConfirm ? "Konfirmasi pembayaran" : "Tolak pembayaran"}
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        {pendingConfirmations.length === 0 ? (
          <p className="rounded-xl bg-amber-50 p-3 text-sm text-amber-800">
            Belum ada konfirmasi pembayaran pending untuk direview.
          </p>
        ) : (
          <label className="block text-sm font-medium text-neutral-800">
            Konfirmasi yang direview
            <select
              className="mt-2 h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
              value={selectedConfirmationId}
              onChange={(event) => setSelectedId(event.target.value)}
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

        <label className="block text-sm font-medium text-neutral-800">
          Catatan review
          <textarea
            className="mt-2 min-h-24 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-sm text-neutral-950 shadow-sm outline-none transition placeholder:text-neutral-400 focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            placeholder="Opsional"
            value={note}
            onChange={(event) => setNote(event.target.value)}
          />
        </label>

        {error ? <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p> : null}
      </div>
    </Dialog>
  );
}
