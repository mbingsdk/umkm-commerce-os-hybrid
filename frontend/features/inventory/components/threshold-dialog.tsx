"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import type { InventoryStock } from "@/features/inventory/types";

type ThresholdDialogProps = {
  open: boolean;
  stock?: InventoryStock | null;
  isSubmitting?: boolean;
  error?: string;
  onClose: () => void;
  onSubmit: (value: number) => void;
};

export function ThresholdDialog({
  open,
  stock,
  isSubmitting = false,
  error,
  onClose,
  onSubmit
}: ThresholdDialogProps) {
  const [threshold, setThreshold] = useState("0");
  const parsedThreshold = Number(threshold);
  const invalid = !Number.isInteger(parsedThreshold) || parsedThreshold < 0;

  return (
    <Dialog
      open={open}
      title="Ubah batas stok menipis"
      description="Produk akan ditandai stok menipis saat stok tersedia kurang dari atau sama dengan batas ini."
      onClose={onClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={onClose}>
            Batal
          </Button>
          <Button
            type="button"
            isLoading={isSubmitting}
            disabled={invalid}
            onClick={() => onSubmit(parsedThreshold)}
          >
            Simpan threshold
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        {stock ? (
          <div className="rounded-xl bg-neutral-50 p-3 text-sm">
            <p className="font-semibold text-neutral-950">{stock.name}</p>
            <p className="mt-1 text-neutral-500">Threshold saat ini: {stock.lowStockThreshold}</p>
          </div>
        ) : null}

        <label className="block text-sm font-medium text-neutral-800">
          Batas stok menipis
          <Input
            className="mt-2"
            type="number"
            min={0}
            step={1}
            hasError={invalid}
            value={threshold}
            onChange={(event) => setThreshold(event.target.value)}
          />
        </label>

        {error ? <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p> : null}
      </div>
    </Dialog>
  );
}
