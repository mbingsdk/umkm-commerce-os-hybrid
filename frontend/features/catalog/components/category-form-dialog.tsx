"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm, useWatch } from "react-hook-form";
import { useEffect } from "react";
import { SlugInput } from "@/components/forms/slug-input";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import type { Category } from "@/features/catalog/types";
import {
  categorySchema,
  type CategoryFormValues
} from "@/features/catalog/schemas/category.schema";

type CategoryFormDialogProps = {
  open: boolean;
  category?: Category | null;
  isSubmitting?: boolean;
  error?: string;
  onClose: () => void;
  onSubmit: (values: CategoryFormValues) => void;
};

export function CategoryFormDialog({
  open,
  category,
  isSubmitting = false,
  error,
  onClose,
  onSubmit
}: CategoryFormDialogProps) {
  const form = useForm<CategoryFormValues>({
    resolver: zodResolver(categorySchema),
    defaultValues: toDefaults(category)
  });
  const name = useWatch({ control: form.control, name: "name" });

  useEffect(() => {
    form.reset(toDefaults(category));
  }, [category, form, open]);

  function handleSubmit(values: CategoryFormValues) {
    if (isSubmitting) {
      return;
    }

    onSubmit(values);
  }

  return (
    <Dialog
      open={open}
      title={category ? "Edit kategori" : "Tambah kategori"}
      description="Kategori membantu produk lebih mudah dikelola di dashboard."
      onClose={onClose}
      footer={
        <>
          <Button type="button" variant="outline" onClick={onClose}>
            Batal
          </Button>
          <Button type="submit" form="category-form" isLoading={isSubmitting} disabled={isSubmitting}>
            {category ? "Simpan perubahan" : "Buat kategori"}
          </Button>
        </>
      }
    >
      <form id="category-form" className="space-y-4" onSubmit={form.handleSubmit(handleSubmit)}>
        <Field label="Nama kategori" error={form.formState.errors.name?.message}>
          <Input placeholder="Bouquet" hasError={!!form.formState.errors.name} {...form.register("name")} />
        </Field>

        <Field label="Slug" error={form.formState.errors.slug?.message}>
          <Controller
            control={form.control}
            name="slug"
            render={({ field }) => (
              <SlugInput
                key={category?.id ?? "new"}
                sourceValue={name}
                value={field.value}
                onChange={field.onChange}
                hasError={!!form.formState.errors.slug}
                placeholder="bouquet"
              />
            )}
          />
        </Field>

        <Field label="Deskripsi" error={form.formState.errors.description?.message}>
          <textarea
            className="min-h-24 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-sm text-neutral-950 outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            placeholder="Aneka bouquet untuk hadiah"
            {...form.register("description")}
          />
        </Field>

        <Field label="Urutan" error={form.formState.errors.sortOrder?.message}>
          <Input
            type="number"
            min={0}
            hasError={!!form.formState.errors.sortOrder}
            {...form.register("sortOrder", { valueAsNumber: true })}
          />
        </Field>

        <label className="flex items-center gap-3 text-sm text-neutral-700">
          <input type="checkbox" className="h-4 w-4 rounded border-neutral-300 text-primary-600" {...form.register("isActive")} />
          Kategori aktif
        </label>

        {error ? <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p> : null}
      </form>
    </Dialog>
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

function toDefaults(category?: Category | null): CategoryFormValues {
  return {
    name: category?.name ?? "",
    slug: category?.slug ?? "",
    description: category?.description ?? "",
    sortOrder: category?.sortOrder ?? 0,
    isActive: category?.isActive ?? true
  };
}
