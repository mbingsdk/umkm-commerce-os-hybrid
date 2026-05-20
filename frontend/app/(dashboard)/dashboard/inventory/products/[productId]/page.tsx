"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  adjustProductStock,
  updateStockThreshold
} from "@/features/inventory/api/inventory.api";
import { StockAdjustmentDialog } from "@/features/inventory/components/stock-adjustment-dialog";
import { MovementTypeBadge, StockStatusBadge } from "@/features/inventory/components/stock-status-badge";
import { ThresholdDialog } from "@/features/inventory/components/threshold-dialog";
import { useInventoryStocks, useProductMovements } from "@/features/inventory/hooks/use-inventory";
import type { InventoryStock, StockMovement } from "@/features/inventory/types";
import { formatDateTime } from "@/lib/format/date";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";
import { useToastStore } from "@/lib/stores/toast.store";

const movementLimit = 20;

export default function InventoryProductDetailPage() {
  const params = useParams<{ productId: string }>();
  const router = useRouter();
  const queryClient = useQueryClient();
  const pushToast = useToastStore((state) => state.pushToast);
  const tenantId = useTenantStore((state) => state.selectedTenantId);
  const userPermissions = useTenantStore((state) => state.permissions);
  const productId = params.productId;
  const canRead = userPermissions.includes(permissions.inventoryRead);
  const canReadMovement = userPermissions.includes(permissions.inventoryReadMovement);
  const canAdjust = userPermissions.includes(permissions.inventoryAdjust);
  const canUpdateThreshold = userPermissions.includes(permissions.inventoryUpdateThreshold);
  const [movementCursor, setMovementCursor] = useState<string | null>(null);
  const [movementCursorStack, setMovementCursorStack] = useState<string[]>([]);
  const [adjustOpen, setAdjustOpen] = useState(false);
  const [thresholdOpen, setThresholdOpen] = useState(false);

  const stockQuery = useInventoryStocks({ limit: 100 }, canRead);
  const movementFilters = useMemo(() => ({ cursor: movementCursor, limit: movementLimit }), [movementCursor]);
  const movementsQuery = useProductMovements(productId, movementFilters, canReadMovement);

  const stock = useMemo(
    () => stockQuery.data?.stocks.find((item) => item.productId === productId) ?? null,
    [productId, stockQuery.data?.stocks]
  );

  const invalidateInventory = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["inventory-stocks", tenantId] }),
      queryClient.invalidateQueries({ queryKey: ["inventory-movements", tenantId, productId] })
    ]);
  };

  const adjustMutation = useMutation({
    mutationFn: (values: Parameters<typeof adjustProductStock>[1]) => adjustProductStock(productId, values),
    onSuccess: async () => {
      await invalidateInventory();
      setAdjustOpen(false);
      pushToast({ tone: "success", title: "Stok berhasil disesuaikan" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Penyesuaian stok gagal", description: error.message });
    }
  });

  const thresholdMutation = useMutation({
    mutationFn: (lowStockThreshold: number) => updateStockThreshold(productId, { lowStockThreshold }),
    onSuccess: async () => {
      await invalidateInventory();
      setThresholdOpen(false);
      pushToast({ tone: "success", title: "Threshold stok diperbarui" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Threshold gagal diperbarui", description: error.message });
    }
  });

  if (!canRead) {
    return (
      <EmptyState
        title="Akses inventori belum tersedia"
        description="Role aktifmu belum memiliki izin untuk membuka detail stok."
        action={
          <Button type="button" variant="outline" onClick={() => router.push("/dashboard/inventory")}>
            Kembali ke inventori
          </Button>
        }
      />
    );
  }

  if (stockQuery.isPending || (canReadMovement && movementsQuery.isPending)) {
    return <LoadingState lines={5} />;
  }

  if (stockQuery.isError || (canReadMovement && movementsQuery.isError)) {
    return (
      <ErrorState
        title="Detail stok belum bisa dimuat"
        description="Coba muat ulang sebelum melakukan penyesuaian stok."
        onRetry={() => {
          void stockQuery.refetch();
          void movementsQuery.refetch();
        }}
      />
    );
  }

  const movements = canReadMovement ? movementsQuery.data?.movements ?? [] : [];
  const movementPagination = canReadMovement
    ? movementsQuery.data?.pagination ?? { limit: movementLimit, nextCursor: null, hasMore: false }
    : { limit: movementLimit, nextCursor: null, hasMore: false };
  const movementColumns: Array<DataTableColumn<StockMovement>> = [
    {
      key: "type",
      header: "Tipe",
      render: (movement) => <MovementTypeBadge type={movement.movementType} />
    },
    {
      key: "quantity",
      header: "Qty",
      render: (movement) => (
        <span className={movement.quantity < 0 ? "font-semibold text-red-700" : "font-semibold text-green-700"}>
          {movement.quantity > 0 ? `+${movement.quantity}` : movement.quantity}
        </span>
      )
    },
    {
      key: "balance",
      header: "Saldo akhir",
      render: (movement) => movement.balanceAfter
    },
    {
      key: "reason",
      header: "Alasan",
      render: (movement) => (
        <div>
          <p className="font-medium text-neutral-950">{movement.reason || "-"}</p>
          {movement.note ? <p className="mt-1 text-xs text-neutral-500">{movement.note}</p> : null}
        </div>
      )
    },
    {
      key: "actor",
      header: "Aktor",
      render: (movement) => movement.createdBy?.name || movement.createdBy?.id || "Sistem"
    },
    {
      key: "created",
      header: "Waktu",
      render: (movement) => <span className="text-xs text-neutral-500">{formatDateTime(movement.createdAt)}</span>
    }
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="-ml-3"
            onClick={() => router.push("/dashboard/inventory")}
          >
            ← Kembali
          </Button>
          <h1 className="mt-2 text-2xl font-semibold text-neutral-950">{stock?.name ?? "Detail stok produk"}</h1>
          <p className="mt-1 text-sm text-neutral-500">
            {stock?.sku ? `SKU ${stock.sku}` : "Pantau pergerakan stok dan threshold produk."}
          </p>
          {stock ? (
            <div className="mt-3">
              <StockStatusBadge stock={stock} />
            </div>
          ) : null}
        </div>

        <div className="flex flex-wrap gap-2">
          <Button type="button" variant="outline" disabled={!canUpdateThreshold || !stock} onClick={() => setThresholdOpen(true)}>
            Ubah threshold
          </Button>
          <Button type="button" disabled={!canAdjust || !stock} onClick={() => setAdjustOpen(true)}>
            Sesuaikan stok
          </Button>
        </div>
      </div>

      {!stock ? (
        <EmptyState
          title="Ringkasan stok belum ditemukan"
          description="Produk ini tidak muncul di halaman pertama daftar stok. Buka dari tabel inventori agar ringkasan stok bisa tampil."
        />
      ) : (
        <StockSummary stock={stock} />
      )}

      <Card>
        <CardHeader>
          <CardTitle>Riwayat movement stok</CardTitle>
          <CardDescription>Stock movement adalah sumber riwayat operasional inventori.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {!canReadMovement ? (
            <EmptyState
              title="Akses movement belum tersedia"
              description="Role aktifmu belum memiliki izin untuk melihat riwayat movement stok."
            />
          ) : movements.length === 0 ? (
            <EmptyState
              title="Belum ada movement"
              description="Perubahan stok seperti stok awal, reserved, POS sale, dan adjustment akan tampil di sini."
            />
          ) : (
            <>
              <DataTable columns={movementColumns} rows={movements} getRowKey={(movement) => movement.id} />
              <div className="flex items-center justify-between gap-3">
                <Button
                  type="button"
                  variant="outline"
                  disabled={movementCursorStack.length === 0}
                  onClick={() => {
                    const previous = movementCursorStack[movementCursorStack.length - 1];
                    setMovementCursor(previous || null);
                    setMovementCursorStack((items) => items.slice(0, -1));
                  }}
                >
                  Sebelumnya
                </Button>
                <p className="text-sm text-neutral-500">
                  Menampilkan maksimal {movementPagination.limit} movement per halaman.
                </p>
                <Button
                  type="button"
                  variant="outline"
                  disabled={!movementPagination.hasMore || !movementPagination.nextCursor}
                  onClick={() => {
                    setMovementCursorStack((items) => [...items, movementCursor ?? ""]);
                    setMovementCursor(movementPagination.nextCursor ?? null);
                  }}
                >
                  Berikutnya
                </Button>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      <StockAdjustmentDialog
        key={adjustOpen ? "stock-adjustment-open" : "stock-adjustment-closed"}
        open={adjustOpen}
        stock={stock}
        isSubmitting={adjustMutation.isPending}
        error={adjustMutation.isError ? adjustMutation.error.message : undefined}
        onClose={() => setAdjustOpen(false)}
        onSubmit={(values) => adjustMutation.mutate(values)}
      />

      <ThresholdDialog
        key={`${stock?.productId ?? productId}-${stock?.lowStockThreshold ?? 0}-${thresholdOpen ? "open" : "closed"}`}
        open={thresholdOpen}
        stock={stock}
        isSubmitting={thresholdMutation.isPending}
        error={thresholdMutation.isError ? thresholdMutation.error.message : undefined}
        onClose={() => setThresholdOpen(false)}
        onSubmit={(value) => thresholdMutation.mutate(value)}
      />
    </div>
  );
}

function StockSummary({ stock }: { stock: InventoryStock }) {
  return (
    <div className="grid gap-4 md:grid-cols-4">
      <SummaryCard title="Stok tersedia" value={stock.quantityAvailable} helper="Bisa dijual sekarang" />
      <SummaryCard title="Stok fisik" value={stock.quantityOnHand} helper="Total on hand" />
      <SummaryCard title="Reserved" value={stock.quantityReserved} helper="Terkunci untuk order" />
      <SummaryCard title="Threshold" value={stock.lowStockThreshold} helper="Batas stok menipis" />
    </div>
  );
}

function SummaryCard({ title, value, helper }: { title: string; value: number; helper: string }) {
  return (
    <Card>
      <CardContent>
        <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">{title}</p>
        <p className="mt-2 text-3xl font-semibold text-neutral-950">{value}</p>
        <p className="mt-1 text-sm text-neutral-500">{helper}</p>
      </CardContent>
    </Card>
  );
}
