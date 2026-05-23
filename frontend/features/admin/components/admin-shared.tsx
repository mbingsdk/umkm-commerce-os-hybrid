import type { ReactNode } from "react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { formatDateTime } from "@/lib/format/date";

export function AdminPageHeader({
  title,
  description,
  action
}: {
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
      <div>
        <h1 className="text-2xl font-semibold text-neutral-950">{title}</h1>
        <p className="mt-1 max-w-2xl text-sm leading-6 text-neutral-500">{description}</p>
      </div>
      {action}
    </div>
  );
}

export function StatCard({ label, value, helper }: { label: string; value: string; helper?: string }) {
  return (
    <Card>
      <CardContent>
        <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">{label}</p>
        <p className="mt-2 text-2xl font-semibold text-neutral-950">{value}</p>
        {helper ? <p className="mt-1 text-sm text-neutral-500">{helper}</p> : null}
      </CardContent>
    </Card>
  );
}

export function StatusBadge({ status }: { status: string }) {
  const tone =
    status === "active" || status === "trialing" || status === "published" || status === "paid"
      ? "success"
      : status === "suspended" || status === "cancelled" || status === "inactive"
        ? "danger"
        : status === "draft" || status === "unpublished"
          ? "warning"
          : "neutral";

  return <Badge tone={tone}>{status}</Badge>;
}

export function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="space-y-1 text-sm font-medium text-neutral-700">
      {label}
      {children}
    </label>
  );
}

export function formatMaybeDate(value?: string | null) {
  return value ? formatDateTime(value) : "—";
}

export function formatNumber(value: number) {
  return new Intl.NumberFormat("id-ID").format(value);
}
