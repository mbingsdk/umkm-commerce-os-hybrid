"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import {
  createCourierZone,
  deleteCourierZone,
  updateCourierZone
} from "@/features/courier/api/courier.api";
import { CourierZoneDialog } from "@/features/courier/components/courier-zone-dialog";
import { useCourierZones } from "@/features/courier/hooks/use-courier-zones";
import type { CourierZone, CourierZoneInput } from "@/features/courier/types";
import { queryKeys } from "@/lib/api/query-keys";
import { formatDateTime } from "@/lib/format/date";
import { formatRupiah } from "@/lib/format/money";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";
import { useToastStore } from "@/lib/stores/toast.store";

export function CourierZonesPage() {
  const queryClient = useQueryClient();
  const pushToast = useToastStore((state) => state.pushToast);
  const tenantId = useTenantStore((state) => state.selectedTenantId);
  const userPermissions = useTenantStore((state) => state.permissions);
  const canRead = userPermissions.includes(permissions.courierReadZone);
  const canCreate = userPermissions.includes(permissions.courierCreateZone);
  const canUpdate = userPermissions.includes(permissions.courierUpdateZone);
  const canDelete = userPermissions.includes(permissions.courierDeleteZone);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingZone, setEditingZone] = useState<CourierZone | null>(null);
  const [deletingZone, setDeletingZone] = useState<CourierZone | null>(null);
  const zonesQuery = useCourierZones(canRead);

  const invalidate = async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeys.courierZones(tenantId) });
  };

  const createMutation = useMutation({
    mutationFn: createCourierZone,
    onSuccess: async () => {
      await invalidate();
      setDialogOpen(false);
      pushToast({ tone: "success", title: "Zona kurir ditambahkan" });
    },
    onError: (error) => pushToast({ tone: "error", title: "Gagal menambah zona", description: error.message })
  });

  const updateMutation = useMutation({
    mutationFn: ({ zoneId, values }: { zoneId: string; values: Partial<CourierZoneInput> }) =>
      updateCourierZone(zoneId, values),
    onSuccess: async () => {
      await invalidate();
      setDialogOpen(false);
      setEditingZone(null);
      pushToast({ tone: "success", title: "Zona kurir diperbarui" });
    },
    onError: (error) => pushToast({ tone: "error", title: "Gagal memperbarui zona", description: error.message })
  });

  const deleteMutation = useMutation({
    mutationFn: deleteCourierZone,
    onSuccess: async () => {
      await invalidate();
      setDeletingZone(null);
      pushToast({ tone: "success", title: "Zona kurir dihapus" });
    },
    onError: (error) => pushToast({ tone: "error", title: "Gagal menghapus zona", description: error.message })
  });

  if (!canRead) {
    return (
      <EmptyState
        title="Akses zona kurir belum tersedia"
        description="Role aktifmu belum memiliki izin untuk melihat pengaturan zona kurir."
      />
    );
  }

  if (zonesQuery.isPending) {
    return <LoadingState lines={4} />;
  }

  if (zonesQuery.isError) {
    return (
      <ErrorState
        title="Zona kurir belum bisa dimuat"
        description="Coba muat ulang sebelum mengubah pengaturan ongkir."
        onRetry={() => void zonesQuery.refetch()}
      />
    );
  }

  const zones = zonesQuery.data ?? [];
  const columns: Array<DataTableColumn<CourierZone>> = [
    {
      key: "name",
      header: "Zona",
      render: (zone) => (
        <div>
          <p className="font-semibold text-neutral-950">{zone.name}</p>
          <p className="mt-1 text-xs text-neutral-500">{zone.description || "Tanpa deskripsi"}</p>
        </div>
      )
    },
    {
      key: "rate",
      header: "Ongkir",
      render: (zone) => <span className="font-semibold text-neutral-950">{formatRupiah(zone.rate)}</span>
    },
    {
      key: "status",
      header: "Status",
      render: (zone) => <Badge tone={zone.isActive ? "success" : "neutral"}>{zone.isActive ? "Aktif" : "Nonaktif"}</Badge>
    },
    {
      key: "order",
      header: "Urutan",
      render: (zone) => zone.sortOrder
    },
    {
      key: "updated",
      header: "Update",
      render: (zone) => <span className="text-xs text-neutral-500">{formatDateTime(zone.updatedAt)}</span>
    },
    {
      key: "actions",
      header: "Aksi",
      render: (zone) => (
        <div className="flex flex-wrap gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={!canUpdate || updateMutation.isPending}
            onClick={() =>
              updateMutation.mutate({
                zoneId: zone.id,
                values: { isActive: !zone.isActive }
              })
            }
          >
            {zone.isActive ? "Nonaktifkan" : "Aktifkan"}
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={!canUpdate}
            onClick={() => {
              setEditingZone(zone);
              setDialogOpen(true);
            }}
          >
            Edit
          </Button>
          <Button type="button" variant="danger" size="sm" disabled={!canDelete} onClick={() => setDeletingZone(zone)}>
            Hapus
          </Button>
        </div>
      )
    }
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-neutral-950">Zona kurir</h1>
          <p className="mt-1 text-sm text-neutral-500">
            Atur opsi ongkir lokal yang akan tampil ke customer saat checkout dan tracking.
          </p>
        </div>
        <Button
          type="button"
          disabled={!canCreate}
          onClick={() => {
            setEditingZone(null);
            setDialogOpen(true);
          }}
        >
          Tambah zona
        </Button>
      </div>

      {zones.length === 0 ? (
        <EmptyState
          title="Belum ada zona kurir"
          description="Tambahkan zona seperti dalam kota, luar kota, atau pickup manual untuk checkout publik."
          action={
            canCreate ? (
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setEditingZone(null);
                  setDialogOpen(true);
                }}
              >
                Tambah zona pertama
              </Button>
            ) : null
          }
        />
      ) : (
        <DataTable columns={columns} rows={zones} getRowKey={(zone) => zone.id} />
      )}

      <CourierZoneDialog
        open={dialogOpen}
        zone={editingZone}
        isSubmitting={createMutation.isPending || updateMutation.isPending}
        error={createMutation.error?.message ?? updateMutation.error?.message}
        onClose={() => {
          setDialogOpen(false);
          setEditingZone(null);
        }}
        onSubmit={(values) => {
          if (editingZone) {
            updateMutation.mutate({ zoneId: editingZone.id, values });
            return;
          }
          createMutation.mutate(values);
        }}
      />

      <Dialog
        open={!!deletingZone}
        title="Hapus zona kurir?"
        description="Zona yang dihapus tidak akan tampil sebagai pilihan pengiriman publik."
        onClose={() => setDeletingZone(null)}
        footer={
          <>
            <Button type="button" variant="outline" onClick={() => setDeletingZone(null)}>
              Batal
            </Button>
            <Button
              type="button"
              variant="danger"
              isLoading={deleteMutation.isPending}
              onClick={() => deletingZone && deleteMutation.mutate(deletingZone.id)}
            >
              Hapus
            </Button>
          </>
        }
      >
        <p className="text-sm leading-6 text-neutral-600">
          Kamu akan menghapus <span className="font-semibold text-neutral-950">{deletingZone?.name}</span>.
        </p>
      </Dialog>
    </div>
  );
}
