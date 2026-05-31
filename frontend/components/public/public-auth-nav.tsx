"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo } from "react";
import { me, logout as logoutApi } from "@/features/auth/api/auth.api";
import { listTenants } from "@/features/tenant/api/tenant.api";
import { useAuthStore } from "@/lib/stores/auth.store";
import { useTenantStore } from "@/lib/stores/tenant.store";

export function PublicAuthNav() {
  const router = useRouter();
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

  const activeTenant = useMemo(
    () => tenants.find((tenant) => tenant.id === selectedTenantId) ?? tenants[0],
    [selectedTenantId, tenants]
  );

  useEffect(() => {
    if (!accessToken) {
      return;
    }

    let cancelled = false;

    async function verifySession() {
      try {
        const profile = await me();
        const memberships = await listTenants();

        if (cancelled) {
          return;
        }

        setUser(profile.user);
        setTenants(memberships);

        const currentTenantId = useTenantStore.getState().selectedTenantId;
        if (!currentTenantId && memberships[0]) {
          selectTenant(memberships[0]);
        }
      } catch {
        if (!cancelled) {
          clearSession();
          clearTenant();
        }
      }
    }

    void verifySession();

    return () => {
      cancelled = true;
    };
  }, [accessToken, clearSession, clearTenant, selectTenant, setTenants, setUser]);

  async function handleLogout() {
    if (refreshToken) {
      try {
        await logoutApi(refreshToken);
      } catch {
        // Logout must still clear local session even if the API token is already expired.
      }
    }

    clearSession();
    clearTenant();
    router.push("/");
  }

  if (!accessToken || !user) {
    return (
      <div className="hidden shrink-0 items-center gap-2 sm:flex">
        <Link
          href="/login"
          className="hidden px-2.5 py-2 text-sm font-semibold text-[#6F6256] transition hover:text-[#251F1A] sm:inline-flex"
        >
          Masuk
        </Link>
        <Link
          href="/register"
          className="inline-flex h-9 items-center justify-center rounded-full bg-[#251F1A] px-3.5 text-sm font-semibold text-[#FFFDF8] shadow-sm transition hover:bg-[#16110E]"
        >
          Daftar Toko
        </Link>
      </div>
    );
  }

  if (activeTenant) {
    return (
      <div className="hidden shrink-0 items-center gap-2 sm:flex">
        <Link
          href="/dashboard"
          className="inline-flex h-9 items-center justify-center rounded-full bg-[#251F1A] px-3.5 text-sm font-semibold text-[#FFFDF8] transition hover:bg-[#16110E]"
        >
          Dashboard
        </Link>
        <Link
          href={`/s/${activeTenant.store.slug}`}
          className="hidden h-9 items-center justify-center rounded-full border border-[#E3D2BC] bg-white px-3 text-sm font-semibold text-[#6F6256] transition hover:border-[#B96E45] hover:text-[#251F1A] sm:inline-flex"
        >
          Toko Saya
        </Link>
        <button
          className="hidden h-9 items-center justify-center rounded-full px-2.5 text-sm font-semibold text-[#6F6256] transition hover:text-[#251F1A] md:inline-flex"
          onClick={() => void handleLogout()}
          type="button"
        >
          Keluar
        </button>
      </div>
    );
  }

  return (
    <div className="hidden shrink-0 items-center gap-2 sm:flex">
      <Link
        href={user.platformRole === "super_admin" ? "/admin" : "/onboarding/create-store"}
        className="inline-flex h-9 items-center justify-center rounded-full bg-[#251F1A] px-3.5 text-sm font-semibold text-[#FFFDF8] transition hover:bg-[#16110E]"
      >
        {user.platformRole === "super_admin" ? "Admin" : "Buat toko"}
      </Link>
      <button
        className="hidden h-9 items-center justify-center rounded-full px-2.5 text-sm font-semibold text-[#6F6256] transition hover:text-[#251F1A] sm:inline-flex"
        onClick={() => void handleLogout()}
        type="button"
      >
        Keluar
      </button>
    </div>
  );
}
