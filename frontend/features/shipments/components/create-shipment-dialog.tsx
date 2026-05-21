"use client";

import { useState, type FormEvent } from "react";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import type { CourierType, CreateShipmentInput } from "@/features/shipments/types";

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
  const [localError, setLocalError] = useState<string>();

  function handleClose() {
    setLocalError(undefined);
    onClose();
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    const parsedShippingCost = Number(formData.get("shipping_cost") || 0);

    if (!Number.isFinite(parsedShippingCost) || parsedShippingCost < 0) {
      setLocalError("Ongkir harus berupa angka 0 atau lebih.");
      return;
    }

    onSubmit({
      courierType: String(formData.get("courier_type") ?? "internal") as CourierType,
      courierName: String(formData.get("courier_name") ?? "").trim() || undefined,
      trackingNumber: String(formData.get("tracking_number") ?? "").trim() || undefined,
      shippingCost: parsedShippingCost,
      assignedToName: String(formData.get("assigned_to_name") ?? "").trim() || undefined,
      assignedToPhone: String(formData.get("assigned_to_phone") ?? "").trim() || undefined,
      note: String(formData.get("note") ?? "").trim() || undefined
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
          <Button type="submit" form="create-shipment-form" isLoading={isSubmitting}>
            Buat shipment
          </Button>
        </>
      }
    >
      <form key={open ? "open-create-shipment" : "closed-create-shipment"} id="create-shipment-form" className="space-y-4" onSubmit={handleSubmit}>
        {(localError || error) ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            {localError || error}
          </div>
        ) : null}

        <label className="space-y-1">
          <span className="text-sm font-medium text-neutral-700">Tipe kurir</span>
          <select
            name="courier_type"
            className="h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            defaultValue="internal"
          >
            <option value="internal">Kurir internal</option>
            <option value="manual">Manual / pihak ketiga</option>
          </select>
        </label>

        <div className="grid gap-4 sm:grid-cols-2">
          <label className="space-y-1">
            <span className="text-sm font-medium text-neutral-700">Nama kurir opsional</span>
            <Input name="courier_name" />
          </label>
          <label className="space-y-1">
            <span className="text-sm font-medium text-neutral-700">Nomor resi opsional</span>
            <Input name="tracking_number" />
          </label>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <label className="space-y-1">
            <span className="text-sm font-medium text-neutral-700">Ongkir (Rp)</span>
            <Input name="shipping_cost" min={0} type="number" defaultValue={0} />
          </label>
          <label className="space-y-1">
            <span className="text-sm font-medium text-neutral-700">Nama petugas opsional</span>
            <Input name="assigned_to_name" />
          </label>
        </div>

        <label className="space-y-1">
          <span className="text-sm font-medium text-neutral-700">HP petugas opsional</span>
          <Input name="assigned_to_phone" />
        </label>

        <label className="space-y-1">
          <span className="text-sm font-medium text-neutral-700">Catatan internal opsional</span>
          <Input name="note" placeholder="Tidak tampil di tracking publik" />
        </label>
      </form>
    </Dialog>
  );
}
