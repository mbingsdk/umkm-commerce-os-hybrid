"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Badge } from "@/components/ui/badge";
import {
  getCurrentStore,
  publishCurrentStore,
  unpublishCurrentStore,
  updateBusinessHours,
  updateCurrentStore
} from "@/features/settings/api/store.api";
import { StoreBusinessHoursForm } from "@/features/settings/components/store-business-hours-form";
import { StoreProfileForm } from "@/features/settings/components/store-profile-form";
import { StorePublishCard } from "@/features/settings/components/store-publish-card";
import type {
  BusinessHoursFormValues,
  StoreProfileFormValues
} from "@/features/settings/schemas/store.schema";
import { queryKeys } from "@/lib/api/query-keys";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";
import { useToastStore } from "@/lib/stores/toast.store";

const siteBaseUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";

export function StoreSettingsPage() {
  const queryClient = useQueryClient();
  const pushToast = useToastStore((state) => state.pushToast);
  const selectedTenantId = useTenantStore((state) => state.selectedTenantId);
  const userPermissions = useTenantStore((state) => state.permissions);
  const canRead = userPermissions.includes(permissions.storeRead);
  const canUpdate = userPermissions.includes(permissions.storeUpdate);
  const canPublish = userPermissions.includes(permissions.storePublish);
  const canUpdateBusinessHours = userPermissions.includes(permissions.storeUpdateBusinessHours);
  const [businessHours, setBusinessHours] = useState<BusinessHoursFormValues>(createEmptyBusinessHours());

  const storeQuery = useQuery({
    queryKey: queryKeys.currentStore(selectedTenantId),
    queryFn: getCurrentStore,
    enabled: !!selectedTenantId && canRead
  });

  const updateProfileMutation = useMutation({
    mutationFn: (values: StoreProfileFormValues) =>
      updateCurrentStore({
        name: values.name,
        description: values.description ?? "",
        phone: values.phone ?? "",
        whatsapp: values.whatsapp ?? "",
        email: values.email ?? "",
        address: values.address ?? "",
        city: values.city ?? "",
        province: values.province ?? "",
        postal_code: values.postalCode ?? "",
        is_discoverable: values.isDiscoverable
      }),
    onSuccess: (store) => {
      queryClient.setQueryData(queryKeys.currentStore(selectedTenantId), store);
      pushToast({ tone: "success", title: "Profil toko disimpan" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Profil toko gagal disimpan", description: getErrorMessage(error) });
    }
  });

  const publishMutation = useMutation({
    mutationFn: publishCurrentStore,
    onSuccess: (store) => {
      queryClient.setQueryData(queryKeys.currentStore(selectedTenantId), store);
      pushToast({ tone: "success", title: "Toko berhasil dipublish" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Toko gagal dipublish", description: getErrorMessage(error) });
    }
  });

  const unpublishMutation = useMutation({
    mutationFn: unpublishCurrentStore,
    onSuccess: (store) => {
      queryClient.setQueryData(queryKeys.currentStore(selectedTenantId), store);
      pushToast({ tone: "success", title: "Toko berhasil di-unpublish" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Toko gagal di-unpublish", description: getErrorMessage(error) });
    }
  });

  const businessHoursMutation = useMutation({
    mutationFn: (values: BusinessHoursFormValues) =>
      updateBusinessHours({
        items: values.items.map((item) => ({
          day_of_week: item.dayOfWeek,
          open_time: item.isClosed ? "" : item.openTime ?? "",
          close_time: item.isClosed ? "" : item.closeTime ?? "",
          is_closed: item.isClosed
        }))
      }),
    onSuccess: (result) => {
      setBusinessHours({
        items: result.items.map((item) => ({
          dayOfWeek: item.day_of_week,
          openTime: item.open_time ?? "",
          closeTime: item.close_time ?? "",
          isClosed: item.is_closed
        }))
      });
      pushToast({ tone: "success", title: "Jam operasional disimpan" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Jam operasional gagal disimpan", description: getErrorMessage(error) });
    }
  });

  if (!canRead) {
    return (
      <EmptyState
        title="Akses toko belum tersedia"
        description="Role aktifmu belum memiliki izin untuk melihat pengaturan toko."
      />
    );
  }

  if (!selectedTenantId) {
    return (
      <EmptyState
        title="Tenant belum dipilih"
        description="Pilih tenant aktif terlebih dahulu sebelum mengatur profil toko."
      />
    );
  }

  if (storeQuery.isPending) {
    return <LoadingState lines={5} />;
  }

  if (storeQuery.isError) {
    return (
      <ErrorState
        title="Pengaturan toko gagal dimuat"
        description="Pastikan tenant aktif masih tersedia lalu coba muat ulang."
        onRetry={() => void storeQuery.refetch()}
      />
    );
  }

  if (!storeQuery.data) {
    return (
      <EmptyState
        title="Toko belum tersedia"
        description="Buat toko melalui onboarding sebelum mengubah pengaturan toko."
      />
    );
  }

  const store = storeQuery.data;
  const publicUrl = buildStorePublicUrl(store.slug);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <Badge tone="primary">Pengaturan toko</Badge>
          <h1 className="mt-3 text-2xl font-bold text-neutral-950">Toko</h1>
          <p className="mt-1 max-w-2xl text-sm leading-6 text-neutral-500">
            Atur profil, kontak, publish, dan jam operasional dasar untuk storefront publik.
          </p>
        </div>
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.25fr_0.75fr]">
        <StoreProfileForm
          store={store}
          canUpdate={canUpdate}
          isSubmitting={updateProfileMutation.isPending}
          error={updateProfileMutation.isError ? getErrorMessage(updateProfileMutation.error) : undefined}
          onSubmit={(values) => updateProfileMutation.mutate(values)}
        />

        <div className="space-y-6">
          <StorePublishCard
            store={store}
            publicUrl={publicUrl}
            canPublish={canPublish}
            isSubmitting={publishMutation.isPending || unpublishMutation.isPending}
            onPublish={() => publishMutation.mutate()}
            onUnpublish={() => unpublishMutation.mutate()}
          />
          <StoreBusinessHoursForm
            initialValues={businessHours}
            canUpdate={canUpdateBusinessHours}
            isSubmitting={businessHoursMutation.isPending}
            error={businessHoursMutation.isError ? getErrorMessage(businessHoursMutation.error) : undefined}
            onSubmit={(values) => businessHoursMutation.mutate(values)}
          />
        </div>
      </div>
    </div>
  );
}

function buildStorePublicUrl(storeSlug: string) {
  return new URL(`/s/${storeSlug}`, siteBaseUrl).toString();
}

function createEmptyBusinessHours(): BusinessHoursFormValues {
  return {
    items: Array.from({ length: 7 }, (_, index) => ({
      dayOfWeek: index + 1,
      openTime: "",
      closeTime: "",
      isClosed: true
    }))
  };
}

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : "Permintaan gagal diproses.";
}
