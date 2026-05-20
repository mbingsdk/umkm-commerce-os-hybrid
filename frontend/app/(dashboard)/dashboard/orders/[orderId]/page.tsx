"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";
import { useState } from "react";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  cancelOrder,
  confirmPayment,
  rejectPayment,
  updateOrderStatus
} from "@/features/orders/api/orders.api";
import { canCancelOrderStatus, nextOperationalStatus, statusActionLabels } from "@/features/orders/constants";
import { CancelOrderDialog } from "@/features/orders/components/cancel-order-dialog";
import { PaymentReviewDialog } from "@/features/orders/components/payment-review-dialog";
import {
  OrderSourceLabel,
  OrderStatusBadge,
  PaymentStatusBadge
} from "@/features/orders/components/status-badges";
import { useOrderDetail, usePaymentConfirmations } from "@/features/orders/hooks/use-orders";
import type { OrderItem, PaymentConfirmation, ReservationSummary, StatusLog } from "@/features/orders/types";
import { formatDateTime } from "@/lib/format/date";
import { formatRupiah } from "@/lib/format/money";
import { queryKeys } from "@/lib/api/query-keys";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";
import { useToastStore } from "@/lib/stores/toast.store";

type PaymentReviewMode = "confirm" | "reject";

export default function OrderDetailPage() {
  const params = useParams<{ orderId: string }>();
  const router = useRouter();
  const queryClient = useQueryClient();
  const pushToast = useToastStore((state) => state.pushToast);
  const orderId = params.orderId;
  const tenantId = useTenantStore((state) => state.selectedTenantId);
  const userPermissions = useTenantStore((state) => state.permissions);
  const canReadDetail = userPermissions.includes(permissions.orderReadDetail);
  const canUpdateStatus = userPermissions.includes(permissions.orderUpdateStatus);
  const canUpdatePayment = userPermissions.includes(permissions.orderUpdatePaymentStatus);
  const canCancel = userPermissions.includes(permissions.orderCancel);
  const [cancelOpen, setCancelOpen] = useState(false);
  const [paymentReviewMode, setPaymentReviewMode] = useState<PaymentReviewMode | null>(null);
  const detailQuery = useOrderDetail(orderId, canReadDetail);
  const confirmationsQuery = usePaymentConfirmations(orderId, canReadDetail);

  const invalidateOrderQueries = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.order(tenantId, orderId) }),
      queryClient.invalidateQueries({ queryKey: queryKeys.paymentConfirmations(tenantId, orderId) }),
      queryClient.invalidateQueries({ queryKey: ["orders", tenantId] }),
      queryClient.invalidateQueries({ queryKey: ["dashboard-summary", tenantId] }),
      queryClient.invalidateQueries({ queryKey: ["dashboard-recent-orders", tenantId] }),
      queryClient.invalidateQueries({ queryKey: ["dashboard-low-stock", tenantId] }),
      queryClient.invalidateQueries({ queryKey: ["finance-summary", tenantId] }),
      queryClient.invalidateQueries({ queryKey: ["finance-daily-report", tenantId] }),
      queryClient.invalidateQueries({ queryKey: ["finance-monthly-report", tenantId] }),
      queryClient.invalidateQueries({ queryKey: ["inventory-stocks", tenantId] })
    ]);
  };

  const updateStatusMutation = useMutation({
    mutationFn: (status: NonNullable<ReturnType<typeof nextOperationalStatus>>) =>
      updateOrderStatus(orderId, {
        status,
        note: "Updated from dashboard order detail"
      }),
    onSuccess: async () => {
      await invalidateOrderQueries();
      pushToast({ tone: "success", title: "Status pesanan diperbarui" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Status gagal diperbarui", description: error.message });
    }
  });

  const cancelMutation = useMutation({
    mutationFn: (values: { reason: string; note?: string }) => cancelOrder(orderId, values),
    onSuccess: async () => {
      await invalidateOrderQueries();
      setCancelOpen(false);
      pushToast({ tone: "success", title: "Pesanan berhasil dibatalkan", description: "Reservasi stok dilepas oleh backend." });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Pesanan gagal dibatalkan", description: error.message });
    }
  });

  const confirmPaymentMutation = useMutation({
    mutationFn: (values: { paymentConfirmationId?: string; note?: string }) => confirmPayment(orderId, values),
    onSuccess: async () => {
      await invalidateOrderQueries();
      setPaymentReviewMode(null);
      pushToast({ tone: "success", title: "Pembayaran dikonfirmasi" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Konfirmasi pembayaran gagal", description: error.message });
    }
  });

  const rejectPaymentMutation = useMutation({
    mutationFn: (values: { paymentConfirmationId?: string; note?: string }) => rejectPayment(orderId, values),
    onSuccess: async () => {
      await invalidateOrderQueries();
      setPaymentReviewMode(null);
      pushToast({ tone: "success", title: "Konfirmasi pembayaran ditolak" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Penolakan pembayaran gagal", description: error.message });
    }
  });

  if (!canReadDetail) {
    return (
      <EmptyState
        title="Akses detail pesanan belum tersedia"
        description="Role aktifmu belum memiliki izin untuk membuka detail pesanan."
        action={
          <Button type="button" variant="outline" onClick={() => router.push("/dashboard/orders")}>
            Kembali ke pesanan
          </Button>
        }
      />
    );
  }

  if (detailQuery.isPending || confirmationsQuery.isPending) {
    return <LoadingState lines={5} />;
  }

  if (detailQuery.isError || confirmationsQuery.isError) {
    return (
      <ErrorState
        title="Detail pesanan belum bisa dimuat"
        description="Coba muat ulang sebelum memproses pesanan."
        onRetry={() => {
          void detailQuery.refetch();
          void confirmationsQuery.refetch();
        }}
      />
    );
  }

  const detail = detailQuery.data;
  const order = detail.order;
  const confirmations = confirmationsQuery.data ?? [];
  const pendingConfirmations = confirmations.filter((confirmation) => confirmation.status === "pending");
  const nextStatus = nextOperationalStatus(order.status);
  const canAdvanceStatus = canUpdateStatus && !!nextStatus;
  const canReviewPayment =
    canUpdatePayment && order.paymentStatus !== "paid" && order.status !== "cancelled" && pendingConfirmations.length > 0;
  const canOpenCancel = canCancel && canCancelOrderStatus(order.status);

  const itemColumns: Array<DataTableColumn<OrderItem>> = [
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
    { key: "unitPrice", header: "Harga", render: (item) => formatRupiah(item.unitPrice) },
    { key: "subtotal", header: "Subtotal", render: (item) => formatRupiah(item.subtotal) }
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <Button type="button" variant="ghost" size="sm" className="-ml-3" onClick={() => router.push("/dashboard/orders")}>
            ← Kembali
          </Button>
          <h1 className="mt-2 text-2xl font-semibold text-neutral-950">{order.orderNumber}</h1>
          <div className="mt-3 flex flex-wrap gap-2">
            <OrderStatusBadge status={order.status} />
            <PaymentStatusBadge status={order.paymentStatus} />
            <span className="rounded-full bg-neutral-100 px-2.5 py-1 text-xs font-semibold text-neutral-700">
              <OrderSourceLabel source={order.source} />
            </span>
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          {nextStatus ? (
            <Button
              type="button"
              variant="outline"
              disabled={!canAdvanceStatus}
              isLoading={updateStatusMutation.isPending}
              onClick={() => updateStatusMutation.mutate(nextStatus)}
            >
              {statusActionLabels[nextStatus]}
            </Button>
          ) : null}
          <Button
            type="button"
            variant="outline"
            disabled={!canReviewPayment}
            onClick={() => setPaymentReviewMode("confirm")}
          >
            Konfirmasi pembayaran
          </Button>
          <Button
            type="button"
            variant="outline"
            disabled={!canReviewPayment}
            onClick={() => setPaymentReviewMode("reject")}
          >
            Tolak pembayaran
          </Button>
          <Button type="button" variant="danger" disabled={!canOpenCancel} onClick={() => setCancelOpen(true)}>
            Batalkan pesanan
          </Button>
        </div>
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <SummaryCard title="Total order" value={formatRupiah(order.grandTotal)} helper={formatDateTime(order.createdAt)} />
        <SummaryCard title="Pembayaran" value={paymentLabel(order.paymentStatus)} helper={order.paidAt ? `Dibayar ${formatDateTime(order.paidAt)}` : "Manual transfer MVP"} />
        <SummaryCard title="Customer" value={detail.customer.name} helper={detail.customer.phone} />
      </div>

      <div className="grid gap-6 lg:grid-cols-[2fr_1fr]">
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Item pesanan</CardTitle>
              <CardDescription>Harga dan nama produk berasal dari snapshot order backend.</CardDescription>
            </CardHeader>
            <CardContent>
              <DataTable columns={itemColumns} rows={detail.items} getRowKey={(item) => item.id} />
            </CardContent>
          </Card>

          <PaymentConfirmationsCard confirmations={confirmations} />
          <TimelineCard logs={detail.statusLogs} />
        </div>

        <div className="space-y-6">
          <CustomerCard
            name={detail.customer.name}
            phone={detail.customer.phone}
            email={detail.customer.email}
            address={[
              detail.shippingAddress.address,
              detail.shippingAddress.city,
              detail.shippingAddress.province,
              detail.shippingAddress.postalCode
            ]
              .filter(Boolean)
              .join(", ")}
          />
          <TotalsCard
            subtotal={detail.paymentSummary.subtotal}
            discount={detail.paymentSummary.discountTotal}
            shipping={detail.paymentSummary.shippingCost}
            tax={detail.paymentSummary.taxTotal}
            grandTotal={detail.paymentSummary.grandTotal}
          />
          <ReservationsCard reservations={detail.stockReservations} />
        </div>
      </div>

      <CancelOrderDialog
        open={cancelOpen}
        isSubmitting={cancelMutation.isPending}
        error={cancelMutation.isError ? cancelMutation.error.message : undefined}
        onClose={() => setCancelOpen(false)}
        onSubmit={(values) => cancelMutation.mutate(values)}
      />

      <PaymentReviewDialog
        open={paymentReviewMode !== null}
        mode={paymentReviewMode ?? "confirm"}
        confirmations={confirmations}
        isSubmitting={confirmPaymentMutation.isPending || rejectPaymentMutation.isPending}
        error={
          confirmPaymentMutation.isError
            ? confirmPaymentMutation.error.message
            : rejectPaymentMutation.isError
              ? rejectPaymentMutation.error.message
              : undefined
        }
        onClose={() => setPaymentReviewMode(null)}
        onSubmit={(values) => {
          if (paymentReviewMode === "confirm") {
            confirmPaymentMutation.mutate(values);
            return;
          }
          rejectPaymentMutation.mutate(values);
        }}
      />
    </div>
  );
}

function SummaryCard({ title, value, helper }: { title: string; value: string; helper: string }) {
  return (
    <Card>
      <CardContent>
        <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">{title}</p>
        <p className="mt-2 text-xl font-semibold text-neutral-950">{value}</p>
        <p className="mt-1 text-sm text-neutral-500">{helper}</p>
      </CardContent>
    </Card>
  );
}

function CustomerCard({ name, phone, email, address }: { name: string; phone: string; email?: string; address: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Customer & alamat</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3 text-sm text-neutral-600">
        <div>
          <p className="font-semibold text-neutral-950">{name}</p>
          <p>{phone}</p>
          {email ? <p>{email}</p> : null}
        </div>
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">Alamat kirim</p>
          <p className="mt-1 leading-6">{address || "Belum ada alamat pengiriman."}</p>
        </div>
      </CardContent>
    </Card>
  );
}

function TotalsCard({
  subtotal,
  discount,
  shipping,
  tax,
  grandTotal
}: {
  subtotal: number;
  discount: number;
  shipping: number;
  tax: number;
  grandTotal: number;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Ringkasan pembayaran</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2 text-sm">
        <MoneyRow label="Subtotal" value={subtotal} />
        <MoneyRow label="Diskon" value={discount} />
        <MoneyRow label="Ongkir" value={shipping} />
        <MoneyRow label="Pajak" value={tax} />
        <div className="border-t border-neutral-100 pt-3">
          <MoneyRow label="Total" value={grandTotal} strong />
        </div>
      </CardContent>
    </Card>
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

function ReservationsCard({ reservations }: { reservations: ReservationSummary[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Reservasi stok</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {reservations.length === 0 ? (
          <p className="text-sm text-neutral-500">Tidak ada reservasi stok untuk order ini.</p>
        ) : (
          reservations.map((reservation) => (
            <div key={reservation.status} className="flex items-center justify-between rounded-xl bg-neutral-50 p-3 text-sm">
              <div>
                <p className="font-medium text-neutral-950">{reservation.status}</p>
                <p className="text-xs text-neutral-500">{reservation.count} baris reservasi</p>
              </div>
              <p className="font-semibold text-neutral-950">{reservation.quantity}</p>
            </div>
          ))
        )}
      </CardContent>
    </Card>
  );
}

function PaymentConfirmationsCard({ confirmations }: { confirmations: PaymentConfirmation[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Konfirmasi pembayaran</CardTitle>
        <CardDescription>Data ini berasal dari bukti/manual confirmation customer.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {confirmations.length === 0 ? (
          <p className="text-sm text-neutral-500">Belum ada konfirmasi pembayaran.</p>
        ) : (
          confirmations.map((confirmation) => (
            <div key={confirmation.id} className="rounded-xl border border-neutral-200 p-4 text-sm">
              <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <p className="font-semibold text-neutral-950">{confirmation.payerName}</p>
                  <p className="mt-1 text-neutral-500">
                    {confirmation.bankName} • {formatDateTime(confirmation.transferDate)}
                  </p>
                </div>
                <p className="font-semibold text-neutral-950">{formatRupiah(confirmation.transferAmount)}</p>
              </div>
              {confirmation.note ? <p className="mt-3 text-neutral-600">{confirmation.note}</p> : null}
              {confirmation.proofUrl ? (
                <a
                  className="mt-3 inline-block text-sm font-semibold text-primary-700 hover:text-primary-800"
                  href={confirmation.proofUrl}
                  target="_blank"
                  rel="noreferrer"
                >
                  Lihat bukti pembayaran
                </a>
              ) : null}
              <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-neutral-400">
                Status: {confirmation.status}
              </p>
            </div>
          ))
        )}
      </CardContent>
    </Card>
  );
}

function TimelineCard({ logs }: { logs: StatusLog[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Timeline status</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {logs.length === 0 ? (
          <p className="text-sm text-neutral-500">Belum ada timeline status.</p>
        ) : (
          logs.map((log) => (
            <div key={log.id} className="border-l-2 border-primary-200 pl-4">
              <p className="text-sm font-semibold text-neutral-950">
                {log.fromStatus ? `${log.fromStatus} → ` : ""}
                {log.toStatus}
              </p>
              <p className="mt-1 text-xs text-neutral-500">{formatDateTime(log.createdAt)}</p>
              {log.note ? <p className="mt-2 text-sm leading-6 text-neutral-600">{log.note}</p> : null}
            </div>
          ))
        )}
      </CardContent>
    </Card>
  );
}

function paymentLabel(status: string) {
  if (status === "paid") {
    return "Lunas";
  }
  if (status === "waiting_confirmation") {
    return "Menunggu review";
  }
  if (status === "failed") {
    return "Gagal";
  }
  if (status === "refunded") {
    return "Refund";
  }
  return "Belum dibayar";
}
