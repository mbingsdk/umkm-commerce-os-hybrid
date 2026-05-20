"use client";

import { useMemo, useState } from "react";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { TransactionDetailDialog } from "@/features/pos/components/transaction-detail-dialog";
import { usePOSTransactions } from "@/features/pos/hooks/use-pos";
import type { POSPaymentMethod, POSTransaction, POSTransactionFilters } from "@/features/pos/types";
import { formatDateTime } from "@/lib/format/date";
import { formatRupiah } from "@/lib/format/money";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";

const pageLimit = 20;

export default function POSHistoryPage() {
  const userPermissions = useTenantStore((state) => state.permissions);
  const canRead = userPermissions.includes(permissions.posReadTransaction);
  const [paymentMethod, setPaymentMethod] = useState<POSPaymentMethod | "">("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [cursor, setCursor] = useState<string | null>(null);
  const [cursorStack, setCursorStack] = useState<string[]>([]);
  const [selectedTransactionId, setSelectedTransactionId] = useState<string | null>(null);
  const filters = useMemo<POSTransactionFilters>(
    () => ({
      paymentMethod,
      dateFrom,
      dateTo,
      cursor,
      limit: pageLimit
    }),
    [cursor, dateFrom, dateTo, paymentMethod]
  );
  const transactionsQuery = usePOSTransactions(filters, canRead);

  function resetPagination() {
    setCursor(null);
    setCursorStack([]);
  }

  if (!canRead) {
    return (
      <EmptyState
        title="Akses riwayat POS belum tersedia"
        description="Role aktifmu belum memiliki izin untuk melihat transaksi POS."
      />
    );
  }

  if (transactionsQuery.isPending) {
    return <LoadingState lines={5} />;
  }

  if (transactionsQuery.isError) {
    return (
      <ErrorState
        title="Riwayat POS gagal dimuat"
        description="Coba muat ulang atau ubah filter."
        onRetry={() => void transactionsQuery.refetch()}
      />
    );
  }

  const transactions = transactionsQuery.data.transactions;
  const pagination = transactionsQuery.data.pagination;
  const columns: Array<DataTableColumn<POSTransaction>> = [
    {
      key: "transaction",
      header: "Transaksi",
      render: (transaction) => (
        <div>
          <button
            type="button"
            className="font-semibold text-primary-700 hover:text-primary-800"
            onClick={() => setSelectedTransactionId(transaction.id)}
          >
            {transaction.transactionNumber}
          </button>
          <p className="mt-1 text-xs text-neutral-500">{formatDateTime(transaction.createdAt)}</p>
        </div>
      )
    },
    {
      key: "payment",
      header: "Pembayaran",
      render: (transaction) => paymentMethodLabel(transaction.paymentMethod)
    },
    {
      key: "total",
      header: "Total",
      render: (transaction) => <span className="font-semibold text-neutral-950">{formatRupiah(transaction.grandTotal)}</span>
    },
    {
      key: "paid",
      header: "Dibayar / Kembali",
      render: (transaction) => (
        <div>
          <p>{formatRupiah(transaction.amountPaid)}</p>
          <p className="mt-1 text-xs text-neutral-500">Kembali {formatRupiah(transaction.changeAmount)}</p>
        </div>
      )
    },
    {
      key: "status",
      header: "Status",
      render: (transaction) => transaction.status
    },
    {
      key: "actions",
      header: "Aksi",
      render: (transaction) => (
        <Button type="button" variant="outline" size="sm" onClick={() => setSelectedTransactionId(transaction.id)}>
          Detail
        </Button>
      )
    }
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-neutral-950">Riwayat POS</h1>
        <p className="mt-1 text-sm text-neutral-500">Lihat transaksi POS online-first berdasarkan tanggal dan metode bayar.</p>
      </div>

      <div className="grid gap-3 rounded-2xl border border-neutral-200 bg-white p-4 md:grid-cols-[180px_180px_180px]">
        <select
          className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
          value={paymentMethod}
          onChange={(event) => {
            setPaymentMethod(event.target.value as POSPaymentMethod | "");
            resetPagination();
          }}
        >
          <option value="">Semua metode</option>
          <option value="cash">Tunai</option>
          <option value="qris_manual">QRIS manual</option>
        </select>
        <Input
          type="date"
          value={dateFrom}
          onChange={(event) => {
            setDateFrom(event.target.value);
            resetPagination();
          }}
        />
        <Input
          type="date"
          value={dateTo}
          onChange={(event) => {
            setDateTo(event.target.value);
            resetPagination();
          }}
        />
      </div>

      {transactions.length === 0 ? (
        <EmptyState
          title="Belum ada transaksi POS"
          description="Transaksi dari halaman POS akan muncul di sini setelah pembayaran berhasil."
        />
      ) : (
        <>
          <DataTable columns={columns} rows={transactions} getRowKey={(transaction) => transaction.id} />
          <div className="flex items-center justify-between gap-3">
            <Button
              type="button"
              variant="outline"
              disabled={cursorStack.length === 0}
              onClick={() => {
                const previous = cursorStack[cursorStack.length - 1];
                setCursor(previous || null);
                setCursorStack((items) => items.slice(0, -1));
              }}
            >
              Sebelumnya
            </Button>
            <p className="text-sm text-neutral-500">Menampilkan maksimal {pagination.limit} transaksi per halaman.</p>
            <Button
              type="button"
              variant="outline"
              disabled={!pagination.hasMore || !pagination.nextCursor}
              onClick={() => {
                setCursorStack((items) => [...items, cursor ?? ""]);
                setCursor(pagination.nextCursor ?? null);
              }}
            >
              Berikutnya
            </Button>
          </div>
        </>
      )}

      <TransactionDetailDialog
        transactionId={selectedTransactionId}
        onClose={() => setSelectedTransactionId(null)}
      />
    </div>
  );
}

function paymentMethodLabel(method: string) {
  if (method === "cash") {
    return "Tunai";
  }
  if (method === "qris_manual") {
    return "QRIS manual";
  }
  return method;
}
