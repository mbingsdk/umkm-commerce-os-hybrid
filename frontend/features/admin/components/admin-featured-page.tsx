"use client";

import { useMemo, useState, type FormEvent } from "react";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  AdminPageHeader,
  Field,
  StatusBadge,
  formatMaybeDate
} from "@/features/admin/components/admin-shared";
import { useAdminFeaturedItems, useAdminFeaturedMutations } from "@/features/admin/hooks/use-admin";
import type { AdminFeaturedItem, FeaturedFormInput } from "@/features/admin/types";
import { useToastStore } from "@/lib/stores/toast.store";

export function AdminFeaturedPage() {
  const [itemType, setItemType] = useState("");
  const [placement, setPlacement] = useState("");
  const [isActive, setIsActive] = useState("");
  const [cursor, setCursor] = useState<string | undefined>();
  const [editingItem, setEditingItem] = useState<AdminFeaturedItem | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const filters = useMemo(
    () => ({ itemType, placement, isActive, cursor, limit: 20 }),
    [cursor, isActive, itemType, placement]
  );
  const featuredQuery = useAdminFeaturedItems(filters);
  const { createFeatured, updateFeatured, deleteFeatured } = useAdminFeaturedMutations();
  const pushToast = useToastStore((state) => state.pushToast);

  const columns: Array<DataTableColumn<AdminFeaturedItem>> = [
    {
      key: "target",
      header: "Target",
      render: (item) => (
        <div>
          <p className="font-semibold text-neutral-950">{item.targetName || item.productId || item.storeId || "Target"}</p>
          <p className="mt-1 text-xs text-neutral-500">{item.targetSlug || item.id}</p>
        </div>
      )
    },
    { key: "type", header: "Tipe", render: (item) => item.itemType },
    { key: "placement", header: "Placement", render: (item) => item.placement },
    { key: "order", header: "Urutan", render: (item) => item.sortOrder },
    { key: "active", header: "Status", render: (item) => <StatusBadge status={item.isActive ? "active" : "inactive"} /> },
    {
      key: "schedule",
      header: "Jadwal",
      render: (item) => (
        <p className="text-xs leading-5 text-neutral-500">
          Mulai {formatMaybeDate(item.startsAt)}
          <br />
          Selesai {formatMaybeDate(item.endsAt)}
        </p>
      )
    },
    {
      key: "action",
      header: "",
      render: (item) => (
        <div className="flex gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => {
              setEditingItem(item);
              setDialogOpen(true);
            }}
          >
            Edit
          </Button>
          <Button
            type="button"
            variant="danger"
            size="sm"
            isLoading={deleteFeatured.isPending}
            disabled={deleteFeatured.isPending}
            onClick={() => {
              if (deleteFeatured.isPending) {
                return;
              }

              if (!window.confirm("Hapus featured item ini?")) {
                return;
              }
              deleteFeatured.mutate(item.id, {
                onSuccess: () => pushToast({ tone: "success", title: "Featured item dihapus" }),
                onError: (error) => pushToast({ tone: "error", title: "Featured gagal dihapus", description: error.message })
              });
            }}
          >
            Hapus
          </Button>
        </div>
      )
    }
  ];

  function submitFeatured(values: FeaturedFormInput) {
    if (createFeatured.isPending || updateFeatured.isPending) {
      return;
    }

    if (editingItem) {
      updateFeatured.mutate(
        { featuredId: editingItem.id, input: values },
        {
          onSuccess: () => {
            setDialogOpen(false);
            setEditingItem(null);
            pushToast({ tone: "success", title: "Featured item diperbarui" });
          },
          onError: (error) => pushToast({ tone: "error", title: "Featured gagal disimpan", description: error.message })
        }
      );
      return;
    }
    createFeatured.mutate(values, {
      onSuccess: () => {
        setDialogOpen(false);
        pushToast({ tone: "success", title: "Featured item dibuat" });
      },
      onError: (error) => pushToast({ tone: "error", title: "Featured gagal dibuat", description: error.message })
    });
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Featured discovery"
        description="Kelola item unggulan untuk discovery publik. Backend tetap memvalidasi tenant/store/product eligibility."
        action={
          <Button
            type="button"
            onClick={() => {
              setEditingItem(null);
              setDialogOpen(true);
            }}
          >
            Tambah featured
          </Button>
        }
      />

      <div className="grid gap-3 rounded-2xl border border-neutral-200 bg-white p-4 md:grid-cols-[180px_180px_180px_auto]">
        <select
          className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm"
          value={itemType}
          onChange={(event) => {
            setCursor(undefined);
            setItemType(event.target.value);
          }}
        >
          <option value="">Semua tipe</option>
          <option value="store">Store</option>
          <option value="product">Product</option>
        </select>
        <select
          className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm"
          value={placement}
          onChange={(event) => {
            setCursor(undefined);
            setPlacement(event.target.value);
          }}
        >
          <option value="">Semua placement</option>
          {placements.map((item) => (
            <option key={item} value={item}>
              {item}
            </option>
          ))}
        </select>
        <select
          className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm"
          value={isActive}
          onChange={(event) => {
            setCursor(undefined);
            setIsActive(event.target.value);
          }}
        >
          <option value="">Semua status</option>
          <option value="true">Aktif</option>
          <option value="false">Nonaktif</option>
        </select>
        <Button
          type="button"
          variant="outline"
          onClick={() => {
            setItemType("");
            setPlacement("");
            setIsActive("");
            setCursor(undefined);
          }}
        >
          Reset
        </Button>
      </div>

      {featuredQuery.isPending ? (
        <LoadingState lines={4} />
      ) : featuredQuery.isError ? (
        <ErrorState
          title="Featured gagal dimuat"
          description="Coba muat ulang daftar featured discovery."
          onRetry={() => void featuredQuery.refetch()}
        />
      ) : featuredQuery.data.items.length === 0 ? (
        <EmptyState title="Belum ada featured item" description="Tambahkan store atau produk yang eligible untuk tampil unggulan." />
      ) : (
        <>
          <DataTable columns={columns} rows={featuredQuery.data.items} getRowKey={(item) => item.id} />
          <div className="flex justify-end">
            <Button
              type="button"
              variant="outline"
              disabled={!featuredQuery.data.pagination.hasMore}
              onClick={() => setCursor(featuredQuery.data.pagination.nextCursor ?? undefined)}
            >
              Muat berikutnya
            </Button>
          </div>
        </>
      )}

      <FeaturedDialog
        open={dialogOpen}
        item={editingItem}
        isSubmitting={createFeatured.isPending || updateFeatured.isPending}
        onClose={() => {
          setDialogOpen(false);
          setEditingItem(null);
        }}
        onSubmit={submitFeatured}
      />
    </div>
  );
}

const placements: Array<FeaturedFormInput["placement"]> = ["home", "stores", "products", "category", "city"];

function FeaturedDialog({
  open,
  item,
  isSubmitting,
  onClose,
  onSubmit
}: {
  open: boolean;
  item: AdminFeaturedItem | null;
  isSubmitting: boolean;
  onClose: () => void;
  onSubmit: (values: FeaturedFormInput) => void;
}) {
  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (isSubmitting) {
      return;
    }

    const form = new FormData(event.currentTarget);
    onSubmit({
      itemType: String(form.get("item_type") ?? "store") as FeaturedFormInput["itemType"],
      tenantId: String(form.get("tenant_id") ?? "").trim(),
      storeId: String(form.get("store_id") ?? "").trim(),
      productId: String(form.get("product_id") ?? "").trim(),
      placement: String(form.get("placement") ?? "home") as FeaturedFormInput["placement"],
      sortOrder: Number(form.get("sort_order") ?? 0),
      startsAt: String(form.get("starts_at") ?? ""),
      endsAt: String(form.get("ends_at") ?? ""),
      isActive: form.get("is_active") === "on"
    });
  }

  return (
    <Dialog
      open={open}
      title={item ? "Edit featured item" : "Tambah featured item"}
      description="Masukkan UUID target. Untuk product, product_id wajib dan store_id boleh membantu validasi."
      onClose={onClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={onClose}>
            Batal
          </Button>
          <Button type="submit" form="admin-featured-form" isLoading={isSubmitting} disabled={isSubmitting}>
            Simpan
          </Button>
        </>
      }
    >
      <form key={item?.id ?? "new"} id="admin-featured-form" className="space-y-4" onSubmit={handleSubmit}>
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Tipe">
            <select name="item_type" defaultValue={item?.itemType ?? "store"} className="h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm">
              <option value="store">Store</option>
              <option value="product">Product</option>
            </select>
          </Field>
          <Field label="Placement">
            <select name="placement" defaultValue={item?.placement ?? "home"} className="h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm">
              {placements.map((placement) => (
                <option key={placement} value={placement}>
                  {placement}
                </option>
              ))}
            </select>
          </Field>
        </div>
        <Field label="Tenant ID">
          <Input name="tenant_id" defaultValue={item?.tenantId ?? ""} required />
        </Field>
        <Field label="Store ID">
          <Input name="store_id" defaultValue={item?.storeId ?? ""} placeholder="Wajib untuk store, opsional untuk product" />
        </Field>
        <Field label="Product ID">
          <Input name="product_id" defaultValue={item?.productId ?? ""} placeholder="Wajib untuk product" />
        </Field>
        <div className="grid gap-4 sm:grid-cols-3">
          <Field label="Urutan">
            <Input name="sort_order" type="number" defaultValue={item?.sortOrder ?? 0} />
          </Field>
          <Field label="Mulai">
            <Input name="starts_at" type="datetime-local" defaultValue={toDatetimeLocal(item?.startsAt)} />
          </Field>
          <Field label="Selesai">
            <Input name="ends_at" type="datetime-local" defaultValue={toDatetimeLocal(item?.endsAt)} />
          </Field>
        </div>
        <label className="flex items-center gap-2 rounded-xl border border-neutral-200 p-3 text-sm text-neutral-700">
          <input name="is_active" type="checkbox" defaultChecked={item?.isActive ?? true} />
          Featured aktif
        </label>
      </form>
    </Dialog>
  );
}

function toDatetimeLocal(value?: string | null) {
  if (!value) {
    return "";
  }
  return value.slice(0, 16);
}
