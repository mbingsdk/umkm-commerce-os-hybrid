"use client";

import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { usePOSTransactionDetail } from "@/features/pos/hooks/use-pos";
import type { POSTransactionItem } from "@/features/pos/types";
import { formatDateTime } from "@/lib/format/date";
import { formatRupiah } from "@/lib/format/money";

type TransactionDetailDialogProps = {
  transactionId: string | null;
  onClose: () => void;
};

export function TransactionDetailDialog({ transactionId, onClose }: TransactionDetailDialogProps) {
  const detailQuery = usePOSTransactionDetail(transactionId, !!transactionId);
  const columns: Array<DataTableColumn<POSTransactionItem>> = [
    {
      key: "product",
      header: "Produk",
      render: (item) => (
        <div>
          <p className="font-medium text-neutral-950">{item.productName}</p>
          <p className="mt-1 text-xs text-neutral-500">{item.sku || "Tanpa SKU"}</p>
        </div>
      )
    },
    { key: "quantity", header: "Qty", render: (item) => item.quantity },
    { key: "price", header: "Harga", render: (item) => formatRupiah(item.unitPrice) },
    { key: "subtotal", header: "Subtotal", render: (item) => formatRupiah(item.subtotal) }
  ];

  return (
    <Dialog
      open={!!transactionId}
      title="Detail transaksi POS"
      description="Nama dan harga item berasal dari snapshot transaksi backend."
      onClose={onClose}
      footer={
        <Button type="button" onClick={onClose}>
          Tutup
        </Button>
      }
    >
      {detailQuery.isPending ? (
        <LoadingState lines={3} />
      ) : detailQuery.isError ? (
        <ErrorState
          title="Detail transaksi gagal dimuat"
          description="Coba muat ulang detail transaksi."
          onRetry={() => void detailQuery.refetch()}
        />
      ) : !detailQuery.data ? (
        <EmptyState title="Transaksi tidak ditemukan" description="Detail transaksi belum tersedia." />
      ) : (
        <div className="space-y-4">
          <div className="rounded-xl bg-neutral-50 p-4 text-sm">
            <p className="font-semibold text-neutral-950">{detailQuery.data.transaction.transactionNumber}</p>
            <p className="mt-1 text-neutral-500">{formatDateTime(detailQuery.data.transaction.createdAt)}</p>
            <div className="mt-3 grid gap-2 sm:grid-cols-2">
              <p>Total: {formatRupiah(detailQuery.data.transaction.grandTotal)}</p>
              <p>Kembalian: {formatRupiah(detailQuery.data.transaction.changeAmount)}</p>
            </div>
          </div>
          <DataTable columns={columns} rows={detailQuery.data.items} getRowKey={(item) => item.id} />
        </div>
      )}
    </Dialog>
  );
}
