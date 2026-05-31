"use client";

import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";
import { me } from "@/features/auth/api/auth.api";
import { listTenants } from "@/features/tenant/api/tenant.api";
import { useAuthStore } from "@/lib/stores/auth.store";
import { useTenantStore } from "@/lib/stores/tenant.store";

export function AuthRedirectGuard({ children }: { children: ReactNode }) {
  const router = useRouter();
  const accessToken = useAuthStore((state) => state.accessToken);
  const setUser = useAuthStore((state) => state.setUser);
  const clearSession = useAuthStore((state) => state.clearSession);
  const setTenants = useTenantStore((state) => state.setTenants);
  const selectTenant = useTenantStore((state) => state.selectTenant);
  const clearTenant = useTenantStore((state) => state.clearTenant);

  useEffect(() => {
    if (!accessToken) {
      return;
    }

    let cancelled = false;

    async function redirectAuthenticatedUser() {
      try {
        const profile = await me();
        const tenants = await listTenants();

        if (cancelled) {
          return;
        }

        setUser(profile.user);
        setTenants(tenants);

        if (tenants[0]) {
          selectTenant(tenants[0]);
          router.replace("/dashboard");
          return;
        }

        router.replace(profile.user.platformRole === "super_admin" ? "/admin" : "/onboarding/create-store");
      } catch {
        if (!cancelled) {
          clearSession();
          clearTenant();
        }
      }
    }

    void redirectAuthenticatedUser();

    return () => {
      cancelled = true;
    };
  }, [accessToken, clearSession, clearTenant, router, selectTenant, setTenants, setUser]);

  if (accessToken) {
    return (
      <div className="rounded-2xl border border-[#E3D2BC] bg-[#FFFDF8] p-4 text-sm text-[#6F6256]">
        Mengecek sesi akun...
      </div>
    );
  }

  return <>{children}</>;
}
