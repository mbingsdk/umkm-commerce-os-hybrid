"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";
import { useState } from "react";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { updateShipmentStatus } from "@/features/shipments/api/shipments.api";
import { courierTypeLabels, shipmentStatusLabel } from "@/features/shipments/constants";
import { ShipmentStatusBadge } from "@/features/shipments/components/shipment-status-badge";
import { UpdateShipmentStatusDialog } from "@/features/shipments/components/update-shipment-status-dialog";
import { useShipmentDetail } from "@/features/shipments/hooks/use-shipments";
import type { ShipmentStatusLog } from "@/features/shipments/types";
import { queryKeys } from "@/lib/api/query-keys";
import { formatDateTime } from "@/lib/format/date";
import { formatRupiah } from "@/lib/format/money";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";
import { useToastStore } from "@/lib/stores/toast.store";

export function ShipmentDetailPage() {
  const params = useParams<{ shipmentId: string }>();
  const router = useRouter();
  const queryClient = useQueryClient();
  const pushToast = useToastStore((state) => state.pushToast);
  const tenantId = useTenantStore((state) => state.selectedTenantId);
  const userPermissions = useTenantStore((state) => state.permissions);
  const canRead = userPermissions.includes(permissions.shipmentRead);
  const canUpdateStatus = userPermissions.includes(permissions.shipmentUpdateStatus);
  const shipmentId = params.shipmentId;
  const [statusDialogOpen, setStatusDialogOpen] = useState(false);
  const detailQuery = useShipmentDetail(shipmentId, canRead);

  const updateStatusMutation = useMutation({
    mutationFn: (values: Parameters<typeof updateShipmentStatus>[1]) => updateShipmentStatus(shipmentId, values),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.shipment(tenantId, shipmentId) }),
        queryClient.invalidateQueries({ queryKey: ["shipments", tenantId] }),
        queryClient.invalidateQueries({ queryKey: ["orders", tenantId] }),
        queryClient.invalidateQueries({ queryKey: ["dashboard-recent-orders", tenantId] })
      ]);
      setStatusDialogOpen(false);
      pushToast({ tone: "success", title: "Status pengiriman diperbarui" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Status gagal diperbarui", description: error.message });
    }
  });

  if (!canRead) {
    return (
      <EmptyState
        title="Akses detail pengiriman belum tersedia"
        description="Role aktifmu belum memiliki izin untuk membuka detail shipment."
        action={
          <Button type="button" variant="outline" onClick={() => router.push("/dashboard/shipments")}>
            Kembali ke pengiriman
          </Button>
        }
      />
    );
  }

  if (detailQuery.isPending) {
    return <LoadingState lines={5} />;
  }

  if (detailQuery.isError) {
    return (
      <ErrorState
        title="Detail pengiriman belum bisa dimuat"
        description="Coba muat ulang sebelum mengubah status shipment."
        onRetry={() => void detailQuery.refetch()}
      />
    );
  }

  const { shipment, timeline } = detailQuery.data;
  const canChangeStatus = canUpdateStatus && !["delivered", "cancelled"].includes(shipment.status);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <Button type="button" variant="ghost" size="sm" className="-ml-3" onClick={() => router.push("/dashboard/shipments")}>
            ← Kembali
          </Button>
          <h1 className="mt-2 text-2xl font-semibold text-neutral-950">Pengiriman {shipment.orderNumber}</h1>
          <div className="mt-3 flex flex-wrap gap-2">
            <ShipmentStatusBadge status={shipment.status} />
            <span className="rounded-full bg-neutral-100 px-2.5 py-1 text-xs font-semibold text-neutral-700">
              {courierTypeLabels[shipment.courierType] ?? shipment.courierType}
            </span>
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          <Button type="button" variant="outline" onClick={() => router.push(`/dashboard/orders/${shipment.orderId}`)}>
            Buka order
          </Button>
          <Button type="button" disabled={!canChangeStatus} onClick={() => setStatusDialogOpen(true)}>
            Update status
          </Button>
        </div>
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <SummaryCard title="Status" value={shipmentStatusLabel(shipment.status)} helper={formatDateTime(shipment.updatedAt)} />
        <SummaryCard title="Ongkir" value={formatRupiah(shipment.shippingCost)} helper="Biaya pengiriman tercatat" />
        <SummaryCard title="Resi" value={shipment.trackingNumber || "Belum ada"} helper={shipment.courierName || "Kurir belum diberi nama"} />
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
        <Card>
          <CardHeader>
            <CardTitle>Timeline pengiriman</CardTitle>
            <CardDescription>Riwayat status shipment yang aman untuk operasional toko.</CardDescription>
          </CardHeader>
          <CardContent>
            <ShipmentTimeline logs={timeline} />
          </CardContent>
        </Card>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Ringkasan order</CardTitle>
              <CardDescription>Shipment ini terhubung ke order berikut.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 text-sm text-neutral-600">
              <div>
                <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">Nomor order</p>
                <p className="mt-1 font-semibold text-neutral-950">{shipment.orderNumber}</p>
              </div>
              <Button type="button" variant="outline" onClick={() => router.push(`/dashboard/orders/${shipment.orderId}`)}>
                Lihat detail order
              </Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Info kurir</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm text-neutral-600">
              <InfoRow label="Tipe" value={courierTypeLabels[shipment.courierType] ?? shipment.courierType} />
              <InfoRow label="Kurir" value={shipment.courierName || "Belum diisi"} />
              <InfoRow label="Petugas" value={shipment.assignedToName || "Belum ditugaskan"} />
              <InfoRow label="HP petugas" value={shipment.assignedToPhone || "Belum diisi"} />
              {shipment.note ? <InfoRow label="Catatan internal" value={shipment.note} /> : null}
            </CardContent>
          </Card>
        </div>
      </div>

      <UpdateShipmentStatusDialog
        open={statusDialogOpen}
        currentStatus={shipment.status}
        isSubmitting={updateStatusMutation.isPending}
        error={updateStatusMutation.error?.message}
        onClose={() => setStatusDialogOpen(false)}
        onSubmit={(values) => updateStatusMutation.mutate(values)}
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

function ShipmentTimeline({ logs }: { logs: ShipmentStatusLog[] }) {
  if (logs.length === 0) {
    return <p className="text-sm text-neutral-500">Belum ada timeline pengiriman.</p>;
  }

  return (
    <div className="space-y-4">
      {logs.map((log) => (
        <div key={log.id} className="border-l-2 border-primary-200 pl-4">
          <p className="text-sm font-semibold text-neutral-950">
            {log.fromStatus ? `${shipmentStatusLabel(log.fromStatus)} → ` : ""}
            {shipmentStatusLabel(log.toStatus)}
          </p>
          <p className="mt-1 text-xs text-neutral-500">{formatDateTime(log.createdAt)}</p>
          {log.note ? <p className="mt-2 text-sm leading-6 text-neutral-600">{log.note}</p> : null}
        </div>
      ))}
    </div>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">{label}</p>
      <p className="mt-1 text-neutral-700">{value}</p>
    </div>
  );
}
