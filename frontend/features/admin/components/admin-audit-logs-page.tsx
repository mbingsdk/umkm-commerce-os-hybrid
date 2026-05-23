"use client";

import { useMemo, useState } from "react";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { AdminPageHeader } from "@/features/admin/components/admin-shared";
import { useAdminAuditLogs } from "@/features/admin/hooks/use-admin";
import type { AdminAuditLog } from "@/features/admin/types";
import { formatDateTime } from "@/lib/format/date";

export function AdminAuditLogsPage() {
  const [action, setAction] = useState("");
  const [targetType, setTargetType] = useState("");
  const [targetId, setTargetId] = useState("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [cursor, setCursor] = useState<string | undefined>();
  const [selected, setSelected] = useState<AdminAuditLog | null>(null);
  const filters = useMemo(
    () => ({ action, targetType, targetId, dateFrom, dateTo, cursor, limit: 20 }),
    [action, cursor, dateFrom, dateTo, targetId, targetType]
  );
  const auditQuery = useAdminAuditLogs(filters);

  const columns: Array<DataTableColumn<AdminAuditLog>> = [
    {
      key: "action",
      header: "Action",
      render: (log) => (
        <div>
          <p className="font-semibold text-neutral-950">{log.action}</p>
          <p className="mt-1 text-xs text-neutral-500">{log.targetType || "—"}</p>
        </div>
      )
    },
    { key: "actor", header: "Actor", render: (log) => log.actorName || log.actorUserId || "System" },
    { key: "target", header: "Target", render: (log) => log.targetId ?? "—" },
    { key: "created", header: "Waktu", render: (log) => formatDateTime(log.createdAt) },
    {
      key: "detail",
      header: "",
      render: (log) => (
        <Button type="button" variant="outline" size="sm" onClick={() => setSelected(log)}>
          Detail
        </Button>
      )
    }
  ];

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Audit logs"
        description="Browse admin audit dengan payload yang sudah dire-dact oleh backend/UI agar field sensitif tidak terbuka."
      />

      <Card>
        <CardContent className="grid gap-3 md:grid-cols-5">
          <Input placeholder="Action" value={action} onChange={(event) => setActionAndReset(event.target.value, setAction, setCursor)} />
          <Input placeholder="Target type" value={targetType} onChange={(event) => setActionAndReset(event.target.value, setTargetType, setCursor)} />
          <Input placeholder="Target ID" value={targetId} onChange={(event) => setActionAndReset(event.target.value, setTargetId, setCursor)} />
          <Input type="date" value={dateFrom} onChange={(event) => setActionAndReset(event.target.value, setDateFrom, setCursor)} />
          <Input type="date" value={dateTo} onChange={(event) => setActionAndReset(event.target.value, setDateTo, setCursor)} />
        </CardContent>
      </Card>

      {auditQuery.isPending ? (
        <LoadingState lines={4} />
      ) : auditQuery.isError ? (
        <ErrorState
          title="Audit log gagal dimuat"
          description="Coba muat ulang audit log admin."
          onRetry={() => void auditQuery.refetch()}
        />
      ) : auditQuery.data.items.length === 0 ? (
        <EmptyState title="Audit log kosong" description="Tidak ada audit log sesuai filter saat ini." />
      ) : (
        <>
          <DataTable columns={columns} rows={auditQuery.data.items} getRowKey={(log) => log.id} />
          <div className="flex justify-end">
            <Button
              type="button"
              variant="outline"
              disabled={!auditQuery.data.pagination.hasMore}
              onClick={() => setCursor(auditQuery.data.pagination.nextCursor ?? undefined)}
            >
              Muat berikutnya
            </Button>
          </div>
        </>
      )}

      <Dialog
        open={!!selected}
        title="Detail audit"
        description="Payload sensitif harus tetap dalam bentuk redacted."
        onClose={() => setSelected(null)}
      >
        {selected ? (
          <div className="space-y-4">
            <AuditPayload title="Before" value={selected.beforeData} />
            <AuditPayload title="After" value={selected.afterData} />
            <div className="rounded-xl bg-neutral-50 p-3 text-xs leading-5 text-neutral-500">
              <p>IP: {selected.ipAddress || "—"}</p>
              <p>User agent: {selected.userAgent || "—"}</p>
            </div>
          </div>
        ) : null}
      </Dialog>
    </div>
  );
}

function setActionAndReset(value: string, setter: (value: string) => void, setCursor: (value: string | undefined) => void) {
  setCursor(undefined);
  setter(value);
}

function AuditPayload({ title, value }: { title: string; value: unknown }) {
  return (
    <div>
      <p className="mb-2 text-sm font-semibold text-neutral-950">{title}</p>
      <pre className="max-h-72 overflow-auto rounded-xl bg-neutral-950 p-3 text-xs leading-5 text-neutral-100">
        {JSON.stringify(value ?? {}, null, 2)}
      </pre>
    </div>
  );
}
