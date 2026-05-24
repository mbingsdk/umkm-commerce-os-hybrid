"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { createShipmentSchema, type CreateShipmentFormValues } from "@/features/shipments/schemas/shipment.schema";
import type { CreateShipmentInput } from "@/features/shipments/types";

type CreateShipmentDialogProps = {
  open: boolean;
  orderNumber?: string;
  isSubmitting?: boolean;
  error?: string;
  onClose: () => void;
  onSubmit: (values: CreateShipmentInput) => void;
};

export function CreateShipmentDialog({
  open,
  orderNumber,
  isSubmitting = false,
  error,
  onClose,
  onSubmit
}: CreateShipmentDialogProps) {
  const form = useForm<CreateShipmentFormValues>({
    resolver: zodResolver(createShipmentSchema),
    defaultValues: defaultCreateShipmentValues()
  });

  useEffect(() => {
    if (open) {
      form.reset(defaultCreateShipmentValues());
    }
  }, [form, open]);

  function handleClose() {
    form.reset();
    onClose();
  }

  function handleSubmit(values: CreateShipmentFormValues) {
    if (isSubmitting) {
      return;
    }

    onSubmit({
      courierType: values.courierType,
      courierName: values.courierName?.trim() || undefined,
      trackingNumber: values.trackingNumber?.trim() || undefined,
      shippingCost: values.shippingCost,
      assignedToName: values.assignedToName?.trim() || undefined,
      assignedToPhone: values.assignedToPhone?.trim() || undefined,
      note: values.note?.trim() || undefined
    });
  }

  return (
    <Dialog
      open={open}
      title="Buat pengiriman"
      description={orderNumber ? `Buat shipment untuk pesanan ${orderNumber}.` : "Buat shipment untuk pesanan ini."}
      onClose={handleClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={handleClose}>
            Batal
          </Button>
          <Button type="submit" form="create-shipment-form" isLoading={isSubmitting} disabled={isSubmitting}>
            Buat shipment
          </Button>
        </>
      }
    >
      <form
        key={open ? "open-create-shipment" : "closed-create-shipment"}
        id="create-shipment-form"
        className="space-y-4"
        onSubmit={form.handleSubmit(handleSubmit)}
      >
        {error ? <div className="rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700">{error}</div> : null}

        <label className="space-y-1">
          <span className="text-sm font-medium text-neutral-700">Tipe kurir</span>
          <select
            className="h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            {...form.register("courierType")}
          >
            <option value="internal">Kurir internal</option>
            <option value="manual">Manual / pihak ketiga</option>
          </select>
        </label>
        {form.formState.errors.courierType ? (
          <p className="-mt-2 text-xs font-medium text-red-600">{form.formState.errors.courierType.message}</p>
        ) : null}

        <div className="grid gap-4 sm:grid-cols-2">
          <label className="space-y-1">
            <span className="text-sm font-medium text-neutral-700">Nama kurir opsional</span>
            <Input hasError={!!form.formState.errors.courierName} {...form.register("courierName")} />
            {form.formState.errors.courierName ? (
              <span className="block text-xs font-medium text-red-600">{form.formState.errors.courierName.message}</span>
            ) : null}
          </label>
          <label className="space-y-1">
            <span className="text-sm font-medium text-neutral-700">Nomor resi opsional</span>
            <Input hasError={!!form.formState.errors.trackingNumber} {...form.register("trackingNumber")} />
            {form.formState.errors.trackingNumber ? (
              <span className="block text-xs font-medium text-red-600">{form.formState.errors.trackingNumber.message}</span>
            ) : null}
          </label>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <label className="space-y-1">
            <span className="text-sm font-medium text-neutral-700">Ongkir (Rp)</span>
            <Input
              hasError={!!form.formState.errors.shippingCost}
              min={0}
              type="number"
              {...form.register("shippingCost", { valueAsNumber: true })}
            />
            {form.formState.errors.shippingCost ? (
              <span className="block text-xs font-medium text-red-600">{form.formState.errors.shippingCost.message}</span>
            ) : null}
          </label>
          <label className="space-y-1">
            <span className="text-sm font-medium text-neutral-700">Nama petugas opsional</span>
            <Input hasError={!!form.formState.errors.assignedToName} {...form.register("assignedToName")} />
            {form.formState.errors.assignedToName ? (
              <span className="block text-xs font-medium text-red-600">{form.formState.errors.assignedToName.message}</span>
            ) : null}
          </label>
        </div>

        <label className="space-y-1">
          <span className="text-sm font-medium text-neutral-700">HP petugas opsional</span>
          <Input hasError={!!form.formState.errors.assignedToPhone} {...form.register("assignedToPhone")} />
          {form.formState.errors.assignedToPhone ? (
            <span className="block text-xs font-medium text-red-600">{form.formState.errors.assignedToPhone.message}</span>
          ) : null}
        </label>

        <label className="space-y-1">
          <span className="text-sm font-medium text-neutral-700">Catatan internal opsional</span>
          <Input hasError={!!form.formState.errors.note} placeholder="Tidak tampil di tracking publik" {...form.register("note")} />
          {form.formState.errors.note ? (
            <span className="block text-xs font-medium text-red-600">{form.formState.errors.note.message}</span>
          ) : null}
        </label>
      </form>
    </Dialog>
  );
}

function defaultCreateShipmentValues(): CreateShipmentFormValues {
  return {
    courierType: "internal",
    courierName: "",
    trackingNumber: "",
    shippingCost: 0,
    assignedToName: "",
    assignedToPhone: "",
    note: ""
  };
}
