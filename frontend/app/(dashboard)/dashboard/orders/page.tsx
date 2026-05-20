"use client";

import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  OrderSourceLabel,
  OrderStatusBadge,
  PaymentStatusBadge
} from "@/features/orders/components/status-badges";
import { useOrders } from "@/features/orders/hooks/use-orders";
import type { OrderFilters, OrderListItem, OrderSource, OrderStatus, PaymentStatus } from "@/features/orders/types";
import { formatDateTime } from "@/lib/format/date";
import { formatRupiah } from "@/lib/format/money";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";

const pageLimit = 20;

export default function OrdersPage() {
  const router = useRouter();
  const userPermissions = useTenantStore((state) => state.permissions);
  const canRead = userPermissions.includes(permissions.orderRead);
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState<OrderStatus | "">("");
  const [paymentStatus, setPaymentStatus] = useState<PaymentStatus | "">("");
  const [source, setSource] = useState<OrderSource | "">("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [cursor, setCursor] = useState<string | null>(null);
  const [cursorStack, setCursorStack] = useState<string[]>([]);

  const filters = useMemo<OrderFilters>(
    () => ({
      query: query.trim(),
      status,
      paymentStatus,
      source,
      dateFrom,
      dateTo,
      cursor,
      limit: pageLimit
    }),
    [cursor, dateFrom, dateTo, paymentStatus, query, source, status]
  );
  const ordersQuery = useOrders(filters, canRead);

  function resetPagination() {
    setCursor(null);
    setCursorStack([]);
  }

  if (!canRead) {
    return (
      <EmptyState
        title="Akses pesanan belum tersedia"
        description="Role aktifmu belum memiliki izin untuk melihat pesanan."
      />
    );
  }

  if (ordersQuery.isPending) {
    return <LoadingState lines={5} />;
  }

  if (ordersQuery.isError) {
    return (
      <ErrorState
        title="Pesanan belum bisa dimuat"
        description="Coba muat ulang atau periksa filter yang digunakan."
        onRetry={() => void ordersQuery.refetch()}
      />
    );
  }

  const orders = ordersQuery.data.orders;
  const pagination = ordersQuery.data.pagination;
  const columns: Array<DataTableColumn<OrderListItem>> = [
    {
      key: "order",
      header: "Order",
      render: (order) => (
        <div>
          <button
            type="button"
            className="font-semibold text-primary-700 hover:text-primary-800"
            onClick={() => router.push(`/dashboard/orders/${order.id}`)}
          >
            {order.orderNumber}
          </button>
          <p className="mt-1 text-xs text-neutral-500">{formatDateTime(order.createdAt)}</p>
        </div>
      )
    },
    {
      key: "customer",
      header: "Customer",
      render: (order) => (
        <div>
          <p className="font-medium text-neutral-950">{order.customerName}</p>
          <p className="mt-1 text-xs text-neutral-500">{order.customerPhone}</p>
        </div>
      )
    },
    {
      key: "source",
      header: "Sumber",
      render: (order) => <OrderSourceLabel source={order.source} />
    },
    {
      key: "status",
      header: "Status",
      render: (order) => (
        <div className="space-y-2">
          <OrderStatusBadge status={order.status} />
          <PaymentStatusBadge status={order.paymentStatus} />
        </div>
      )
    },
    {
      key: "items",
      header: "Item",
      render: (order) => `${order.itemCount} item`
    },
    {
      key: "total",
      header: "Total",
      render: (order) => <span className="font-semibold text-neutral-950">{formatRupiah(order.grandTotal)}</span>
    },
    {
      key: "actions",
      header: "Aksi",
      render: (order) => (
        <Button type="button" variant="outline" size="sm" onClick={() => router.push(`/dashboard/orders/${order.id}`)}>
          Detail
        </Button>
      )
    }
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-neutral-950">Pesanan</h1>
        <p className="mt-1 text-sm text-neutral-500">
          Pantau order storefront, status pembayaran, dan proses pemenuhan pesanan.
        </p>
      </div>

      <div className="grid gap-3 rounded-2xl border border-neutral-200 bg-white p-4 lg:grid-cols-[1.2fr_160px_180px_160px_160px_160px]">
        <Input
          placeholder="Cari nomor order, nama, atau telepon..."
          value={query}
          onChange={(event) => {
            setQuery(event.target.value);
            resetPagination();
          }}
        />
        <select
          className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
          value={status}
          onChange={(event) => {
            setStatus(event.target.value as OrderStatus | "");
            resetPagination();
          }}
        >
          <option value="">Semua status</option>
          <option value="pending">Menunggu</option>
          <option value="confirmed">Terkonfirmasi</option>
          <option value="processing">Diproses</option>
          <option value="ready_to_ship">Siap dikirim</option>
          <option value="shipped">Dikirim</option>
          <option value="completed">Selesai</option>
          <option value="cancelled">Dibatalkan</option>
        </select>
        <select
          className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
          value={paymentStatus}
          onChange={(event) => {
            setPaymentStatus(event.target.value as PaymentStatus | "");
            resetPagination();
          }}
        >
          <option value="">Semua pembayaran</option>
          <option value="unpaid">Belum dibayar</option>
          <option value="waiting_confirmation">Menunggu review</option>
          <option value="paid">Lunas</option>
          <option value="failed">Gagal</option>
          <option value="refunded">Refund</option>
        </select>
        <select
          className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
          value={source}
          onChange={(event) => {
            setSource(event.target.value as OrderSource | "");
            resetPagination();
          }}
        >
          <option value="">Semua sumber</option>
          <option value="storefront">Storefront</option>
          <option value="pos">POS</option>
          <option value="whatsapp_manual">WhatsApp manual</option>
          <option value="admin_manual">Admin manual</option>
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

      {orders.length === 0 ? (
        <EmptyState
          title="Belum ada pesanan"
          description="Saat customer checkout dari storefront, pesanan akan muncul di sini."
        />
      ) : (
        <>
          <DataTable columns={columns} rows={orders} getRowKey={(order) => order.id} />
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
            <p className="text-sm text-neutral-500">Menampilkan maksimal {pagination.limit} pesanan per halaman.</p>
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
    </div>
  );
}
