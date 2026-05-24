"use client";

import { useState, type FormEvent } from "react";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
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
  const [localError, setLocalError] = useState<string>();

  function handleClose() {
    setLocalError(undefined);
    onClose();
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (isSubmitting) {
      return;
    }

    const formData = new FormData(event.currentTarget);
    const name = String(formData.get("name") ?? "").trim();
    const description = String(formData.get("description") ?? "").trim();
    const parsedRate = Number(formData.get("rate") ?? 0);
    const parsedSortOrder = Number(formData.get("sort_order") || 0);

    if (!name) {
      setLocalError("Nama zona wajib diisi.");
      return;
    }
    if (!Number.isFinite(parsedRate) || parsedRate < 0) {
      setLocalError("Ongkir harus berupa angka 0 atau lebih.");
      return;
    }
    if (!Number.isFinite(parsedSortOrder) || parsedSortOrder < 0) {
      setLocalError("Urutan harus berupa angka 0 atau lebih.");
      return;
    }

    onSubmit({
      name,
      description: description || undefined,
      rate: parsedRate,
      sortOrder: parsedSortOrder,
      isActive: formData.get("is_active") === "on"
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
      <form key={zone?.id ?? "new-zone"} id="courier-zone-form" className="space-y-4" onSubmit={handleSubmit}>
        {(localError || error) ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            {localError || error}
          </div>
        ) : null}

        <label className="space-y-1">
          <span className="text-sm font-medium text-neutral-700">Nama zona</span>
          <Input name="name" defaultValue={zone?.name ?? ""} placeholder="Contoh: Dalam kota" />
        </label>

        <label className="space-y-1">
          <span className="text-sm font-medium text-neutral-700">Deskripsi opsional</span>
          <Input
            name="description"
            defaultValue={zone?.description ?? ""}
            placeholder="Contoh: radius sekitar toko"
          />
        </label>

        <div className="grid gap-4 sm:grid-cols-2">
          <label className="space-y-1">
            <span className="text-sm font-medium text-neutral-700">Ongkir (Rp)</span>
            <Input name="rate" min={0} type="number" defaultValue={zone?.rate ?? 0} />
          </label>
          <label className="space-y-1">
            <span className="text-sm font-medium text-neutral-700">Urutan</span>
            <Input name="sort_order" min={0} type="number" defaultValue={zone?.sortOrder ?? 0} />
          </label>
        </div>

        <label className="flex items-center gap-2 rounded-xl border border-neutral-200 p-3 text-sm text-neutral-700">
          <input name="is_active" type="checkbox" defaultChecked={zone?.isActive ?? true} />
          Aktif dan tampil di checkout publik
        </label>
      </form>
    </Dialog>
  );
}
