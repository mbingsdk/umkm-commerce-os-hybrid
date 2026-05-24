"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

type OpenSessionFormProps = {
  isSubmitting?: boolean;
  error?: string;
  onSubmit: (values: { openingCashAmount: number; note?: string }) => void;
};

export function OpenSessionForm({ isSubmitting = false, error, onSubmit }: OpenSessionFormProps) {
  const [openingCash, setOpeningCash] = useState("0");
  const [note, setNote] = useState("");
  const parsedOpeningCash = Number(openingCash);
  const invalid = !Number.isInteger(parsedOpeningCash) || parsedOpeningCash < 0;

  return (
    <Card className="mx-auto max-w-xl">
      <CardHeader>
        <CardTitle>Buka sesi kasir</CardTitle>
        <CardDescription>
          Sesi POS harus dibuka sebelum transaksi. Nominal kas awal dipakai untuk rekonsiliasi saat sesi ditutup.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form
          className="space-y-4"
          onSubmit={(event) => {
            event.preventDefault();
            if (isSubmitting) {
              return;
            }

            if (!invalid) {
              onSubmit({ openingCashAmount: parsedOpeningCash, note: note.trim() || undefined });
            }
          }}
        >
          <label className="block text-sm font-medium text-neutral-800">
            Kas awal
            <Input
              className="mt-2"
              type="number"
              min={0}
              step={1}
              hasError={invalid}
              value={openingCash}
              onChange={(event) => setOpeningCash(event.target.value)}
            />
          </label>

          <label className="block text-sm font-medium text-neutral-800">
            Catatan shift
            <textarea
              className="mt-2 min-h-24 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-sm text-neutral-950 shadow-sm outline-none transition placeholder:text-neutral-400 focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
              placeholder="Opsional"
              value={note}
              onChange={(event) => setNote(event.target.value)}
            />
          </label>

          {error ? <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p> : null}

          <Button className="w-full" type="submit" isLoading={isSubmitting} disabled={isSubmitting || invalid}>
            Buka sesi
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
