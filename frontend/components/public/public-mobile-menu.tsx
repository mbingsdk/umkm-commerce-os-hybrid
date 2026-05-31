"use client";

import Link from "next/link";
import { Menu, X } from "lucide-react";
import { usePathname, useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import { logout as logoutApi } from "@/features/auth/api/auth.api";
import { useAuthStore } from "@/lib/stores/auth.store";
import { useTenantStore } from "@/lib/stores/tenant.store";
import { cn } from "@/lib/utils/cn";

const navItems = [
  { label: "Produk", href: "/products" },
  { label: "Toko", href: "/stores" },
  { label: "Harga", href: "/pricing" },
  { label: "Cari", href: "/search" }
];

export function PublicMobileMenu() {
  const router = useRouter();
  const pathname = usePathname();
  const [open, setOpen] = useState(false);
  const accessToken = useAuthStore((state) => state.accessToken);
  const refreshToken = useAuthStore((state) => state.refreshToken);
  const user = useAuthStore((state) => state.user);
  const clearSession = useAuthStore((state) => state.clearSession);
  const tenants = useTenantStore((state) => state.tenants);
  const selectedTenantId = useTenantStore((state) => state.selectedTenantId);
  const clearTenant = useTenantStore((state) => state.clearTenant);

  const activeTenant = useMemo(
    () => tenants.find((tenant) => tenant.id === selectedTenantId) ?? tenants[0],
    [selectedTenantId, tenants]
  );
  const isLoggedIn = Boolean(accessToken && user);

  async function handleLogout() {
    if (refreshToken) {
      try {
        await logoutApi(refreshToken);
      } catch {
        // Local logout still needs to succeed when the server session is already expired.
      }
    }

    clearSession();
    clearTenant();
    setOpen(false);
    router.push("/");
  }

  return (
    <div className="relative md:hidden">
      <button
        aria-expanded={open}
        aria-label={open ? "Tutup menu" : "Buka menu"}
        className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-[#E3D2BC] bg-white text-[#251F1A] shadow-sm"
        onClick={() => setOpen((value) => !value)}
        type="button"
      >
        {open ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
      </button>

      {open ? (
        <div className="absolute right-0 top-11 z-50 w-[min(88vw,320px)] overflow-hidden rounded-3xl border border-[#E3D2BC] bg-[#FFFDF8] p-2 shadow-[0_18px_50px_rgba(80,57,34,0.16)]">
          <nav aria-label="Menu publik mobile" className="grid gap-1">
            {navItems.map((item) => (
              <Link
                className={cn(
                  "rounded-2xl px-3 py-2.5 text-sm font-semibold transition",
                  pathname === item.href
                    ? "bg-[#251F1A] text-[#FFFDF8]"
                    : "text-[#6F6256] hover:bg-[#F1E7D8] hover:text-[#251F1A]"
                )}
                href={item.href}
                key={item.href}
                onClick={() => setOpen(false)}
              >
                {item.label}
              </Link>
            ))}
          </nav>

          <div className="mt-2 border-t border-[#E3D2BC] pt-2">
            {!isLoggedIn ? (
              <div className="grid gap-1">
                <Link
                  className="rounded-2xl px-3 py-2.5 text-sm font-semibold text-[#6F6256] hover:bg-[#F1E7D8] hover:text-[#251F1A]"
                  href="/login"
                  onClick={() => setOpen(false)}
                >
                  Masuk
                </Link>
                <Link
                  className="rounded-2xl bg-[#251F1A] px-3 py-2.5 text-center text-sm font-semibold text-[#FFFDF8]"
                  href="/register"
                  onClick={() => setOpen(false)}
                >
                  Daftar Toko
                </Link>
              </div>
            ) : activeTenant ? (
              <div className="grid gap-1">
                <Link
                  className="rounded-2xl bg-[#251F1A] px-3 py-2.5 text-center text-sm font-semibold text-[#FFFDF8]"
                  href="/dashboard"
                  onClick={() => setOpen(false)}
                >
                  Dashboard
                </Link>
                <Link
                  className="rounded-2xl px-3 py-2.5 text-sm font-semibold text-[#6F6256] hover:bg-[#F1E7D8] hover:text-[#251F1A]"
                  href={`/s/${activeTenant.store.slug}`}
                  onClick={() => setOpen(false)}
                >
                  Toko Saya
                </Link>
                <button
                  className="rounded-2xl px-3 py-2.5 text-left text-sm font-semibold text-[#B96E45] hover:bg-[#FFF5DE]"
                  onClick={() => void handleLogout()}
                  type="button"
                >
                  Keluar
                </button>
              </div>
            ) : (
              <div className="grid gap-1">
                <Link
                  className="rounded-2xl bg-[#251F1A] px-3 py-2.5 text-center text-sm font-semibold text-[#FFFDF8]"
                  href={user?.platformRole === "super_admin" ? "/admin" : "/onboarding/create-store"}
                  onClick={() => setOpen(false)}
                >
                  {user?.platformRole === "super_admin" ? "Admin" : "Buat toko"}
                </Link>
                <button
                  className="rounded-2xl px-3 py-2.5 text-left text-sm font-semibold text-[#B96E45] hover:bg-[#FFF5DE]"
                  onClick={() => void handleLogout()}
                  type="button"
                >
                  Keluar
                </button>
              </div>
            )}
          </div>
        </div>
      ) : null}
    </div>
  );
}
