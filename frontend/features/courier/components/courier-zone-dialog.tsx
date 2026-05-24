"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { courierZoneSchema, type CourierZoneFormValues } from "@/features/courier/schemas/courier.schema";
import type { CourierZone, CourierZoneInput } from "@/features/courier/types";

type CourierZoneDialogProps = {
  open: boolean;
  zone?: CourierZone | null;
  isSubmitting?: boolean;
  error?: string;
  onClose: () => void;
  onSubmit: (values: CourierZoneInput) => void;
};

export function CourierZoneDialog({
  open,
  zone,
  isSubmitting = false,
  error,
  onClose,
  onSubmit
}: CourierZoneDialogProps) {
  const form = useForm<CourierZoneFormValues>({
    resolver: zodResolver(courierZoneSchema),
    defaultValues: toFormValues(zone)
  });

  useEffect(() => {
    if (open) {
      form.reset(toFormValues(zone));
    }
  }, [form, open, zone]);

  function handleClose() {
    form.reset(toFormValues(zone));
    onClose();
  }

  function handleSubmit(values: CourierZoneFormValues) {
    if (isSubmitting) {
      return;
    }

    onSubmit({
      name: values.name.trim(),
      description: values.description?.trim() || undefined,
      rate: values.rate,
      sortOrder: values.sortOrder,
      isActive: values.isActive
    });
  }

  return (
    <Dialog
      open={open}
      title={zone ? "Edit zona kurir" : "Tambah zona kurir"}
      description="Zona kurir dipakai customer untuk melihat opsi ongkir lokal toko."
      onClose={handleClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={handleClose}>
            Batal
          </Button>
          <Button type="submit" form="courier-zone-form" isLoading={isSubmitting} disabled={isSubmitting}>
            Simpan
          </Button>
        </>
      }
    >
      <form key={zone?.id ?? "new-zone"} id="courier-zone-form" className="space-y-4" onSubmit={form.handleSubmit(handleSubmit)}>
        {error ? <div className="rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700">{error}</div> : null}

        <label className="space-y-1">
          <span className="text-sm font-medium text-neutral-700">Nama zona</span>
          <Input hasError={!!form.formState.errors.name} placeholder="Contoh: Dalam kota" {...form.register("name")} />
          {form.formState.errors.name ? (
            <span className="block text-xs font-medium text-red-600">{form.formState.errors.name.message}</span>
          ) : null}
        </label>

        <label className="space-y-1">
          <span className="text-sm font-medium text-neutral-700">Deskripsi opsional</span>
          <Input
            hasError={!!form.formState.errors.description}
            placeholder="Contoh: radius sekitar toko"
            {...form.register("description")}
          />
          {form.formState.errors.description ? (
            <span className="block text-xs font-medium text-red-600">{form.formState.errors.description.message}</span>
          ) : null}
        </label>

        <div className="grid gap-4 sm:grid-cols-2">
          <label className="space-y-1">
            <span className="text-sm font-medium text-neutral-700">Ongkir (Rp)</span>
            <Input hasError={!!form.formState.errors.rate} min={0} type="number" {...form.register("rate", { valueAsNumber: true })} />
            {form.formState.errors.rate ? (
              <span className="block text-xs font-medium text-red-600">{form.formState.errors.rate.message}</span>
            ) : null}
          </label>
          <label className="space-y-1">
            <span className="text-sm font-medium text-neutral-700">Urutan</span>
            <Input
              hasError={!!form.formState.errors.sortOrder}
              min={0}
              type="number"
              {...form.register("sortOrder", { valueAsNumber: true })}
            />
            {form.formState.errors.sortOrder ? (
              <span className="block text-xs font-medium text-red-600">{form.formState.errors.sortOrder.message}</span>
            ) : null}
          </label>
        </div>

        <label className="flex items-center gap-2 rounded-xl border border-neutral-200 p-3 text-sm text-neutral-700">
          <input type="checkbox" {...form.register("isActive")} />
          Aktif dan tampil di checkout publik
        </label>
      </form>
    </Dialog>
  );
}

function toFormValues(zone?: CourierZone | null): CourierZoneFormValues {
  return {
    name: zone?.name ?? "",
    description: zone?.description ?? "",
    rate: zone?.rate ?? 0,
    sortOrder: zone?.sortOrder ?? 0,
    isActive: zone?.isActive ?? true
  };
}
