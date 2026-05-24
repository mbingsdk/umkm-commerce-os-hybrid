"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { closeSessionSchema, type CloseSessionFormValues } from "@/features/pos/schemas/pos.schema";
import type { POSSession } from "@/features/pos/types";
import { formatRupiah } from "@/lib/format/money";

type CloseSessionDialogProps = {
  open: boolean;
  session: POSSession;
  closedSession?: POSSession | null;
  isSubmitting?: boolean;
  error?: string;
  onClose: () => void;
  onSubmit: (values: { closingCashAmount: number; note?: string }) => void;
};

export function CloseSessionDialog({
  open,
  session,
  closedSession,
  isSubmitting = false,
  error,
  onClose,
  onSubmit
}: CloseSessionDialogProps) {
  const form = useForm<CloseSessionFormValues>({
    resolver: zodResolver(closeSessionSchema),
    defaultValues: toFormValues(session)
  });

  useEffect(() => {
    if (open && !closedSession) {
      form.reset(toFormValues(session));
    }
  }, [closedSession, form, open, session]);

  function handleClose() {
    form.reset(toFormValues(session));
    onClose();
  }

  function handleSubmit(values: CloseSessionFormValues) {
    if (isSubmitting) {
      return;
    }

    onSubmit({ closingCashAmount: values.closingCashAmount, note: values.note?.trim() || undefined });
  }

  return (
    <Dialog
      open={open}
      title={closedSession ? "Sesi kasir ditutup" : "Tutup sesi kasir"}
      description={
        closedSession
          ? "Backend sudah menghitung expected cash dan selisih kas sesi ini."
          : "Masukkan kas aktual di laci. Expected cash dihitung backend berdasarkan kas awal dan transaksi tunai."
      }
      onClose={handleClose}
      footer={
        closedSession ? (
          <Button type="button" onClick={handleClose}>
            Selesai
          </Button>
        ) : (
          <>
            <Button type="button" variant="outline" onClick={handleClose}>
              Batal
            </Button>
            <Button type="submit" form="close-session-form" variant="danger" isLoading={isSubmitting} disabled={isSubmitting}>
              Tutup sesi
            </Button>
          </>
        )
      }
    >
      <div className="space-y-4">
        <div className="grid gap-3 sm:grid-cols-2">
          <InfoBox label="Kas awal" value={formatRupiah(session.openingCash)} />
          <InfoBox
            label="Expected cash"
            value={session.expectedCash != null ? formatRupiah(session.expectedCash) : "Dihitung saat tutup"}
          />
        </div>

        {closedSession ? (
          <div className="grid gap-3 sm:grid-cols-3">
            <InfoBox label="Kas tutup" value={formatRupiah(closedSession.closingCash ?? 0)} />
            <InfoBox label="Expected" value={formatRupiah(closedSession.expectedCash ?? 0)} />
            <InfoBox label="Selisih" value={formatRupiah(closedSession.difference ?? 0)} />
          </div>
        ) : (
          <form id="close-session-form" className="space-y-4" onSubmit={form.handleSubmit(handleSubmit)}>
            <label className="block text-sm font-medium text-neutral-800">
              Kas aktual saat tutup
              <Input
                className="mt-2"
                type="number"
                min={0}
                step={1}
                hasError={!!form.formState.errors.closingCashAmount}
                {...form.register("closingCashAmount", { valueAsNumber: true })}
              />
            </label>
            {form.formState.errors.closingCashAmount ? (
              <p className="-mt-2 text-xs font-medium text-red-600">{form.formState.errors.closingCashAmount.message}</p>
            ) : null}

            <label className="block text-sm font-medium text-neutral-800">
              Catatan
              <textarea
                className="mt-2 min-h-24 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-sm text-neutral-950 shadow-sm outline-none transition placeholder:text-neutral-400 focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
                placeholder="Opsional"
                {...form.register("note")}
              />
            </label>
            {form.formState.errors.note ? (
              <p className="-mt-2 text-xs font-medium text-red-600">{form.formState.errors.note.message}</p>
            ) : null}
          </form>
        )}

        {error ? <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p> : null}
      </div>
    </Dialog>
  );
}

function toFormValues(session: POSSession): CloseSessionFormValues {
  return {
    closingCashAmount: session.expectedCash ?? session.openingCash,
    note: ""
  };
}

function InfoBox({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl bg-neutral-50 p-3">
      <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">{label}</p>
      <p className="mt-1 font-semibold text-neutral-950">{value}</p>
    </div>
  );
}
