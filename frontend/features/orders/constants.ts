import type { OrderSource, OrderStatus, PaymentStatus } from "@/features/orders/types";

export const orderStatusPresentation: Record<OrderStatus, { label: string; tone: "neutral" | "primary" | "success" | "warning" | "danger" | "info" }> = {
  pending: { label: "Menunggu", tone: "warning" },
  confirmed: { label: "Terkonfirmasi", tone: "info" },
  processing: { label: "Diproses", tone: "primary" },
  ready_to_ship: { label: "Siap dikirim", tone: "info" },
  shipped: { label: "Dikirim", tone: "primary" },
  delivered: { label: "Terkirim", tone: "success" },
  completed: { label: "Selesai", tone: "success" },
  cancelled: { label: "Dibatalkan", tone: "danger" },
  returned: { label: "Retur", tone: "warning" },
  refunded: { label: "Refund", tone: "neutral" }
};

export const paymentStatusPresentation: Record<PaymentStatus, { label: string; tone: "neutral" | "success" | "warning" | "danger" | "info" }> = {
  unpaid: { label: "Belum dibayar", tone: "warning" },
  waiting_confirmation: { label: "Menunggu review", tone: "info" },
  paid: { label: "Lunas", tone: "success" },
  failed: { label: "Gagal", tone: "danger" },
  refunded: { label: "Refund", tone: "neutral" }
};

export const sourceLabels: Record<OrderSource, string> = {
  storefront: "Storefront",
  marketplace_discovery: "Discovery",
  pos: "POS",
  whatsapp_manual: "WhatsApp manual",
  admin_manual: "Admin manual",
  marketplace_sync: "Marketplace sync",
  reseller: "Reseller",
  api_partner: "API partner"
};

export const statusActionLabels: Partial<Record<OrderStatus, string>> = {
  processing: "Tandai diproses",
  ready_to_ship: "Tandai siap kirim",
  shipped: "Tandai dikirim",
  completed: "Tandai selesai"
};

export function nextOperationalStatus(status: OrderStatus): OrderStatus | null {
  if (status === "confirmed") {
    return "processing";
  }
  if (status === "processing") {
    return "ready_to_ship";
  }
  if (status === "ready_to_ship") {
    return "shipped";
  }
  if (status === "shipped") {
    return "completed";
  }
  return null;
}

export function canCancelOrderStatus(status: OrderStatus) {
  return ["pending", "confirmed", "processing", "ready_to_ship"].includes(status);
}
