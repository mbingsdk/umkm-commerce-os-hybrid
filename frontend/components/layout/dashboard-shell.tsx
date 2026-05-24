"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, useSyncExternalStore, type ReactNode } from "react";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { DashboardMobileNav } from "@/components/layout/dashboard-mobile-nav";
import { DashboardSidebar } from "@/components/layout/dashboard-sidebar";
import { DashboardTopbar } from "@/components/layout/dashboard-topbar";
import { Button } from "@/components/ui/button";
import { logout, me } from "@/features/auth/api/auth.api";
import { getCurrentStore } from "@/features/settings/api/store.api";
import { listTenants } from "@/features/tenant/api/tenant.api";
import { queryKeys } from "@/lib/api/query-keys";
import { useAuthStore } from "@/lib/stores/auth.store";
import { useTenantStore } from "@/lib/stores/tenant.store";

export function DashboardShell({ children }: { children: ReactNode }) {
  const router = useRouter();
  const hydrated = usePersistedStoresHydrated();
  const accessToken = useAuthStore((state) => state.accessToken);
  const refreshToken = useAuthStore((state) => state.refreshToken);
  const user = useAuthStore((state) => state.user);
  const setUser = useAuthStore((state) => state.setUser);
  const clearSession = useAuthStore((state) => state.clearSession);
  const tenants = useTenantStore((state) => state.tenants);
  const selectedTenantId = useTenantStore((state) => state.selectedTenantId);
  const setTenants = useTenantStore((state) => state.setTenants);
  const selectTenant = useTenantStore((state) => state.selectTenant);
  const clearTenant = useTenantStore((state) => state.clearTenant);
  const shouldLoadMe = hydrated && !!accessToken && !user;

  useEffect(() => {
    if (hydrated && !accessToken) {
      router.replace("/login");
    }
  }, [accessToken, hydrated, router]);

  const meQuery = useQuery({
    queryKey: queryKeys.me,
    queryFn: me,
    enabled: shouldLoadMe
  });

  const tenantsQuery = useQuery({
    queryKey: queryKeys.tenants,
    queryFn: listTenants,
    enabled: hydrated && !!accessToken
  });

  useEffect(() => {
    if (meQuery.data?.user) {
      setUser(meQuery.data.user);
    }
  }, [meQuery.data, setUser]);

  useEffect(() => {
    if (!tenantsQuery.data) {
      return;
    }

    setTenants(tenantsQuery.data);

    const selectedTenant = tenantsQuery.data.find((tenant) => tenant.id === selectedTenantId);
    if (selectedTenant) {
      selectTenant(selectedTenant);
      return;
    }

    if (tenantsQuery.data[0]) {
      selectTenant(tenantsQuery.data[0]);
    }
  }, [selectTenant, selectedTenantId, setTenants, tenantsQuery.data]);

  const currentStoreQuery = useQuery({
    queryKey: queryKeys.currentStore(selectedTenantId),
    queryFn: getCurrentStore,
    enabled: hydrated && !!accessToken && !!selectedTenantId
  });

  const logoutMutation = useMutation({
    mutationFn: async () => {
      if (refreshToken) {
        await logout(refreshToken);
      }
    },
    onSettled: () => {
      clearSession();
      clearTenant();
      router.replace("/login");
    }
  });

  if (!hydrated || (accessToken && (tenantsQuery.isPending || (shouldLoadMe && meQuery.isPending)))) {
    return (
      <main className="min-h-screen bg-neutral-50 p-4 sm:p-6 lg:p-8">
        <LoadingState lines={3} />
      </main>
    );
  }

  if (!accessToken) {
    return null;
  }

  if (tenantsQuery.isError || (shouldLoadMe && meQuery.isError)) {
    return (
      <main className="min-h-screen bg-neutral-50 p-4 sm:p-6 lg:p-8">
        <ErrorState
          title="Sesi belum bisa dimuat"
          description="Kami belum bisa mengambil data akun atau tenant. Coba muat ulang."
          onRetry={() => {
            void tenantsQuery.refetch();
            void meQuery.refetch();
          }}
        />
      </main>
    );
  }

  if (tenants.length === 0) {
    return (
      <main className="min-h-screen bg-neutral-50 p-4 sm:p-6 lg:p-8">
        <EmptyState
          title="Belum ada toko"
          description="Mulai dengan membuat tenant dan toko pertama agar dashboard bisa digunakan."
          action={<Button onClick={() => router.push("/onboarding/create-store")}>Buat toko pertama</Button>}
        />
      </main>
    );
  }

  if (currentStoreQuery.isPending) {
    return (
      <main className="min-h-screen bg-neutral-50 p-4 sm:p-6 lg:p-8">
        <LoadingState lines={3} />
      </main>
    );
  }

  if (currentStoreQuery.isError) {
    return (
      <main className="min-h-screen bg-neutral-50 p-4 sm:p-6 lg:p-8">
        <ErrorState
          title="Toko aktif belum bisa dimuat"
          description="Pastikan tenant masih aktif lalu coba lagi."
          onRetry={() => {
            void currentStoreQuery.refetch();
          }}
        />
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-neutral-50">
      <div className="mx-auto flex min-h-screen max-w-7xl">
        <DashboardSidebar />

        <div className="flex min-w-0 flex-1 flex-col">
          <DashboardTopbar
            currentStore={currentStoreQuery.data}
            onLogout={() => logoutMutation.mutate()}
            isLoggingOut={logoutMutation.isPending}
          />
          <DashboardMobileNav />

          <div className="flex-1 px-4 py-6 sm:px-6 lg:px-8">{children}</div>
        </div>
      </div>
    </main>
  );
}

function usePersistedStoresHydrated() {
  return useSyncExternalStore(
    (onStoreChange) => {
      const stopAuthHydrate = useAuthStore.persist.onHydrate(onStoreChange);
      const stopAuthFinish = useAuthStore.persist.onFinishHydration(onStoreChange);
      const stopTenantHydrate = useTenantStore.persist.onHydrate(onStoreChange);
      const stopTenantFinish = useTenantStore.persist.onFinishHydration(onStoreChange);

      return () => {
        stopAuthHydrate();
        stopAuthFinish();
        stopTenantHydrate();
        stopTenantFinish();
      };
    },
    () => useAuthStore.persist.hasHydrated() && useTenantStore.persist.hasHydrated(),
    () => false
  );
}
