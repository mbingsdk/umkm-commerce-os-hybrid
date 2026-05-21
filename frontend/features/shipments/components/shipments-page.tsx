"use client";

import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { courierTypeLabels, shipmentStatusOptions } from "@/features/shipments/constants";
import { ShipmentStatusBadge } from "@/features/shipments/components/shipment-status-badge";
import { useShipments } from "@/features/shipments/hooks/use-shipments";
import type { Shipment, ShipmentFilters } from "@/features/shipments/types";
import { formatDateTime } from "@/lib/format/date";
import { formatRupiah } from "@/lib/format/money";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";

const pageLimit = 20;

export function ShipmentsPage() {
  const router = useRouter();
  const userPermissions = useTenantStore((state) => state.permissions);
  const canRead = userPermissions.includes(permissions.shipmentRead);
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [cursor, setCursor] = useState<string | null>(null);
  const [cursorStack, setCursorStack] = useState<string[]>([]);
  const filters = useMemo<ShipmentFilters>(
    () => ({
      query: query.trim(),
      status: status as ShipmentFilters["status"],
      dateFrom,
      dateTo,
      cursor,
      limit: pageLimit
    }),
    [cursor, dateFrom, dateTo, query, status]
  );
  const shipmentsQuery = useShipments(filters, canRead);

  function resetPagination() {
    setCursor(null);
    setCursorStack([]);
  }

  if (!canRead) {
    return (
      <EmptyState
        title="Akses pengiriman belum tersedia"
        description="Role aktifmu belum memiliki izin untuk melihat daftar shipment."
      />
    );
  }

  if (shipmentsQuery.isPending) {
    return <LoadingState lines={5} />;
  }

  if (shipmentsQuery.isError) {
    return (
      <ErrorState
        title="Daftar pengiriman belum bisa dimuat"
        description="Coba muat ulang atau ubah filter pencarian."
        onRetry={() => void shipmentsQuery.refetch()}
      />
    );
  }

  const shipments = shipmentsQuery.data.shipments;
  const pagination = shipmentsQuery.data.pagination;
  const columns: Array<DataTableColumn<Shipment>> = [
    {
      key: "order",
      header: "Order",
      render: (shipment) => (
        <div>
          <button
            type="button"
            className="font-semibold text-primary-700 hover:text-primary-800"
            onClick={() => router.push(`/dashboard/shipments/${shipment.id}`)}
          >
            {shipment.orderNumber}
          </button>
          <p className="mt-1 text-xs text-neutral-500">{formatDateTime(shipment.createdAt)}</p>
        </div>
      )
    },
    {
      key: "courier",
      header: "Kurir",
      render: (shipment) => (
        <div>
          <p className="font-medium text-neutral-950">
            {shipment.courierName || courierTypeLabels[shipment.courierType] || shipment.courierType}
          </p>
          <p className="mt-1 text-xs text-neutral-500">{shipment.trackingNumber || "Tanpa nomor resi"}</p>
        </div>
      )
    },
    {
      key: "status",
      header: "Status",
      render: (shipment) => <ShipmentStatusBadge status={shipment.status} />
    },
    {
      key: "cost",
      header: "Ongkir",
      render: (shipment) => formatRupiah(shipment.shippingCost)
    },
    {
      key: "assigned",
      header: "Petugas",
      render: (shipment) => shipment.assignedToName || shipment.assignedToPhone || "Belum ditugaskan"
    },
    {
      key: "actions",
      header: "Aksi",
      render: (shipment) => (
        <Button type="button" variant="outline" size="sm" onClick={() => router.push(`/dashboard/shipments/${shipment.id}`)}>
          Detail
        </Button>
      )
    }
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-neutral-950">Pengiriman</h1>
        <p className="mt-1 text-sm text-neutral-500">
          Pantau shipment dari order yang sudah siap dikirim dan update statusnya secara bertahap.
        </p>
      </div>

      <div className="grid gap-3 rounded-2xl border border-neutral-200 bg-white p-4 lg:grid-cols-[1.2fr_180px_160px_160px]">
        <Input
          placeholder="Cari nomor order, resi, kurir, atau petugas..."
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
            setStatus(event.target.value);
            resetPagination();
          }}
        >
          <option value="">Semua status</option>
          {shipmentStatusOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
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

      {shipments.length === 0 ? (
        <EmptyState
          title="Belum ada pengiriman"
          description="Shipment akan muncul setelah dibuat dari detail pesanan yang sudah dibayar dan siap diproses."
          action={
            <Button type="button" variant="outline" onClick={() => router.push("/dashboard/orders")}>
              Lihat pesanan
            </Button>
          }
        />
      ) : (
        <>
          <DataTable columns={columns} rows={shipments} getRowKey={(shipment) => shipment.id} />
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
            <p className="text-sm text-neutral-500">Menampilkan maksimal {pagination.limit} pengiriman per halaman.</p>
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
