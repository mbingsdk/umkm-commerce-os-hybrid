"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useState, type ReactNode } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { createStoreOnboarding } from "@/features/onboarding/api/onboarding.api";
import { useAuthStore } from "@/lib/stores/auth.store";
import { useTenantStore } from "@/lib/stores/tenant.store";

const slugPattern = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

const createStoreSchema = z.object({
  tenantName: z.string().trim().min(1, "Nama tenant wajib diisi."),
  tenantSlug: z.string().trim().regex(slugPattern, "Slug tenant hanya boleh huruf kecil, angka, dan tanda hubung."),
  storeName: z.string().trim().min(1, "Nama toko wajib diisi."),
  storeSlug: z.string().trim().regex(slugPattern, "Slug toko hanya boleh huruf kecil, angka, dan tanda hubung."),
  description: z.string().trim().optional(),
  whatsapp: z.string().trim().optional(),
  city: z.string().trim().optional(),
  province: z.string().trim().optional()
});

type CreateStoreFormValues = z.infer<typeof createStoreSchema>;

const stepFields: Array<Array<keyof CreateStoreFormValues>> = [
  ["tenantName", "tenantSlug", "storeName", "storeSlug"],
  ["description", "whatsapp", "city", "province"]
];

export function CreateStoreWizard() {
  const router = useRouter();
  const accessToken = useAuthStore((state) => state.accessToken);
  const upsertTenant = useTenantStore((state) => state.upsertTenant);
  const selectTenant = useTenantStore((state) => state.selectTenant);
  const [step, setStep] = useState(0);

  const form = useForm<CreateStoreFormValues>({
    resolver: zodResolver(createStoreSchema),
    defaultValues: {
      tenantName: "",
      tenantSlug: "",
      storeName: "",
      storeSlug: "",
      description: "",
      whatsapp: "",
      city: "",
      province: ""
    }
  });

  const createStoreMutation = useMutation({
    mutationFn: (values: CreateStoreFormValues) =>
      createStoreOnboarding({
        tenant_name: values.tenantName,
        tenant_slug: values.tenantSlug,
        store: {
          name: values.storeName,
          slug: values.storeSlug,
          description: values.description,
          whatsapp: values.whatsapp,
          city: values.city,
          province: values.province
        }
      }),
    onSuccess: (tenant) => {
      upsertTenant(tenant);
      selectTenant(tenant);
      router.push("/dashboard");
    }
  });

  async function goNext() {
    if (createStoreMutation.isPending) {
      return;
    }

    const valid = await form.trigger(stepFields[step]);
    if (valid) {
      setStep((current) => Math.min(current + 1, 2));
    }
  }

  function handleSubmit(values: CreateStoreFormValues) {
    if (createStoreMutation.isPending) {
      return;
    }

    createStoreMutation.mutate(values);
  }

  if (!accessToken) {
    return (
      <div className="space-y-4">
        <p className="text-sm leading-6 text-neutral-600">
          Kamu perlu login lebih dulu sebelum membuat toko.
        </p>
        <Button onClick={() => router.push("/login")}>Ke halaman login</Button>
      </div>
    );
  }

  return (
    <form className="space-y-6" onSubmit={form.handleSubmit(handleSubmit)}>
      <ol className="grid grid-cols-3 gap-2 text-xs font-semibold text-neutral-500">
        {["Identitas", "Profil singkat", "Tinjau"].map((label, index) => (
          <li
            key={label}
            className={
              index <= step
                ? "rounded-xl bg-primary-100 px-3 py-2 text-primary-800"
                : "rounded-xl bg-neutral-100 px-3 py-2"
            }
          >
            {index + 1}. {label}
          </li>
        ))}
      </ol>

      {step === 0 ? (
        <div className="space-y-4">
          <WizardField label="Nama tenant" error={form.formState.errors.tenantName?.message}>
            <Input placeholder="Toko Bunga Ayu" hasError={!!form.formState.errors.tenantName} {...form.register("tenantName")} />
          </WizardField>
          <WizardField label="Slug tenant" error={form.formState.errors.tenantSlug?.message}>
            <Input placeholder="toko-bunga-ayu" hasError={!!form.formState.errors.tenantSlug} {...form.register("tenantSlug")} />
          </WizardField>
          <WizardField label="Nama toko" error={form.formState.errors.storeName?.message}>
            <Input placeholder="Toko Bunga Ayu" hasError={!!form.formState.errors.storeName} {...form.register("storeName")} />
          </WizardField>
          <WizardField label="Slug toko" error={form.formState.errors.storeSlug?.message}>
            <Input placeholder="toko-bunga-ayu" hasError={!!form.formState.errors.storeSlug} {...form.register("storeSlug")} />
          </WizardField>
        </div>
      ) : null}

      {step === 1 ? (
        <div className="space-y-4">
          <WizardField label="Deskripsi singkat" error={form.formState.errors.description?.message}>
            <Input placeholder="Bouquet dan hampers lokal Makassar" {...form.register("description")} />
          </WizardField>
          <WizardField label="WhatsApp" error={form.formState.errors.whatsapp?.message}>
            <Input placeholder="08123456789" {...form.register("whatsapp")} />
          </WizardField>
          <WizardField label="Kota" error={form.formState.errors.city?.message}>
            <Input placeholder="Makassar" {...form.register("city")} />
          </WizardField>
          <WizardField label="Provinsi" error={form.formState.errors.province?.message}>
            <Input placeholder="Sulawesi Selatan" {...form.register("province")} />
          </WizardField>
        </div>
      ) : null}

      {step === 2 ? (
        <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-4 text-sm text-neutral-700">
          <p className="font-semibold text-neutral-950">Tinjau sebelum membuat toko</p>
          <dl className="mt-4 grid gap-3">
            <ReviewRow label="Tenant" value={`${form.getValues("tenantName")} (${form.getValues("tenantSlug")})`} />
            <ReviewRow label="Toko" value={`${form.getValues("storeName")} (${form.getValues("storeSlug")})`} />
            <ReviewRow label="Kota" value={form.getValues("city") || "Belum diisi"} />
            <ReviewRow label="WhatsApp" value={form.getValues("whatsapp") || "Belum diisi"} />
          </dl>
        </div>
      ) : null}

      {createStoreMutation.isError ? (
        <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">
          {createStoreMutation.error instanceof Error
            ? createStoreMutation.error.message
            : "Toko belum berhasil dibuat. Coba lagi."}
        </p>
      ) : null}

      <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-between">
        <Button
          type="button"
          variant="outline"
          disabled={step === 0 || createStoreMutation.isPending}
          onClick={() => setStep((current) => Math.max(current - 1, 0))}
        >
          Kembali
        </Button>

        {step < 2 ? (
          <Button type="button" disabled={createStoreMutation.isPending} onClick={goNext}>
            Lanjut
          </Button>
        ) : (
          <Button type="submit" isLoading={createStoreMutation.isPending} disabled={createStoreMutation.isPending}>
            Buat toko
          </Button>
        )}
      </div>
    </form>
  );
}

function WizardField({
  label,
  children,
  error
}: {
  label: string;
  children: ReactNode;
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

function ReviewRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid gap-1 sm:grid-cols-[120px_1fr]">
      <dt className="text-neutral-500">{label}</dt>
      <dd className="font-medium text-neutral-900">{value}</dd>
    </div>
  );
}
