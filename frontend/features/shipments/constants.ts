import type { ShipmentStatus } from "@/features/shipments/types";

export const shipmentStatusOptions: Array<{ value: ShipmentStatus; label: string }> = [
  { value: "pending", label: "Menunggu" },
  { value: "ready_for_pickup", label: "Siap diambil" },
  { value: "picked_up", label: "Sudah diambil" },
  { value: "on_delivery", label: "Dalam pengiriman" },
  { value: "delivered", label: "Terkirim" },
  { value: "failed", label: "Gagal kirim" },
  { value: "cancelled", label: "Dibatalkan" }
];

export const shipmentStatusPresentation: Record<ShipmentStatus, { label: string; tone: "neutral" | "primary" | "success" | "warning" | "danger" | "info" }> = {
  pending: { label: "Menunggu", tone: "warning" },
  ready_for_pickup: { label: "Siap diambil", tone: "primary" },
  picked_up: { label: "Sudah diambil", tone: "info" },
  on_delivery: { label: "Dalam pengiriman", tone: "info" },
  delivered: { label: "Terkirim", tone: "success" },
  failed: { label: "Gagal kirim", tone: "danger" },
  cancelled: { label: "Dibatalkan", tone: "neutral" }
};

export const courierTypeLabels: Record<string, string> = {
  internal: "Kurir internal",
  manual: "Manual / pihak ketiga"
};

export function shipmentStatusLabel(status: string) {
  return shipmentStatusPresentation[status as ShipmentStatus]?.label ?? status;
}
