"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import {
  attachProductImage,
  deleteProductImage,
  updateProduct
} from "@/features/catalog/api/products.api";
import { ProductForm } from "@/features/catalog/components/product-form";
import { useCategories } from "@/features/catalog/hooks/use-categories";
import { useProduct } from "@/features/catalog/hooks/use-products";
import type { ProductFormValues } from "@/features/catalog/schemas/product.schema";
import { queryKeys } from "@/lib/api/query-keys";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";

export default function EditProductPage() {
  const params = useParams<{ productId: string }>();
  const productId = params.productId;
  const queryClient = useQueryClient();
  const tenantId = useTenantStore((state) => state.selectedTenantId);
  const userPermissions = useTenantStore((state) => state.permissions);
  const canUpdate = userPermissions.includes(permissions.productUpdate);
  const canUploadImage = userPermissions.includes(permissions.productUploadImage);
  const productQuery = useProduct(productId, canUpdate);
  const categoriesQuery = useCategories(undefined, canUpdate);

  const updateMutation = useMutation({
    mutationFn: async ({ values, files }: { values: ProductFormValues; files: File[] }) => {
      await updateProduct(productId, toUpdateInput(values));

      if (canUploadImage) {
        for (const [index, file] of files.entries()) {
          await attachProductImage(productId, {
            file,
            altText: values.name,
            isPrimary: productQuery.data?.images.length === 0 && index === 0,
            sortOrder: (productQuery.data?.images.length ?? 0) + index
          });
        }
      }
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.product(tenantId, productId) }),
        queryClient.invalidateQueries({ queryKey: ["products", tenantId] })
      ]);
    }
  });

  const deleteImageMutation = useMutation({
    mutationFn: (imageId: string) => deleteProductImage(productId, imageId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.product(tenantId, productId) });
    }
  });

  if (!canUpdate) {
    return (
      <EmptyState
        title="Akses edit produk belum tersedia"
        description="Role aktifmu belum memiliki izin untuk mengubah produk."
      />
    );
  }

  if (productQuery.isPending || categoriesQuery.isPending) {
    return <LoadingState lines={4} />;
  }

  if (productQuery.isError || categoriesQuery.isError) {
    return (
      <ErrorState
        title="Produk belum bisa dimuat"
        description="Coba muat ulang sebelum melanjutkan edit."
        onRetry={() => {
          void productQuery.refetch();
          void categoriesQuery.refetch();
        }}
      />
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-neutral-950">Edit produk</h1>
        <p className="mt-1 text-sm text-neutral-500">Perbarui detail katalog tanpa mengubah riwayat order lama.</p>
      </div>

      <ProductForm
        key={`${productQuery.data.id}-${productQuery.data.images.length}`}
        mode="edit"
        categories={categoriesQuery.data ?? []}
        initialProduct={productQuery.data}
        isSubmitting={updateMutation.isPending || deleteImageMutation.isPending}
        error={
          updateMutation.isError
            ? updateMutation.error.message
            : deleteImageMutation.isError
              ? deleteImageMutation.error.message
              : undefined
        }
        onSubmit={(values, files) => updateMutation.mutate({ values, files })}
        onDeleteImage={canUploadImage ? (imageId) => deleteImageMutation.mutate(imageId) : undefined}
      />
    </div>
  );
}

function toUpdateInput(values: ProductFormValues) {
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
    allowBackorder: values.allowBackorder
  };
}
