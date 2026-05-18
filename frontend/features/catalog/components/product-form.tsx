"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm, useWatch } from "react-hook-form";
import { useEffect, useState } from "react";
import { ImageUploader } from "@/components/forms/image-uploader";
import { MoneyInput } from "@/components/forms/money-input";
import { SlugInput } from "@/components/forms/slug-input";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import type { Category, ProductDetail } from "@/features/catalog/types";
import {
  productSchema,
  type ProductFormValues
} from "@/features/catalog/schemas/product.schema";
import { formatRupiah } from "@/lib/format/money";
import { useTenantStore } from "@/lib/stores/tenant.store";

type ProductFormProps = {
  mode: "create" | "edit";
  categories: Category[];
  initialProduct?: ProductDetail;
  isSubmitting?: boolean;
  error?: string;
  onSubmit: (values: ProductFormValues, imageFiles: File[]) => void;
  onDeleteImage?: (imageId: string) => void;
};

export function ProductForm({
  mode,
  categories,
  initialProduct,
  isSubmitting = false,
  error,
  onSubmit,
  onDeleteImage
}: ProductFormProps) {
  const role = useTenantStore((state) => state.role);
  const canSeeCostPrice = role === "owner" || role === "manager";
  const [imageFiles, setImageFiles] = useState<File[]>([]);
  const form = useForm<ProductFormValues>({
    resolver: zodResolver(productSchema),
    defaultValues: toDefaults(initialProduct)
  });
  const name = useWatch({ control: form.control, name: "name" });

  useEffect(() => {
    form.reset(toDefaults(initialProduct));
  }, [form, initialProduct]);

  return (
    <form className="space-y-6" onSubmit={form.handleSubmit((values) => onSubmit(values, imageFiles))}>
      <Card>
        <CardHeader>
          <CardTitle>Informasi dasar</CardTitle>
          <CardDescription>Nama, slug, kategori, dan identitas produk yang terlihat di dashboard.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 lg:grid-cols-2">
          <Field label="Nama produk" error={form.formState.errors.name?.message}>
            <Input placeholder="Bouquet Mawar Merah" hasError={!!form.formState.errors.name} {...form.register("name")} />
          </Field>

          <Field label="Slug" error={form.formState.errors.slug?.message}>
            <Controller
              control={form.control}
              name="slug"
              render={({ field }) => (
                <SlugInput
                  sourceValue={name}
                  value={field.value}
                  onChange={field.onChange}
                  hasError={!!form.formState.errors.slug}
                  placeholder="bouquet-mawar-merah"
                />
              )}
            />
          </Field>

          <Field label="Kategori" error={form.formState.errors.categoryId?.message}>
            <select
              className="h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
              {...form.register("categoryId")}
            >
              <option value="">Tanpa kategori</option>
              {categories.map((category) => (
                <option key={category.id} value={category.id}>
                  {category.name}
                </option>
              ))}
            </select>
          </Field>

          <Field label="SKU" error={form.formState.errors.sku?.message}>
            <Input placeholder="BQT-MWR-001" hasError={!!form.formState.errors.sku} {...form.register("sku")} />
          </Field>

          <Field label="Barcode" error={form.formState.errors.barcode?.message}>
            <Input placeholder="899000000001" hasError={!!form.formState.errors.barcode} {...form.register("barcode")} />
          </Field>

          <Field label="Berat (gram)" error={form.formState.errors.weightGram?.message}>
            <Input
              type="number"
              min={0}
              hasError={!!form.formState.errors.weightGram}
              {...form.register("weightGram", { valueAsNumber: true })}
            />
          </Field>

          <label className="block space-y-2 text-sm font-medium text-neutral-700 lg:col-span-2">
            Deskripsi
            <textarea
              className="min-h-28 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-sm text-neutral-950 outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
              placeholder="Ceritakan produk secara singkat."
              {...form.register("description")}
            />
          </label>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Harga</CardTitle>
          <CardDescription>Gunakan Rupiah. Backend tetap memvalidasi aturan harga saat disimpan.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 lg:grid-cols-3">
          <Field label="Harga jual" error={form.formState.errors.price?.message}>
            <Controller
              control={form.control}
              name="price"
              render={({ field }) => (
                <MoneyInput
                  value={field.value}
                  onChange={(value) => field.onChange(value ?? 0)}
                  hasError={!!form.formState.errors.price}
                />
              )}
            />
          </Field>

          <Field label="Harga pembanding" error={form.formState.errors.compareAtPrice?.message}>
            <Controller
              control={form.control}
              name="compareAtPrice"
              render={({ field }) => (
                <MoneyInput
                  value={field.value}
                  onChange={field.onChange}
                  hasError={!!form.formState.errors.compareAtPrice}
                />
              )}
            />
          </Field>

          {canSeeCostPrice ? (
            <Field label="Harga modal" error={form.formState.errors.costPrice?.message}>
              <Controller
                control={form.control}
                name="costPrice"
                render={({ field }) => (
                  <MoneyInput
                    value={field.value}
                    onChange={field.onChange}
                    hasError={!!form.formState.errors.costPrice}
                  />
                )}
              />
            </Field>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Inventori</CardTitle>
          <CardDescription>
            {mode === "create"
              ? "Stok awal akan membuat snapshot dan movement awal di backend."
              : "Perubahan stok lanjutan akan dikelola dari modul inventori pada sprint berikutnya."}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <label className="flex items-center gap-3 text-sm text-neutral-700">
            <input type="checkbox" className="h-4 w-4 rounded border-neutral-300 text-primary-600" {...form.register("trackInventory")} />
            Lacak stok produk
          </label>

          {mode === "create" ? (
            <Field label="Stok awal" error={form.formState.errors.initialStock?.message}>
              <Input
                type="number"
                min={0}
                hasError={!!form.formState.errors.initialStock}
                {...form.register("initialStock", { valueAsNumber: true })}
              />
            </Field>
          ) : initialProduct ? (
            <div className="rounded-2xl bg-neutral-50 p-4 text-sm text-neutral-700">
              <p className="font-medium text-neutral-950">Ringkasan stok saat ini</p>
              <p className="mt-2">
                Tersedia {initialProduct.stock.quantityAvailable} dari {initialProduct.stock.quantityOnHand} stok fisik.
              </p>
            </div>
          ) : null}

          <label className="flex items-center gap-3 text-sm text-neutral-700">
            <input type="checkbox" className="h-4 w-4 rounded border-neutral-300 text-primary-600" {...form.register("allowBackorder")} />
            Izinkan backorder
          </label>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Visibilitas</CardTitle>
          <CardDescription>Atur status produk sebelum toko publik dibangun pada sprint berikutnya.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 lg:grid-cols-2">
          <Field label="Status" error={form.formState.errors.status?.message}>
            <select
              className="h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
              {...form.register("status")}
            >
              <option value="draft">Draft</option>
              <option value="active">Aktif</option>
              <option value="inactive">Nonaktif</option>
              <option value="archived">Diarsipkan</option>
            </select>
          </Field>

          <label className="flex items-center gap-3 self-end pb-2 text-sm text-neutral-700">
            <input type="checkbox" className="h-4 w-4 rounded border-neutral-300 text-primary-600" {...form.register("isDiscoverable")} />
            Tampilkan di discovery
          </label>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Gambar</CardTitle>
          <CardDescription>
            {mode === "create"
              ? "File akan diunggah setelah produk berhasil dibuat."
              : "Tambah atau hapus gambar yang sudah terhubung ke produk ini."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ImageUploader
            files={imageFiles}
            onFilesChange={setImageFiles}
            existingImages={initialProduct?.images}
            onDeleteExisting={onDeleteImage}
            disabled={isSubmitting}
          />
        </CardContent>
      </Card>

      {initialProduct ? (
        <p className="text-sm text-neutral-500">
          Harga saat ini: <span className="font-medium text-neutral-900">{formatRupiah(initialProduct.price)}</span>. Perubahan
          produk tidak mengubah riwayat order lama.
        </p>
      ) : null}

      {error ? <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p> : null}

      <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
        <Button type="submit" isLoading={isSubmitting}>
          {mode === "create" ? "Simpan produk" : "Simpan perubahan"}
        </Button>
      </div>
    </form>
  );
}

function Field({
  label,
  children,
  error
}: {
  label: string;
  children: React.ReactNode;
  error?: string;
}) {
  return (
    <label className="block space-y-2 text-sm font-medium text-neutral-700">
      {label}
      {children}
      {error ? <span className="block text-sm font-normal text-red-600">{error}</span> : null}
    </label>
  );
}

function toDefaults(product?: ProductDetail): ProductFormValues {
  return {
    categoryId: product?.categoryId ?? "",
    name: product?.name ?? "",
    slug: product?.slug ?? "",
    description: product?.description ?? "",
    sku: product?.sku ?? "",
    barcode: product?.barcode ?? "",
    price: product?.price ?? 0,
    compareAtPrice: product?.compareAtPrice ?? null,
    costPrice: product?.costPrice ?? null,
    weightGram: product?.weightGram ?? 0,
    initialStock: 0,
    status: product?.status ?? "draft",
    isDiscoverable: product?.isDiscoverable ?? false,
    trackInventory: product?.trackInventory ?? true,
    allowBackorder: product?.allowBackorder ?? false
  };
}
