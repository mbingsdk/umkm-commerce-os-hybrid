"use client";

import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { StockStatusBadge } from "@/features/inventory/components/stock-status-badge";
import { useInventoryStocks } from "@/features/inventory/hooks/use-inventory";
import type { InventoryStock, StockFilters } from "@/features/inventory/types";
import { formatDateTime } from "@/lib/format/date";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";

const pageLimit = 20;

export default function InventoryPage() {
  const router = useRouter();
  const userPermissions = useTenantStore((state) => state.permissions);
  const canRead = userPermissions.includes(permissions.inventoryRead);
  const [query, setQuery] = useState("");
  const [lowStock, setLowStock] = useState(false);
  const [outOfStock, setOutOfStock] = useState(false);
  const [cursor, setCursor] = useState<string | null>(null);
  const [cursorStack, setCursorStack] = useState<string[]>([]);

  const filters = useMemo<StockFilters>(
    () => ({
      query: query.trim(),
      lowStock,
      outOfStock,
      cursor,
      limit: pageLimit
    }),
    [cursor, lowStock, outOfStock, query]
  );
  const stocksQuery = useInventoryStocks(filters, canRead);

  function resetPagination() {
    setCursor(null);
    setCursorStack([]);
  }

  if (!canRead) {
    return (
      <EmptyState
        title="Akses inventori belum tersedia"
        description="Role aktifmu belum memiliki izin untuk melihat stok produk."
      />
    );
  }

  if (stocksQuery.isPending) {
    return <LoadingState lines={5} />;
  }

  if (stocksQuery.isError) {
    return (
      <ErrorState
        title="Inventori belum bisa dimuat"
        description="Coba muat ulang atau periksa filter stok yang digunakan."
        onRetry={() => void stocksQuery.refetch()}
      />
    );
  }

  const stocks = stocksQuery.data.stocks;
  const pagination = stocksQuery.data.pagination;
  const columns: Array<DataTableColumn<InventoryStock>> = [
    {
      key: "product",
      header: "Produk",
      render: (stock) => (
        <div>
          <button
            type="button"
            className="font-semibold text-primary-700 hover:text-primary-800"
            onClick={() => router.push(`/dashboard/inventory/products/${stock.productId}`)}
          >
            {stock.name}
          </button>
          <p className="mt-1 text-xs text-neutral-500">{stock.sku || "Tanpa SKU"}</p>
        </div>
      )
    },
    {
      key: "category",
      header: "Kategori",
      render: (stock) => stock.categoryName || "Tanpa kategori"
    },
    {
      key: "status",
      header: "Status stok",
      render: (stock) => <StockStatusBadge stock={stock} />
    },
    {
      key: "available",
      header: "Tersedia",
      render: (stock) => <span className="font-semibold text-neutral-950">{stock.quantityAvailable}</span>
    },
    {
      key: "detail",
      header: "Fisik / Reserved",
      render: (stock) => (
        <div>
          <p className="font-medium text-neutral-950">{stock.quantityOnHand} fisik</p>
          <p className="mt-1 text-xs text-neutral-500">{stock.quantityReserved} reserved</p>
        </div>
      )
    },
    {
      key: "threshold",
      header: "Threshold",
      render: (stock) => stock.lowStockThreshold
    },
    {
      key: "updated",
      header: "Update",
      render: (stock) => <span className="text-xs text-neutral-500">{formatDateTime(stock.updatedAt)}</span>
    },
    {
      key: "actions",
      header: "Aksi",
      render: (stock) => (
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => router.push(`/dashboard/inventory/products/${stock.productId}`)}
        >
          Detail
        </Button>
      )
    }
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-neutral-950">Inventori</h1>
        <p className="mt-1 text-sm text-neutral-500">
          Pantau stok tersedia, stok fisik, reserved, dan produk yang mulai menipis.
        </p>
      </div>

      <div className="grid gap-3 rounded-2xl border border-neutral-200 bg-white p-4 lg:grid-cols-[1fr_auto_auto]">
        <Input
          placeholder="Cari nama produk atau SKU..."
          value={query}
          onChange={(event) => {
            setQuery(event.target.value);
            resetPagination();
          }}
        />
        <label className="flex h-10 items-center gap-2 rounded-xl border border-neutral-200 px-3 text-sm text-neutral-700">
          <input
            type="checkbox"
            checked={lowStock}
            onChange={(event) => {
              setLowStock(event.target.checked);
              resetPagination();
            }}
          />
          Stok menipis
        </label>
        <label className="flex h-10 items-center gap-2 rounded-xl border border-neutral-200 px-3 text-sm text-neutral-700">
          <input
            type="checkbox"
            checked={outOfStock}
            onChange={(event) => {
              setOutOfStock(event.target.checked);
              resetPagination();
            }}
          />
          Stok habis
        </label>
      </div>

      {stocks.length === 0 ? (
        <EmptyState
          title="Belum ada stok yang cocok"
          description="Produk dengan stok snapshot akan tampil di sini. Coba ubah kata kunci atau filter stok."
        />
      ) : (
        <>
          <DataTable columns={columns} rows={stocks} getRowKey={(stock) => stock.productId} />
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
            <p className="text-sm text-neutral-500">Menampilkan maksimal {pagination.limit} produk per halaman.</p>
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
