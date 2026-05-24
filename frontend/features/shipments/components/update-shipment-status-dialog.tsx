"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { shipmentStatusOptions } from "@/features/shipments/constants";
import {
  updateShipmentStatusSchema,
  type UpdateShipmentStatusFormValues
} from "@/features/shipments/schemas/shipment.schema";
import type { ShipmentStatus, UpdateShipmentStatusInput } from "@/features/shipments/types";

type UpdateShipmentStatusDialogProps = {
  open: boolean;
  currentStatus?: string;
  isSubmitting?: boolean;
  error?: string;
  onClose: () => void;
  onSubmit: (values: UpdateShipmentStatusInput) => void;
};

export function UpdateShipmentStatusDialog({
  open,
  currentStatus = "pending",
  isSubmitting = false,
  error,
  onClose,
  onSubmit
}: UpdateShipmentStatusDialogProps) {
  const form = useForm<UpdateShipmentStatusFormValues>({
    resolver: zodResolver(updateShipmentStatusSchema),
    defaultValues: {
      status: nextShipmentStatus(currentStatus),
      note: ""
    }
  });

  useEffect(() => {
    if (open) {
      form.reset({ status: nextShipmentStatus(currentStatus), note: "" });
    }
  }, [currentStatus, form, open]);

  function handleSubmit(values: UpdateShipmentStatusFormValues) {
    if (isSubmitting) {
      return;
    }

    onSubmit({
      status: values.status,
      note: values.note?.trim() || undefined
    });
  }

  return (
    <Dialog
      open={open}
      title="Update status pengiriman"
      description="Backend tetap memvalidasi transisi status yang diperbolehkan."
      onClose={onClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={onClose}>
            Batal
          </Button>
          <Button type="submit" form="update-shipment-status-form" isLoading={isSubmitting} disabled={isSubmitting}>
            Update status
          </Button>
        </>
      }
    >
      <form
        key={currentStatus}
        id="update-shipment-status-form"
        className="space-y-4"
        onSubmit={form.handleSubmit(handleSubmit)}
      >
        {error ? <div className="rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700">{error}</div> : null}

        <label className="space-y-1">
          <span className="text-sm font-medium text-neutral-700">Status baru</span>
          <select
            className="h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            {...form.register("status")}
          >
            {shipmentStatusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        {form.formState.errors.status ? (
          <p className="-mt-2 text-xs font-medium text-red-600">{form.formState.errors.status.message}</p>
        ) : null}

        <label className="space-y-1">
          <span className="text-sm font-medium text-neutral-700">Catatan opsional</span>
          <Input
            hasError={!!form.formState.errors.note}
            placeholder="Contoh: paket sudah diterima kurir"
            {...form.register("note")}
          />
          {form.formState.errors.note ? (
            <span className="block text-xs font-medium text-red-600">{form.formState.errors.note.message}</span>
          ) : null}
        </label>
      </form>
    </Dialog>
  );
}

function nextShipmentStatus(status: string): ShipmentStatus {
  switch (status) {
    case "pending":
      return "ready_for_pickup";
    case "ready_for_pickup":
      return "picked_up";
    case "picked_up":
      return "on_delivery";
    case "on_delivery":
      return "delivered";
    case "failed":
      return "ready_for_pickup";
    default:
      return "pending";
  }
}
