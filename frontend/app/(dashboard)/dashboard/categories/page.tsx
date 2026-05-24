"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Dialog } from "@/components/ui/dialog";
import {
  createCategory,
  deleteCategory,
  updateCategory
} from "@/features/catalog/api/categories.api";
import { CategoryFormDialog } from "@/features/catalog/components/category-form-dialog";
import { useCategories } from "@/features/catalog/hooks/use-categories";
import type { Category } from "@/features/catalog/types";
import type { CategoryFormValues } from "@/features/catalog/schemas/category.schema";
import { queryKeys } from "@/lib/api/query-keys";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";

export default function CategoriesPage() {
  const queryClient = useQueryClient();
  const tenantId = useTenantStore((state) => state.selectedTenantId);
  const userPermissions = useTenantStore((state) => state.permissions);
  const canRead = userPermissions.includes(permissions.categoryRead);
  const canCreate = userPermissions.includes(permissions.categoryCreate);
  const canUpdate = userPermissions.includes(permissions.categoryUpdate);
  const canDelete = userPermissions.includes(permissions.categoryDelete);
  const categoriesQuery = useCategories(undefined, canRead);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingCategory, setEditingCategory] = useState<Category | null>(null);
  const [deletingCategory, setDeletingCategory] = useState<Category | null>(null);

  const saveMutation = useMutation({
    mutationFn: async (values: CategoryFormValues) => {
      const payload = {
        name: values.name,
        slug: values.slug,
        description: values.description,
        sort_order: values.sortOrder,
        is_active: values.isActive
      };

      if (editingCategory) {
        return updateCategory(editingCategory.id, payload);
      }

      return createCategory(payload);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.categories(tenantId) });
      setDialogOpen(false);
      setEditingCategory(null);
    }
  });

  const deleteMutation = useMutation({
    mutationFn: (categoryId: string) => deleteCategory(categoryId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.categories(tenantId) });
      setDeletingCategory(null);
    }
  });

  if (!canRead) {
    return (
      <EmptyState
        title="Akses kategori belum tersedia"
        description="Role aktifmu belum memiliki izin untuk melihat kategori. Backend tetap menjadi sumber kebenaran izin."
      />
    );
  }

  if (categoriesQuery.isPending) {
    return <LoadingState lines={4} />;
  }

  if (categoriesQuery.isError) {
    return (
      <ErrorState
        title="Kategori belum bisa dimuat"
        description="Coba muat ulang beberapa saat lagi."
        onRetry={() => void categoriesQuery.refetch()}
      />
    );
  }

  const categories = categoriesQuery.data ?? [];
  const columns: Array<DataTableColumn<Category>> = [
    {
      key: "name",
      header: "Nama",
      render: (category) => (
        <div>
          <p className="font-medium text-neutral-950">{category.name}</p>
          <p className="mt-1 text-xs text-neutral-500">/{category.slug}</p>
        </div>
      )
    },
    {
      key: "description",
      header: "Deskripsi",
      render: (category) => category.description || <span className="text-neutral-400">Belum diisi</span>
    },
    {
      key: "status",
      header: "Status",
      render: (category) => (
        <Badge tone={category.isActive ? "success" : "neutral"}>{category.isActive ? "Aktif" : "Nonaktif"}</Badge>
      )
    },
    {
      key: "sort",
      header: "Urutan",
      render: (category) => category.sortOrder
    },
    {
      key: "actions",
      header: "Aksi",
      render: (category) => (
        <div className="flex flex-wrap gap-2">
          {canUpdate ? (
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => {
                setEditingCategory(category);
                setDialogOpen(true);
              }}
            >
              Edit
            </Button>
          ) : null}
          {canDelete ? (
            <Button type="button" variant="danger" size="sm" onClick={() => setDeletingCategory(category)}>
              Hapus
            </Button>
          ) : null}
        </div>
      )
    }
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-neutral-950">Kategori</h1>
          <p className="mt-1 text-sm text-neutral-500">Kelola pengelompokan produk tokomu.</p>
        </div>
        {canCreate ? (
          <Button
            type="button"
            onClick={() => {
              setEditingCategory(null);
              setDialogOpen(true);
            }}
          >
            Tambah kategori
          </Button>
        ) : null}
      </div>

      {categories.length === 0 ? (
        <EmptyState
          title="Belum ada kategori"
          description="Buat kategori pertama agar produk lebih mudah dicari di dashboard."
          action={canCreate ? <Button onClick={() => setDialogOpen(true)}>Tambah kategori</Button> : undefined}
        />
      ) : (
        <DataTable columns={columns} rows={categories} getRowKey={(category) => category.id} />
      )}

      <CategoryFormDialog
        open={dialogOpen}
        category={editingCategory}
        isSubmitting={saveMutation.isPending}
        error={saveMutation.isError ? saveMutation.error.message : undefined}
        onClose={() => {
          setDialogOpen(false);
          setEditingCategory(null);
        }}
        onSubmit={(values) => {
          if (!saveMutation.isPending) {
            saveMutation.mutate(values);
          }
        }}
      />

      <Dialog
        open={!!deletingCategory}
        title="Hapus kategori?"
        description="Aksi ini memakai soft delete di backend dan perlu dikonfirmasi."
        onClose={() => setDeletingCategory(null)}
        footer={
          <>
            <Button type="button" variant="outline" onClick={() => setDeletingCategory(null)}>
              Batal
            </Button>
            <Button
              type="button"
              variant="danger"
              isLoading={deleteMutation.isPending}
              disabled={deleteMutation.isPending}
              onClick={() => {
                if (!deleteMutation.isPending && deletingCategory) {
                  deleteMutation.mutate(deletingCategory.id);
                }
              }}
            >
              Hapus kategori
            </Button>
          </>
        }
      >
        <p className="text-sm text-neutral-600">
          Kategori <span className="font-semibold text-neutral-950">{deletingCategory?.name}</span> tidak akan tampil lagi
          setelah dihapus.
        </p>
        {deleteMutation.isError ? (
          <p className="mt-4 rounded-xl bg-red-50 p-3 text-sm text-red-700">{deleteMutation.error.message}</p>
        ) : null}
      </Dialog>
    </div>
  );
}
