"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { openSessionSchema, type OpenSessionFormValues } from "@/features/pos/schemas/pos.schema";

type OpenSessionFormProps = {
  isSubmitting?: boolean;
  error?: string;
  onSubmit: (values: { openingCashAmount: number; note?: string }) => void;
};

export function OpenSessionForm({ isSubmitting = false, error, onSubmit }: OpenSessionFormProps) {
  const form = useForm<OpenSessionFormValues>({
    resolver: zodResolver(openSessionSchema),
    defaultValues: {
      openingCashAmount: 0,
      note: ""
    }
  });

  function handleSubmit(values: OpenSessionFormValues) {
    if (isSubmitting) {
      return;
    }

    onSubmit({ openingCashAmount: values.openingCashAmount, note: values.note?.trim() || undefined });
  }

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
          onSubmit={form.handleSubmit(handleSubmit)}
        >
          <label className="block text-sm font-medium text-neutral-800">
            Kas awal
            <Input
              className="mt-2"
              type="number"
              min={0}
              step={1}
              hasError={!!form.formState.errors.openingCashAmount}
              {...form.register("openingCashAmount", { valueAsNumber: true })}
            />
          </label>
          {form.formState.errors.openingCashAmount ? (
            <p className="-mt-2 text-xs font-medium text-red-600">{form.formState.errors.openingCashAmount.message}</p>
          ) : null}

          <label className="block text-sm font-medium text-neutral-800">
            Catatan shift
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

          <Button className="w-full" type="submit" isLoading={isSubmitting} disabled={isSubmitting}>
            Buka sesi
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
