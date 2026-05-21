import { Badge } from "@/components/ui/badge";
import { shipmentStatusPresentation } from "@/features/shipments/constants";
import type { ShipmentStatus } from "@/features/shipments/types";

export function ShipmentStatusBadge({ status }: { status: ShipmentStatus | string }) {
  const presentation =
    shipmentStatusPresentation[status as ShipmentStatus] ?? ({ label: status, tone: "neutral" } as const);

  return <Badge tone={presentation.tone}>{presentation.label}</Badge>;
}
