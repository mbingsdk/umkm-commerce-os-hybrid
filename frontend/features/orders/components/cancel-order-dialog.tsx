"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils/cn";

type CancelOrderDialogProps = {
  open: boolean;
  isSubmitting?: boolean;
  error?: string;
  onClose: () => void;
  onSubmit: (values: { reason: string; note?: string }) => void;
};

export function CancelOrderDialog({ open, isSubmitting = false, error, onClose, onSubmit }: CancelOrderDialogProps) {
  const [reason, setReason] = useState("");
  const [note, setNote] = useState("");
  const reasonInvalid = reason.trim().length === 0;

  return (
    <Dialog
      open={open}
      title="Batalkan pesanan?"
      description="Stok reserved akan dilepas oleh backend. Pastikan alasan pembatalan jelas untuk riwayat operasional."
      onClose={onClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={onClose}>
            Batal
          </Button>
          <Button
            type="button"
            variant="danger"
            isLoading={isSubmitting}
            disabled={reasonInvalid}
            onClick={() => onSubmit({ reason: reason.trim(), note: note.trim() || undefined })}
          >
            Batalkan pesanan
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        <label className="block text-sm font-medium text-neutral-800">
          Alasan pembatalan
          <Input
            className="mt-2"
            hasError={!!error && reasonInvalid}
            placeholder="Contoh: Customer meminta pembatalan"
            value={reason}
            onChange={(event) => setReason(event.target.value)}
          />
        </label>

        <label className="block text-sm font-medium text-neutral-800">
          Catatan internal
          <textarea
            className={cn(
              "mt-2 min-h-24 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-sm text-neutral-950 shadow-sm outline-none transition placeholder:text-neutral-400 focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            )}
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
