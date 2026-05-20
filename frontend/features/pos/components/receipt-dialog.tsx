"use client";

import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import type { POSCartLine, POSTransaction } from "@/features/pos/types";
import { formatDateTime } from "@/lib/format/date";
import { formatRupiah } from "@/lib/format/money";

type ReceiptDialogProps = {
  open: boolean;
  transaction: POSTransaction | null;
  items: POSCartLine[];
  onClose: () => void;
};

export function ReceiptDialog({ open, transaction, items, onClose }: ReceiptDialogProps) {
  return (
    <Dialog
      open={open}
      title="Transaksi berhasil"
      description="Total final berasal dari backend. Item di bawah adalah snapshot cart saat submit."
      onClose={onClose}
      footer={
        <Button type="button" onClick={onClose}>
          Selesai
        </Button>
      }
    >
      {!transaction ? null : (
        <div className="space-y-4 text-sm">
          <div className="rounded-xl bg-neutral-50 p-4">
            <p className="font-semibold text-neutral-950">{transaction.transactionNumber}</p>
            <p className="mt-1 text-neutral-500">{formatDateTime(transaction.createdAt)}</p>
          </div>

          <div className="space-y-2">
            {items.map((item) => (
              <div key={item.productId} className="flex items-start justify-between gap-3">
                <div>
                  <p className="font-medium text-neutral-950">{item.name}</p>
                  <p className="text-xs text-neutral-500">
                    {item.quantity} × {formatRupiah(item.price)}
                  </p>
                </div>
                <p className="font-semibold text-neutral-950">{formatRupiah(item.price * item.quantity)}</p>
              </div>
            ))}
          </div>

          <div className="space-y-2 border-t border-neutral-100 pt-4">
            <MoneyRow label="Subtotal" value={transaction.subtotal} />
            <MoneyRow label="Diskon" value={transaction.discountTotal} />
            <MoneyRow label="Pajak" value={transaction.taxTotal} />
            <MoneyRow label="Total final" value={transaction.grandTotal} strong />
            <MoneyRow label="Dibayar" value={transaction.amountPaid} />
            <MoneyRow label="Kembalian" value={transaction.changeAmount} strong />
          </div>
        </div>
      )}
    </Dialog>
  );
}

function MoneyRow({ label, value, strong = false }: { label: string; value: number; strong?: boolean }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <span className={strong ? "font-semibold text-neutral-950" : "text-neutral-500"}>{label}</span>
      <span className={strong ? "font-semibold text-neutral-950" : "text-neutral-700"}>{formatRupiah(value)}</span>
    </div>
  );
}
