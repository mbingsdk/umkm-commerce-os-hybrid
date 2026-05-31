import type { ReactNode } from "react";
import Link from "next/link";
import { PublicBottomDock } from "@/components/public/public-bottom-dock";
import { PublicAuthNav } from "@/components/public/public-auth-nav";
import { PublicMobileMenu } from "@/components/public/public-mobile-menu";
import { publicTheme } from "@/components/public/public-ui";
import { cn } from "@/lib/utils/cn";

const navItems = [
  { label: "Toko", href: "/stores" },
  { label: "Produk", href: "/products" },
  { label: "Harga", href: "/pricing" }
];

export function PublicMarketingShell({ children }: { children: ReactNode }) {
  return (
    <div className={cn("min-h-screen pb-20 md:pb-0", publicTheme.bg, publicTheme.text)}>
      <header className="sticky top-0 z-40 border-b border-[#E3D2BC] bg-[#FFFDF8]/95 backdrop-blur">
        <div className="mx-auto flex max-w-6xl items-center gap-3 px-4 py-2.5 sm:px-6 lg:px-8">
          <Link href="/" className="flex min-w-0 shrink-0 items-center gap-2" aria-label="UMKM Commerce OS">
            <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-[#251F1A] text-sm font-bold text-[#FFFDF8] shadow-[0_8px_18px_rgba(47,41,35,0.14)]">
              U
            </span>
            <span className="min-w-0">
              <span className="block truncate text-sm font-bold tracking-tight text-[#251F1A]">UMKM Commerce OS</span>
              <span className="hidden truncate text-xs text-[#6F6256] sm:block">Discovery toko lokal</span>
            </span>
          </Link>

          <nav className="ml-auto hidden items-center gap-1 md:flex" aria-label="Navigasi publik">
            {navItems.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className="px-2.5 py-2 text-sm font-semibold text-[#6F6256] transition hover:text-[#251F1A]"
              >
                {item.label}
              </Link>
            ))}
          </nav>

          <div className="ml-auto flex shrink-0 items-center gap-2 md:ml-0">
            <PublicMobileMenu />
            <PublicAuthNav />
          </div>
        </div>
      </header>

      {children}
      <PublicBottomDock />

      <footer className="border-t border-[#E3D2BC] bg-[#FFFDF8]">
        <div className="mx-auto grid max-w-6xl gap-6 px-4 py-8 text-sm text-[#6F6256] sm:px-6 md:grid-cols-[minmax(0,1fr)_auto] lg:px-8">
          <div>
            <p className="font-semibold text-[#251F1A]">UMKM Commerce OS</p>
            <p className="mt-2 max-w-2xl leading-6">
              Tempat menemukan toko dan produk UMKM lokal, lalu belanja langsung di storefront resmi masing-masing toko.
            </p>
          </div>
          <div className="flex flex-wrap gap-3">
            <Link className="font-semibold hover:text-[#251F1A]" href="/stores">
              Toko
            </Link>
            <Link className="font-semibold hover:text-[#251F1A]" href="/products">
              Produk
            </Link>
            <Link className="font-semibold hover:text-[#251F1A]" href="/pricing">
              Harga
            </Link>
          </div>
        </div>
      </footer>
    </div>
  );
}
