import { Badge } from "@/components/ui/badge";
import {
  orderStatusPresentation,
  paymentStatusPresentation,
  sourceLabels
} from "@/features/orders/constants";
import type { OrderSource, OrderStatus, PaymentStatus } from "@/features/orders/types";

export function OrderStatusBadge({ status }: { status: OrderStatus | string }) {
  const presentation =
    orderStatusPresentation[status as OrderStatus] ?? ({ label: status, tone: "neutral" } as const);

  return <Badge tone={presentation.tone}>{presentation.label}</Badge>;
}

export function PaymentStatusBadge({ status }: { status: PaymentStatus | string }) {
  const presentation =
    paymentStatusPresentation[status as PaymentStatus] ?? ({ label: status, tone: "neutral" } as const);

  return <Badge tone={presentation.tone}>{presentation.label}</Badge>;
}

export function OrderSourceLabel({ source }: { source: OrderSource | string }) {
  return <span>{sourceLabels[source as OrderSource] ?? source}</span>;
}
