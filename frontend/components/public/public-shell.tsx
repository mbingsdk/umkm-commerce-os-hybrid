import type { ReactNode } from "react";
import Link from "next/link";

const navItems = [
  { label: "Jelajah", href: "/explore" },
  { label: "Toko", href: "/stores" },
  { label: "Produk", href: "/products" },
  { label: "Harga", href: "/pricing" }
];

export function PublicMarketingShell({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen bg-[#f7f1e8] text-[#241c16]">
      <header className="sticky top-0 z-40 border-b border-[#eadfce] bg-[#fffaf2]/90 backdrop-blur">
        <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-3 sm:px-6 lg:px-8">
          <Link href="/" className="flex min-w-0 items-center gap-3">
            <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-[#6f4e37] text-sm font-bold text-[#fffaf2]">
              UM
            </span>
            <span className="min-w-0">
              <span className="block truncate text-sm font-bold tracking-tight text-[#241c16]">
                UMKM Commerce OS
              </span>
              <span className="block truncate text-xs text-[#7a6a58]">Storefront dan discovery UMKM</span>
            </span>
          </Link>

          <nav className="hidden items-center gap-1 md:flex" aria-label="Navigasi publik">
            {navItems.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className="rounded-full px-3 py-2 text-sm font-semibold text-[#5f5042] transition hover:bg-[#f0e4d2] hover:text-[#241c16]"
              >
                {item.label}
              </Link>
            ))}
          </nav>

          <div className="flex shrink-0 items-center gap-2">
            <Link
              href="/login"
              className="hidden rounded-full px-3 py-2 text-sm font-semibold text-[#5f5042] transition hover:bg-[#f0e4d2] sm:inline-flex"
            >
              Masuk
            </Link>
            <Link
              href="/register"
              className="inline-flex min-h-10 items-center justify-center rounded-full bg-[#2f2923] px-4 text-sm font-semibold text-[#fffaf2] shadow-sm transition hover:bg-[#1f1a16]"
            >
              Mulai toko
            </Link>
          </div>
        </div>

        <nav className="mx-auto flex max-w-6xl gap-2 overflow-x-auto px-4 pb-3 text-sm sm:px-6 md:hidden" aria-label="Navigasi publik mobile">
          {navItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="shrink-0 rounded-full border border-[#eadfce] bg-white/70 px-3 py-2 font-semibold text-[#5f5042]"
            >
              {item.label}
            </Link>
          ))}
          <Link
            href="/login"
            className="shrink-0 rounded-full border border-[#eadfce] bg-white/70 px-3 py-2 font-semibold text-[#5f5042]"
          >
            Masuk
          </Link>
        </nav>
      </header>

      {children}

      <footer className="border-t border-[#eadfce] bg-[#fffaf2]">
        <div className="mx-auto grid max-w-6xl gap-6 px-4 py-8 text-sm text-[#6d5e4e] sm:px-6 md:grid-cols-[minmax(0,1fr)_auto] lg:px-8">
          <div>
            <p className="font-semibold text-[#241c16]">UMKM Commerce OS Hybrid</p>
            <p className="mt-2 max-w-2xl leading-6">
              Platform ringan untuk storefront tenant, discovery toko dan produk, checkout manual, serta operasional UMKM.
            </p>
          </div>
          <div className="flex flex-wrap gap-3">
            <Link className="font-semibold hover:text-[#241c16]" href="/stores">
              Toko
            </Link>
            <Link className="font-semibold hover:text-[#241c16]" href="/products">
              Produk
            </Link>
            <Link className="font-semibold hover:text-[#241c16]" href="/pricing">
              Harga
            </Link>
          </div>
        </div>
      </footer>
    </div>
  );
}
