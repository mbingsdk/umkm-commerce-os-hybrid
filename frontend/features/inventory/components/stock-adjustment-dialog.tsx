"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import type { AdjustStockInput, InventoryStock } from "@/features/inventory/types";
import { cn } from "@/lib/utils/cn";

type StockAdjustmentDialogProps = {
  open: boolean;
  stock?: InventoryStock | null;
  isSubmitting?: boolean;
  error?: string;
  onClose: () => void;
  onSubmit: (values: AdjustStockInput) => void;
};

export function StockAdjustmentDialog({
  open,
  stock,
  isSubmitting = false,
  error,
  onClose,
  onSubmit
}: StockAdjustmentDialogProps) {
  const [adjustmentType, setAdjustmentType] = useState<AdjustStockInput["adjustmentType"]>("in");
  const [quantity, setQuantity] = useState("1");
  const [reason, setReason] = useState("");
  const [note, setNote] = useState("");
  const [confirmedStockOut, setConfirmedStockOut] = useState(false);
  const parsedQuantity = Number(quantity);
  const isStockOut = adjustmentType === "out";
  const quantityInvalid = !Number.isInteger(parsedQuantity) || parsedQuantity <= 0;
  const reasonInvalid = reason.trim().length === 0;
  const stockOutNeedsConfirm = isStockOut && !confirmedStockOut;
  const stockOutWouldExceedAvailable = isStockOut && !!stock && parsedQuantity > stock.quantityAvailable;
  const disabled = isSubmitting || quantityInvalid || reasonInvalid || stockOutNeedsConfirm || stockOutWouldExceedAvailable;

  function handleClose() {
    setAdjustmentType("in");
    setQuantity("1");
    setReason("");
    setNote("");
    setConfirmedStockOut(false);
    onClose();
  }

  return (
    <Dialog
      open={open}
      title="Sesuaikan stok"
      description="Gunakan hanya untuk koreksi manual seperti restock, opname, atau barang rusak. Semua perubahan dicatat sebagai movement."
      onClose={handleClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={handleClose}>
            Batal
          </Button>
          <Button
            type="button"
            variant={isStockOut ? "danger" : "primary"}
            isLoading={isSubmitting}
            disabled={disabled}
            onClick={() => {
              if (disabled) {
                return;
              }

              onSubmit({
                adjustmentType,
                quantity: parsedQuantity,
                reason: reason.trim(),
                note: note.trim() || undefined
              });
            }}
          >
            Simpan penyesuaian
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        {stock ? (
          <div className="rounded-xl bg-neutral-50 p-3 text-sm text-neutral-600">
            <p className="font-semibold text-neutral-950">{stock.name}</p>
            <p className="mt-1">
              Tersedia {stock.quantityAvailable} • Fisik {stock.quantityOnHand} • Reserved {stock.quantityReserved}
            </p>
          </div>
        ) : null}

        <label className="block text-sm font-medium text-neutral-800">
          Tipe penyesuaian
          <select
            className="mt-2 h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            value={adjustmentType}
            onChange={(event) => {
              setAdjustmentType(event.target.value as AdjustStockInput["adjustmentType"]);
              setConfirmedStockOut(false);
            }}
          >
            <option value="in">Stok masuk</option>
            <option value="out">Stok keluar</option>
          </select>
        </label>

        <label className="block text-sm font-medium text-neutral-800">
          Jumlah
          <Input
            className="mt-2"
            type="number"
            min={1}
            step={1}
            hasError={quantityInvalid || stockOutWouldExceedAvailable}
            value={quantity}
            onChange={(event) => setQuantity(event.target.value)}
          />
        </label>

        {stockOutWouldExceedAvailable ? (
          <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">
            Jumlah stok keluar tidak boleh melebihi stok tersedia saat ini.
          </p>
        ) : null}

        <label className="block text-sm font-medium text-neutral-800">
          Alasan
          <Input
            className="mt-2"
            hasError={reasonInvalid && !!error}
            placeholder="Contoh: Restock supplier, stock opname, barang rusak"
            value={reason}
            onChange={(event) => setReason(event.target.value)}
          />
        </label>

        <label className="block text-sm font-medium text-neutral-800">
          Catatan
          <textarea
            className={cn(
              "mt-2 min-h-24 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-sm text-neutral-950 shadow-sm outline-none transition placeholder:text-neutral-400 focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            )}
            placeholder="Opsional"
            value={note}
            onChange={(event) => setNote(event.target.value)}
          />
        </label>

        {isStockOut ? (
          <label className="flex items-start gap-3 rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
            <input
              className="mt-1"
              type="checkbox"
              checked={confirmedStockOut}
              onChange={(event) => setConfirmedStockOut(event.target.checked)}
            />
            <span>Saya paham stok keluar akan mengurangi stok fisik dan stok tersedia.</span>
          </label>
        ) : null}

        {error ? <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p> : null}
      </div>
    </Dialog>
  );
}
