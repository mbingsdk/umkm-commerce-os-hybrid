"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { deleteProduct } from "@/features/catalog/api/products.api";
import { useCategories } from "@/features/catalog/hooks/use-categories";
import { useProducts } from "@/features/catalog/hooks/use-products";
import type { ProductListItem, ProductStatus } from "@/features/catalog/types";
import { formatRupiah } from "@/lib/format/money";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";

export default function ProductsPage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const tenantId = useTenantStore((state) => state.selectedTenantId);
  const userPermissions = useTenantStore((state) => state.permissions);
  const canRead = userPermissions.includes(permissions.productRead);
  const canCreate = userPermissions.includes(permissions.productCreate);
  const canUpdate = userPermissions.includes(permissions.productUpdate);
  const canDelete = userPermissions.includes(permissions.productDelete);
  const [query, setQuery] = useState("");
  const [status, setStatus] = useState<ProductStatus | "">("");
  const [categoryId, setCategoryId] = useState("");
  const [deletingProduct, setDeletingProduct] = useState<ProductListItem | null>(null);
  const filters = { query, status, categoryId };
  const productsQuery = useProducts(filters, canRead);
  const categoriesQuery = useCategories(undefined, canRead);

  const deleteMutation = useMutation({
    mutationFn: (productId: string) => deleteProduct(productId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["products", tenantId] });
      setDeletingProduct(null);
    }
  });

  const categoryMap = useMemo(
    () => new Map((categoriesQuery.data ?? []).map((category) => [category.id, category.name])),
    [categoriesQuery.data]
  );

  if (!canRead) {
    return (
      <EmptyState
        title="Akses produk belum tersedia"
        description="Role aktifmu belum memiliki izin untuk melihat produk."
      />
    );
  }

  if (productsQuery.isPending || categoriesQuery.isPending) {
    return <LoadingState lines={4} />;
  }

  if (productsQuery.isError || categoriesQuery.isError) {
    return (
      <ErrorState
        title="Produk belum bisa dimuat"
        description="Coba muat ulang beberapa saat lagi."
        onRetry={() => {
          void productsQuery.refetch();
          void categoriesQuery.refetch();
        }}
      />
    );
  }

  const products = productsQuery.data ?? [];
  const columns: Array<DataTableColumn<ProductListItem>> = [
    {
      key: "name",
      header: "Produk",
      render: (product) => (
        <div>
          <p className="font-medium text-neutral-950">{product.name}</p>
          <p className="mt-1 text-xs text-neutral-500">{product.sku || "Tanpa SKU"}</p>
        </div>
      )
    },
    {
      key: "category",
      header: "Kategori",
      render: (product) =>
        product.categoryId ? categoryMap.get(product.categoryId) ?? "Kategori tidak ditemukan" : "Tanpa kategori"
    },
    {
      key: "price",
      header: "Harga",
      render: (product) => formatRupiah(product.price)
    },
    {
      key: "stock",
      header: "Stok",
      render: (product) => (
        <div>
          <p className="font-medium text-neutral-950">{product.stock.quantityAvailable} tersedia</p>
          <p className="mt-1 text-xs text-neutral-500">
            {product.stock.quantityReserved} reserved / {product.stock.quantityOnHand} fisik
          </p>
        </div>
      )
    },
    {
      key: "status",
      header: "Status",
      render: (product) => <ProductStatusBadge status={product.status} />
    },
    {
      key: "discoverable",
      header: "Discovery",
      render: (product) => (
        <Badge tone={product.isDiscoverable ? "info" : "neutral"}>
          {product.isDiscoverable ? "Tampil" : "Tersembunyi"}
        </Badge>
      )
    },
    {
      key: "actions",
      header: "Aksi",
      render: (product) => (
        <div className="flex flex-wrap gap-2">
          {canUpdate ? (
            <Button type="button" variant="outline" size="sm" onClick={() => router.push(`/dashboard/products/${product.id}/edit`)}>
              Edit
            </Button>
          ) : null}
          {canDelete ? (
            <Button type="button" variant="danger" size="sm" onClick={() => setDeletingProduct(product)}>
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
          <h1 className="text-2xl font-semibold text-neutral-950">Produk</h1>
          <p className="mt-1 text-sm text-neutral-500">Kelola katalog, harga, visibilitas, dan ringkasan stok.</p>
        </div>
        {canCreate ? <Button onClick={() => router.push("/dashboard/products/new")}>Tambah produk</Button> : null}
      </div>

      <div className="grid gap-3 rounded-2xl border border-neutral-200 bg-white p-4 md:grid-cols-[1fr_180px_220px]">
        <Input placeholder="Cari nama produk..." value={query} onChange={(event) => setQuery(event.target.value)} />
        <select
          className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
          value={status}
          onChange={(event) => setStatus(event.target.value as ProductStatus | "")}
        >
          <option value="">Semua status</option>
          <option value="draft">Draft</option>
          <option value="active">Aktif</option>
          <option value="inactive">Nonaktif</option>
          <option value="archived">Diarsipkan</option>
        </select>
        <select
          className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
          value={categoryId}
          onChange={(event) => setCategoryId(event.target.value)}
        >
          <option value="">Semua kategori</option>
          {(categoriesQuery.data ?? []).map((category) => (
            <option key={category.id} value={category.id}>
              {category.name}
            </option>
          ))}
        </select>
      </div>

      {products.length === 0 ? (
        <EmptyState
          title="Belum ada produk"
          description="Tambahkan produk pertama agar tokomu bisa mulai menerima pesanan."
          action={canCreate ? <Button onClick={() => router.push("/dashboard/products/new")}>Tambah produk</Button> : undefined}
        />
      ) : (
        <DataTable columns={columns} rows={products} getRowKey={(product) => product.id} />
      )}

      <Dialog
        open={!!deletingProduct}
        title="Hapus produk?"
        description="Produk akan dihapus secara aman dari katalog aktif."
        onClose={() => setDeletingProduct(null)}
        footer={
          <>
            <Button type="button" variant="outline" onClick={() => setDeletingProduct(null)}>
              Batal
            </Button>
            <Button
              type="button"
              variant="danger"
              isLoading={deleteMutation.isPending}
              onClick={() => {
                if (deletingProduct) {
                  deleteMutation.mutate(deletingProduct.id);
                }
              }}
            >
              Hapus produk
            </Button>
          </>
        }
      >
        <p className="text-sm text-neutral-600">
          Produk <span className="font-semibold text-neutral-950">{deletingProduct?.name}</span> tidak akan tampil lagi di
          katalog dashboard.
        </p>
        {deleteMutation.isError ? (
          <p className="mt-4 rounded-xl bg-red-50 p-3 text-sm text-red-700">{deleteMutation.error.message}</p>
        ) : null}
      </Dialog>
    </div>
  );
}

function ProductStatusBadge({ status }: { status: ProductStatus }) {
  const presentation = {
    active: { label: "Aktif", tone: "success" as const },
    draft: { label: "Draft", tone: "warning" as const },
    inactive: { label: "Nonaktif", tone: "neutral" as const },
    archived: { label: "Diarsipkan", tone: "danger" as const }
  }[status];

  return <Badge tone={presentation.tone}>{presentation.label}</Badge>;
}
