import { Badge } from "@/components/ui/badge";
import type { StockStatus } from "@/features/storefront/types";

const stockLabels: Record<StockStatus, string> = {
  in_stock: "Tersedia",
  low_stock: "Stok menipis",
  out_of_stock: "Stok habis"
};

const stockTones: Record<StockStatus, "success" | "warning" | "danger"> = {
  in_stock: "success",
  low_stock: "warning",
  out_of_stock: "danger"
};

export function StockBadge({ status }: { status: StockStatus }) {
  return <Badge tone={stockTones[status]}>{stockLabels[status]}</Badge>;
}
