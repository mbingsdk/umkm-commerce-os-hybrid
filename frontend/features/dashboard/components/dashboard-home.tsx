"use client";

import Link from "next/link";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { OrderStatusBadge, PaymentStatusBadge } from "@/features/orders/components/status-badges";
import {
  useDashboardLowStock,
  useDashboardRecentOrders,
  useDashboardSummary
} from "@/features/dashboard/hooks/use-dashboard";
import type { DashboardSummary } from "@/features/dashboard/types";
import { formatDateTime } from "@/lib/format/date";
import { formatRupiah } from "@/lib/format/money";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";

const recentLimit = 5;
const lowStockLimit = 5;

export function DashboardHome() {
  const userPermissions = useTenantStore((state) => state.permissions);
  const canReadSummary = userPermissions.includes(permissions.dashboardReadSummary);
  const canReadRecentOrders = userPermissions.includes(permissions.dashboardReadRecentOrders);
  const canReadLowStock = userPermissions.includes(permissions.dashboardReadLowStock);
  const canReadFinance = userPermissions.includes(permissions.financeReadSummary);

  const summaryQuery = useDashboardSummary(canReadSummary);
  const recentOrdersQuery = useDashboardRecentOrders(recentLimit, canReadRecentOrders);
  const lowStockQuery = useDashboardLowStock(lowStockLimit, canReadLowStock);

  if (!canReadSummary) {
    return (
      <EmptyState
        title="Dashboard belum tersedia"
        description="Role aktifmu belum memiliki izin untuk melihat ringkasan dashboard."
      />
    );
  }

  if (summaryQuery.isPending) {
    return <LoadingState lines={5} />;
  }

  if (summaryQuery.isError) {
    return (
      <ErrorState
        title="Dashboard belum bisa dimuat"
        description="Coba muat ulang ringkasan operasional toko."
        onRetry={() => void summaryQuery.refetch()}
      />
    );
  }

  const summary = summaryQuery.data;

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-neutral-950">Dashboard</h1>
          <p className="mt-1 text-sm text-neutral-500">
            Ringkasan cepat untuk penjualan hari ini, pesanan terbaru, dan stok yang perlu diperhatikan.
          </p>
        </div>
        <Badge tone="primary">Ringkasan operasional</Badge>
      </div>

      <SummaryCards summary={summary} canReadFinance={canReadFinance} />

      <div className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <SalesSourceCard summary={summary} />
        <QuickActions canReadFinance={canReadFinance} />
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <RecentOrdersCard
          canRead={canReadRecentOrders}
          isPending={recentOrdersQuery.isPending}
          isError={recentOrdersQuery.isError}
          onRetry={() => void recentOrdersQuery.refetch()}
          orders={recentOrdersQuery.data ?? []}
        />
        <LowStockCard
          canRead={canReadLowStock}
          isPending={lowStockQuery.isPending}
          isError={lowStockQuery.isError}
          onRetry={() => void lowStockQuery.refetch()}
          products={lowStockQuery.data ?? []}
        />
      </div>
    </div>
  );
}

function SummaryCards({ summary, canReadFinance }: { summary: DashboardSummary; canReadFinance: boolean }) {
  const cards = [
    {
      key: "today_sales",
      label: "Penjualan hari ini",
      value: formatRupiah(summary.todaySales),
      description: "Online + POS sesuai izin role."
    },
    {
      key: "today_order_count",
      label: "Order hari ini",
      value: formatNumber(summary.todayOrderCount),
      description: "Pesanan storefront berbayar/aktif."
    },
    {
      key: "pending_orders_count",
      label: "Order perlu diproses",
      value: formatNumber(summary.pendingOrdersCount),
      description: "Masih menunggu aksi toko."
    },
    {
      key: "low_stock_count",
      label: "Stok menipis",
      value: formatNumber(summary.lowStockCount),
      description: "Produk di bawah threshold."
    },
    {
      key: "expense_today",
      label: "Pengeluaran hari ini",
      value: summary.expenseToday == null ? "—" : formatRupiah(summary.expenseToday),
      description: "Tampil hanya untuk role finance.",
      hidden: !canReadFinance
    },
    {
      key: "net_estimate_today",
      label: "Estimasi net hari ini",
      value: summary.netEstimateToday == null ? "—" : formatRupiah(summary.netEstimateToday),
      description: "Belum menghitung HPP/modal detail.",
      hidden: !canReadFinance
    }
  ].filter((card) => !card.hidden);

  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      {cards.map((card) => (
        <Card key={card.key}>
          <CardHeader className="pb-2">
            <CardDescription>{card.label}</CardDescription>
            <CardTitle className="text-2xl">{card.value}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs leading-5 text-neutral-500">{card.description}</p>
          </CardContent>
        </Card>
      ))}
      {canReadFinance && summary.note ? (
        <Card className="border-amber-200 bg-amber-50 shadow-none sm:col-span-2 xl:col-span-3">
          <CardContent>
            <p className="text-sm font-medium text-amber-900">Catatan finance</p>
            <p className="mt-1 text-sm leading-6 text-amber-800">{summary.note}</p>
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}

function SalesSourceCard({ summary }: { summary: DashboardSummary }) {
  const rows = [
    { label: "Storefront", value: summary.onlineSalesToday ?? 0 },
    { label: "POS", value: summary.posSalesToday ?? 0 }
  ];
  const max = Math.max(...rows.map((row) => row.value), 1);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Penjualan berdasarkan sumber</CardTitle>
        <CardDescription>Visual ringan tanpa dependency chart tambahan.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {rows.every((row) => row.value === 0) ? (
          <p className="rounded-xl border border-dashed border-neutral-200 p-4 text-sm text-neutral-500">
            Belum ada penjualan hari ini.
          </p>
        ) : (
          rows.map((row) => (
            <div key={row.label} className="space-y-2">
              <div className="flex items-center justify-between gap-3 text-sm">
                <span className="font-medium text-neutral-700">{row.label}</span>
                <span className="font-semibold text-neutral-950">{formatRupiah(row.value)}</span>
              </div>
              <div className="h-3 overflow-hidden rounded-full bg-neutral-100">
                <div
                  className="h-full rounded-full bg-primary-600"
                  style={{ width: row.value === 0 ? "0%" : `${Math.max(6, (row.value / max) * 100)}%` }}
                />
              </div>
            </div>
          ))
        )}
      </CardContent>
    </Card>
  );
}

function QuickActions({ canReadFinance }: { canReadFinance: boolean }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Aksi cepat</CardTitle>
        <CardDescription>Shortcut operasional harian yang paling sering dipakai.</CardDescription>
      </CardHeader>
      <CardContent className="grid gap-3 sm:grid-cols-2">
        <LinkButton href="/dashboard/orders">
          Lihat pesanan
        </LinkButton>
        <LinkButton href="/dashboard/inventory">
          Cek inventori
        </LinkButton>
        <LinkButton href="/dashboard/pos">
          Buka POS
        </LinkButton>
        {canReadFinance ? (
          <LinkButton href="/dashboard/finance">
            Lihat finance
          </LinkButton>
        ) : (
          <Button variant="outline" disabled>
            Finance terbatas
          </Button>
        )}
      </CardContent>
    </Card>
  );
}

function LinkButton({ href, children }: { href: string; children: string }) {
  return (
    <Link
      href={href}
      className="inline-flex h-10 items-center justify-center rounded-xl border border-neutral-300 bg-white px-4 text-sm font-semibold text-neutral-900 transition hover:bg-neutral-50"
    >
      {children}
    </Link>
  );
}

function RecentOrdersCard({
  canRead,
  isPending,
  isError,
  onRetry,
  orders
}: {
  canRead: boolean;
  isPending: boolean;
  isError: boolean;
  onRetry: () => void;
  orders: Array<{
    orderId: string;
    orderNumber: string;
    customerName: string;
    totalAmount: number;
    status: string;
    paymentStatus: string;
    createdAt: string;
  }>;
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>Pesanan terbaru</CardTitle>
          <CardDescription>Order terbaru dari tenant aktif.</CardDescription>
        </div>
        <Link className="text-sm font-semibold text-primary-700 hover:text-primary-800" href="/dashboard/orders">
          Semua
        </Link>
      </CardHeader>
      <CardContent>
        {!canRead ? (
          <p className="text-sm text-neutral-500">Role aktifmu belum bisa melihat pesanan terbaru.</p>
        ) : isPending ? (
          <LoadingState lines={2} />
        ) : isError ? (
          <ErrorState title="Pesanan gagal dimuat" description="Coba muat ulang kartu ini." onRetry={onRetry} />
        ) : orders.length === 0 ? (
          <p className="rounded-xl border border-dashed border-neutral-200 p-4 text-sm text-neutral-500">
            Belum ada pesanan terbaru.
          </p>
        ) : (
          <div className="space-y-3">
            {orders.map((order) => (
              <Link
                key={order.orderId}
                href={`/dashboard/orders/${order.orderId}`}
                className="block rounded-xl border border-neutral-200 p-3 transition hover:bg-neutral-50"
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="font-semibold text-neutral-950">{order.orderNumber}</p>
                    <p className="mt-1 text-xs text-neutral-500">
                      {order.customerName} • {formatDateTime(order.createdAt)}
                    </p>
                  </div>
                  <p className="text-sm font-semibold text-neutral-950">{formatRupiah(order.totalAmount)}</p>
                </div>
                <div className="mt-3 flex flex-wrap gap-2">
                  <OrderStatusBadge status={order.status} />
                  <PaymentStatusBadge status={order.paymentStatus} />
                </div>
              </Link>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function LowStockCard({
  canRead,
  isPending,
  isError,
  onRetry,
  products
}: {
  canRead: boolean;
  isPending: boolean;
  isError: boolean;
  onRetry: () => void;
  products: Array<{
    productId: string;
    productName: string;
    sku?: string;
    availableQuantity: number;
    lowStockThreshold: number;
  }>;
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>Stok menipis</CardTitle>
          <CardDescription>Produk yang perlu segera dicek atau restock.</CardDescription>
        </div>
        <Link className="text-sm font-semibold text-primary-700 hover:text-primary-800" href="/dashboard/inventory">
          Semua
        </Link>
      </CardHeader>
      <CardContent>
        {!canRead ? (
          <p className="text-sm text-neutral-500">Role aktifmu belum bisa melihat data stok.</p>
        ) : isPending ? (
          <LoadingState lines={2} />
        ) : isError ? (
          <ErrorState title="Stok gagal dimuat" description="Coba muat ulang kartu ini." onRetry={onRetry} />
        ) : products.length === 0 ? (
          <p className="rounded-xl border border-dashed border-neutral-200 p-4 text-sm text-neutral-500">
            Aman. Tidak ada produk di bawah threshold.
          </p>
        ) : (
          <div className="space-y-3">
            {products.map((product) => (
              <Link
                key={product.productId}
                href={`/dashboard/inventory/products/${product.productId}`}
                className="flex items-center justify-between gap-3 rounded-xl border border-neutral-200 p-3 transition hover:bg-neutral-50"
              >
                <div>
                  <p className="font-semibold text-neutral-950">{product.productName}</p>
                  <p className="mt-1 text-xs text-neutral-500">{product.sku || "Tanpa SKU"}</p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-semibold text-red-700">{product.availableQuantity} tersisa</p>
                  <p className="mt-1 text-xs text-neutral-500">Threshold {product.lowStockThreshold}</p>
                </div>
              </Link>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function formatNumber(value?: number | null) {
  if (value == null) {
    return "—";
  }

  return new Intl.NumberFormat("id-ID").format(value);
}
