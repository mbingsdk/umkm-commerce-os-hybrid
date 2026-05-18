"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import {
  attachProductImage,
  createProduct
} from "@/features/catalog/api/products.api";
import { ProductForm } from "@/features/catalog/components/product-form";
import { useCategories } from "@/features/catalog/hooks/use-categories";
import type { ProductFormValues } from "@/features/catalog/schemas/product.schema";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";

export default function CreateProductPage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const tenantId = useTenantStore((state) => state.selectedTenantId);
  const userPermissions = useTenantStore((state) => state.permissions);
  const canCreate = userPermissions.includes(permissions.productCreate);
  const categoriesQuery = useCategories(undefined, canCreate);

  const createMutation = useMutation({
    mutationFn: async ({ values, files }: { values: ProductFormValues; files: File[] }) => {
      const product = await createProduct(toCreateInput(values));

      for (const [index, file] of files.entries()) {
        await attachProductImage(product.id, {
          file,
          altText: values.name,
          isPrimary: index === 0,
          sortOrder: index
        });
      }

      return product;
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["products", tenantId] });
      router.push("/dashboard/products");
    }
  });

  if (!canCreate) {
    return (
      <EmptyState
        title="Akses tambah produk belum tersedia"
        description="Role aktifmu belum memiliki izin untuk membuat produk."
      />
    );
  }

  if (categoriesQuery.isPending) {
    return <LoadingState lines={4} />;
  }

  if (categoriesQuery.isError) {
    return (
      <ErrorState
        title="Form produk belum bisa dimuat"
        description="Kategori belum berhasil dimuat. Coba lagi."
        onRetry={() => void categoriesQuery.refetch()}
      />
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-neutral-950">Tambah produk</h1>
        <p className="mt-1 text-sm text-neutral-500">Mulai dari detail inti dulu; stok awal dan gambar akan ikut diproses.</p>
      </div>

      <ProductForm
        mode="create"
        categories={categoriesQuery.data ?? []}
        isSubmitting={createMutation.isPending}
        error={createMutation.isError ? createMutation.error.message : undefined}
        onSubmit={(values, files) => createMutation.mutate({ values, files })}
      />
    </div>
  );
}

function toCreateInput(values: ProductFormValues) {
  return {
    categoryId: values.categoryId || undefined,
    name: values.name,
    slug: values.slug,
    description: values.description,
    sku: values.sku,
    barcode: values.barcode,
    price: values.price,
    compareAtPrice: values.compareAtPrice,
    costPrice: values.costPrice,
    weightGram: values.weightGram,
    status: values.status,
    isDiscoverable: values.isDiscoverable,
    trackInventory: values.trackInventory,
    allowBackorder: values.allowBackorder,
    initialStock: values.initialStock
  };
}
