"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm, useWatch } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  businessHoursSchema,
  type BusinessHoursFormValues
} from "@/features/settings/schemas/store.schema";

const dayLabels = ["Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"];

type StoreBusinessHoursFormProps = {
  initialValues: BusinessHoursFormValues;
  canUpdate: boolean;
  isSubmitting?: boolean;
  error?: string;
  onSubmit: (values: BusinessHoursFormValues) => void;
};

export function StoreBusinessHoursForm({
  initialValues,
  canUpdate,
  isSubmitting = false,
  error,
  onSubmit
}: StoreBusinessHoursFormProps) {
  const form = useForm<BusinessHoursFormValues>({
    resolver: zodResolver(businessHoursSchema),
    defaultValues: initialValues
  });
  const items = useWatch({ control: form.control, name: "items" });

  useEffect(() => {
    form.reset(initialValues);
  }, [form, initialValues]);

  function handleSubmit(values: BusinessHoursFormValues) {
    if (!canUpdate || isSubmitting) {
      return;
    }

    onSubmit(values);
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Jam operasional</CardTitle>
        <CardDescription>
          Dashboard API saat ini belum menyediakan pembacaan jam tersimpan. Form ini dipakai untuk mengganti jadwal
          operasional; simpan hanya jika jadwal sudah benar.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form className="space-y-4" onSubmit={form.handleSubmit(handleSubmit)}>
          <div className="space-y-3">
            {initialValues.items.map((item, index) => {
              const closed = items?.[index]?.isClosed ?? item.isClosed;
              const itemError = form.formState.errors.items?.[index];

              return (
                <div
                  key={item.dayOfWeek}
                  className="grid gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 p-4 md:grid-cols-[120px_1fr_1fr_120px]"
                >
                  <input type="hidden" {...form.register(`items.${index}.dayOfWeek`, { valueAsNumber: true })} />
                  <div>
                    <p className="text-sm font-semibold text-neutral-950">{dayLabels[index]}</p>
                    <p className="text-xs text-neutral-500">Hari {item.dayOfWeek}</p>
                  </div>

                  <Field label="Buka" error={itemError?.openTime?.message}>
                    <Input
                      type="time"
                      disabled={!canUpdate || isSubmitting || closed}
                      hasError={!!itemError?.openTime}
                      {...form.register(`items.${index}.openTime`)}
                    />
                  </Field>

                  <Field label="Tutup" error={itemError?.closeTime?.message}>
                    <Input
                      type="time"
                      disabled={!canUpdate || isSubmitting || closed}
                      hasError={!!itemError?.closeTime}
                      {...form.register(`items.${index}.closeTime`)}
                    />
                  </Field>

                  <label className="flex items-center gap-2 text-sm font-medium text-neutral-700 md:justify-end">
                    <input
                      type="checkbox"
                      className="h-4 w-4 rounded border-neutral-300 text-primary-600 focus:ring-primary-500"
                      disabled={!canUpdate || isSubmitting}
                      {...form.register(`items.${index}.isClosed`)}
                    />
                    Tutup
                  </label>
                </div>
              );
            })}
          </div>

          {error ? <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p> : null}

          <div className="flex justify-end">
            <Button
              type="submit"
              isLoading={isSubmitting}
              disabled={!canUpdate || isSubmitting || !form.formState.isDirty}
            >
              Simpan jam operasional
            </Button>
          </div>

          {!canUpdate ? (
            <p className="rounded-xl bg-amber-50 p-3 text-sm text-amber-700">
              Role aktifmu belum bisa mengubah jam operasional toko.
            </p>
          ) : null}
        </form>
      </CardContent>
    </Card>
  );
}

function Field({
  label,
  children,
  error
}: {
  label: string;
  children: React.ReactNode;
  error?: string;
}) {
  return (
    <label className="block space-y-2 text-sm font-medium text-neutral-700">
      {label}
      {children}
      {error ? <span className="block text-sm font-normal text-red-600">{error}</span> : null}
    </label>
  );
}
