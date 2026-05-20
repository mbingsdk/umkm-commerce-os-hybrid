"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { openPOSSession } from "@/features/pos/api/pos.api";
import { OpenSessionForm } from "@/features/pos/components/open-session-form";
import { POSRegisterScreen } from "@/features/pos/components/pos-register-screen";
import { useCurrentPOSSession } from "@/features/pos/hooks/use-pos";
import { ApiError } from "@/lib/api/errors";
import { queryKeys } from "@/lib/api/query-keys";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";
import { useToastStore } from "@/lib/stores/toast.store";

export default function POSPage() {
  const queryClient = useQueryClient();
  const pushToast = useToastStore((state) => state.pushToast);
  const tenantId = useTenantStore((state) => state.selectedTenantId);
  const userPermissions = useTenantStore((state) => state.permissions);
  const canReadSession = userPermissions.includes(permissions.posReadSession);
  const canOpenSession = userPermissions.includes(permissions.posOpenSession);
  const currentSessionQuery = useCurrentPOSSession(canReadSession);
  const noOpenSession =
    currentSessionQuery.isError &&
    currentSessionQuery.error instanceof ApiError &&
    currentSessionQuery.error.code === "NOT_FOUND";

  const openMutation = useMutation({
    mutationFn: openPOSSession,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.posCurrentSession(tenantId) });
      pushToast({ tone: "success", title: "Sesi POS dibuka" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Gagal membuka sesi POS", description: error.message });
    }
  });

  if (!canReadSession) {
    return (
      <EmptyState
        title="Akses POS belum tersedia"
        description="Role aktifmu belum memiliki izin untuk membuka halaman POS."
      />
    );
  }

  if (currentSessionQuery.isPending) {
    return <LoadingState lines={5} />;
  }

  if (currentSessionQuery.isError && !noOpenSession) {
    return (
      <ErrorState
        title="Sesi POS gagal dimuat"
        description="Coba muat ulang sebelum membuka transaksi POS."
        onRetry={() => void currentSessionQuery.refetch()}
      />
    );
  }

  if (!currentSessionQuery.data || noOpenSession) {
    if (!canOpenSession) {
      return (
        <EmptyState
          title="Belum ada sesi POS aktif"
          description="Role aktifmu belum memiliki izin untuk membuka sesi kasir baru."
        />
      );
    }

    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-semibold text-neutral-950">POS</h1>
          <p className="mt-1 text-sm text-neutral-500">Buka sesi kasir untuk mulai transaksi online-first.</p>
        </div>
        <OpenSessionForm
          isSubmitting={openMutation.isPending}
          error={openMutation.isError ? openMutation.error.message : undefined}
          onSubmit={(values) => openMutation.mutate(values)}
        />
      </div>
    );
  }

  return <POSRegisterScreen session={currentSessionQuery.data} />;
}
