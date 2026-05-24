"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
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
  const [closingCash, setClosingCash] = useState(String(session.expectedCash ?? session.openingCash));
  const [note, setNote] = useState("");
  const parsedClosingCash = Number(closingCash);
  const invalid = !Number.isInteger(parsedClosingCash) || parsedClosingCash < 0;

  return (
    <Dialog
      open={open}
      title={closedSession ? "Sesi kasir ditutup" : "Tutup sesi kasir"}
      description={
        closedSession
          ? "Backend sudah menghitung expected cash dan selisih kas sesi ini."
          : "Masukkan kas aktual di laci. Expected cash dihitung backend berdasarkan kas awal dan transaksi tunai."
      }
      onClose={onClose}
      footer={
        closedSession ? (
          <Button type="button" onClick={onClose}>
            Selesai
          </Button>
        ) : (
          <>
            <Button type="button" variant="outline" onClick={onClose}>
              Batal
            </Button>
            <Button
              type="button"
              variant="danger"
              isLoading={isSubmitting}
              disabled={isSubmitting || invalid}
              onClick={() => {
                if (isSubmitting || invalid) {
                  return;
                }

                onSubmit({ closingCashAmount: parsedClosingCash, note: note.trim() || undefined });
              }}
            >
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
          <>
            <label className="block text-sm font-medium text-neutral-800">
              Kas aktual saat tutup
              <Input
                className="mt-2"
                type="number"
                min={0}
                step={1}
                hasError={invalid}
                value={closingCash}
                onChange={(event) => setClosingCash(event.target.value)}
              />
            </label>

            <label className="block text-sm font-medium text-neutral-800">
              Catatan
              <textarea
                className="mt-2 min-h-24 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-sm text-neutral-950 shadow-sm outline-none transition placeholder:text-neutral-400 focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
                placeholder="Opsional"
                value={note}
                onChange={(event) => setNote(event.target.value)}
              />
            </label>
          </>
        )}

        {error ? <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p> : null}
      </div>
    </Dialog>
  );
}

function InfoBox({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl bg-neutral-50 p-3">
      <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">{label}</p>
      <p className="mt-1 font-semibold text-neutral-950">{value}</p>
    </div>
  );
}
