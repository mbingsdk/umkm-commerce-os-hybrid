"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import type { Store } from "@/features/settings/api/store.api";
import {
  storeProfileSchema,
  type StoreProfileFormValues
} from "@/features/settings/schemas/store.schema";

type StoreProfileFormProps = {
  store: Store;
  canUpdate: boolean;
  isSubmitting?: boolean;
  error?: string;
  onSubmit: (values: StoreProfileFormValues) => void;
};

export function StoreProfileForm({
  store,
  canUpdate,
  isSubmitting = false,
  error,
  onSubmit
}: StoreProfileFormProps) {
  const form = useForm<StoreProfileFormValues>({
    resolver: zodResolver(storeProfileSchema),
    defaultValues: toDefaults(store)
  });

  useEffect(() => {
    form.reset(toDefaults(store));
  }, [form, store]);

  function handleSubmit(values: StoreProfileFormValues) {
    if (!canUpdate || isSubmitting) {
      return;
    }

    onSubmit(values);
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Profil toko</CardTitle>
        <CardDescription>
          Kelola informasi yang dipakai untuk dashboard dan tampilan publik toko. Slug belum dapat diubah dari UI ini.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form className="space-y-6" onSubmit={form.handleSubmit(handleSubmit)}>
          <div className="grid gap-4 lg:grid-cols-[1fr_220px]">
            <Field label="Nama toko" error={form.formState.errors.name?.message}>
              <Input
                disabled={!canUpdate || isSubmitting}
                hasError={!!form.formState.errors.name}
                placeholder="Toko Bunga Ayu"
                {...form.register("name")}
              />
            </Field>

            <Field label="Slug toko">
              <Input disabled value={store.slug} />
            </Field>
          </div>

          <Field label="Deskripsi singkat" error={form.formState.errors.description?.message}>
            <textarea
              className="min-h-24 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-sm text-neutral-950 outline-none transition placeholder:text-neutral-400 focus:border-primary-500 focus:ring-2 focus:ring-primary-100 disabled:bg-neutral-50 disabled:text-neutral-500"
              disabled={!canUpdate || isSubmitting}
              placeholder="Ceritakan toko, produk unggulan, atau area layanan."
              {...form.register("description")}
            />
          </Field>

          <section className="space-y-4">
            <div>
              <h3 className="text-sm font-semibold text-neutral-950">Kontak toko</h3>
              <p className="mt-1 text-sm text-neutral-500">Minimal isi telepon atau WhatsApp sebelum publish.</p>
            </div>
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              <Field label="Telepon" error={form.formState.errors.phone?.message}>
                <Input
                  disabled={!canUpdate || isSubmitting}
                  hasError={!!form.formState.errors.phone}
                  placeholder="0411..."
                  {...form.register("phone")}
                />
              </Field>
              <Field label="WhatsApp" error={form.formState.errors.whatsapp?.message}>
                <Input
                  disabled={!canUpdate || isSubmitting}
                  hasError={!!form.formState.errors.whatsapp}
                  placeholder="08123456789"
                  {...form.register("whatsapp")}
                />
              </Field>
              <Field label="Email" error={form.formState.errors.email?.message}>
                <Input
                  disabled={!canUpdate || isSubmitting}
                  hasError={!!form.formState.errors.email}
                  placeholder="toko@email.com"
                  {...form.register("email")}
                />
              </Field>
            </div>
          </section>

          <section className="space-y-4">
            <div>
              <h3 className="text-sm font-semibold text-neutral-950">Alamat/lokasi dasar</h3>
              <p className="mt-1 text-sm text-neutral-500">Kota dipakai untuk discovery dan informasi storefront.</p>
            </div>
            <Field label="Alamat" error={form.formState.errors.address?.message}>
              <Input
                disabled={!canUpdate || isSubmitting}
                hasError={!!form.formState.errors.address}
                placeholder="Jl. ..."
                {...form.register("address")}
              />
            </Field>
            <div className="grid gap-4 sm:grid-cols-3">
              <Field label="Kota" error={form.formState.errors.city?.message}>
                <Input
                  disabled={!canUpdate || isSubmitting}
                  hasError={!!form.formState.errors.city}
                  placeholder="Makassar"
                  {...form.register("city")}
                />
              </Field>
              <Field label="Provinsi" error={form.formState.errors.province?.message}>
                <Input
                  disabled={!canUpdate || isSubmitting}
                  hasError={!!form.formState.errors.province}
                  placeholder="Sulawesi Selatan"
                  {...form.register("province")}
                />
              </Field>
              <Field label="Kode pos" error={form.formState.errors.postalCode?.message}>
                <Input
                  disabled={!canUpdate || isSubmitting}
                  hasError={!!form.formState.errors.postalCode}
                  placeholder="90111"
                  {...form.register("postalCode")}
                />
              </Field>
            </div>
          </section>

          <label className="flex items-start gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 p-4 text-sm text-neutral-600">
            <input
              type="checkbox"
              className="mt-1 h-4 w-4 rounded border-neutral-300 text-primary-600 focus:ring-primary-500"
              disabled={!canUpdate || isSubmitting}
              {...form.register("isDiscoverable")}
            />
            <span>
              <span className="block font-semibold text-neutral-900">Tampilkan di discovery publik</span>
              Jika aktif dan toko sudah publish, toko/produk eligible dapat muncul di halaman discovery platform.
            </span>
          </label>

          {error ? <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{error}</p> : null}

          <div className="flex justify-end">
            <Button type="submit" isLoading={isSubmitting} disabled={!canUpdate || isSubmitting}>
              Simpan profil toko
            </Button>
          </div>

          {!canUpdate ? (
            <p className="rounded-xl bg-amber-50 p-3 text-sm text-amber-700">
              Role aktifmu hanya bisa melihat profil toko, belum bisa mengubahnya.
            </p>
          ) : null}
        </form>
      </CardContent>
    </Card>
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

function toDefaults(store: Store): StoreProfileFormValues {
  return {
    name: store.name,
    description: store.description ?? "",
    phone: store.phone ?? "",
    whatsapp: store.whatsapp ?? "",
    email: store.email ?? "",
    address: store.address ?? "",
    city: store.city ?? "",
    province: store.province ?? "",
    postalCode: store.postalCode ?? "",
    isDiscoverable: store.isDiscoverable
  };
}
