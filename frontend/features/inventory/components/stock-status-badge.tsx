import { Badge } from "@/components/ui/badge";
import type { InventoryStock } from "@/features/inventory/types";

export function StockStatusBadge({ stock }: { stock: Pick<InventoryStock, "isLowStock" | "isOutOfStock"> }) {
  if (stock.isOutOfStock) {
    return <Badge tone="danger">Stok habis</Badge>;
  }

  if (stock.isLowStock) {
    return <Badge tone="warning">Stok menipis</Badge>;
  }

  return <Badge tone="success">Stok aman</Badge>;
}

export function MovementTypeBadge({ type }: { type: string }) {
  const presentation = movementTypePresentation[type] ?? { label: type, tone: "neutral" as const };

  return <Badge tone={presentation.tone}>{presentation.label}</Badge>;
}

const movementTypePresentation: Record<
  string,
  { label: string; tone: "neutral" | "primary" | "success" | "warning" | "danger" | "info" }
> = {
  initial: { label: "Stok awal", tone: "info" },
  reserved: { label: "Reserved", tone: "warning" },
  released: { label: "Dilepas", tone: "success" },
  reservation_released: { label: "Reservasi dilepas", tone: "success" },
  cancelled: { label: "Dibatalkan", tone: "neutral" },
  adjustment_in: { label: "Penyesuaian masuk", tone: "success" },
  adjustment_out: { label: "Penyesuaian keluar", tone: "danger" },
  pos_sale: { label: "POS sale", tone: "primary" }
};
