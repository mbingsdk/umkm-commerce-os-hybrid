import Link from "next/link";
import { Search } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { SafeImage } from "@/features/storefront/components/safe-image";
import { StorefrontCartLink } from "@/features/storefront/components/storefront-cart-link";

type StorefrontHeaderProps = {
  storeSlug: string;
  storeName?: string;
  logoUrl?: string;
  city?: string;
};

export function StorefrontHeader({ storeSlug, storeName, logoUrl, city }: StorefrontHeaderProps) {
  const navItems = [
    { label: "Beranda", href: `/s/${storeSlug}` },
    { label: "Produk", href: `/s/${storeSlug}/products` },
    { label: "Tentang", href: `/s/${storeSlug}/about` },
    { label: "Kontak", href: `/s/${storeSlug}/contact` },
    { label: "Lacak", href: `/s/${storeSlug}/track-order` },
    { label: "Jelajah", href: "/" }
  ];

  return (
    <header className="sticky top-0 z-40 border-b border-[#E3D2BC] bg-[#FFFDF8]/92 backdrop-blur">
      <div className="mx-auto max-w-[1500px] px-3 py-1.5 sm:px-6 lg:px-8">
        <div className="grid grid-cols-[minmax(112px,1fr)_auto] items-center gap-2 md:grid-cols-[minmax(220px,1fr)_auto_minmax(280px,1fr)] md:gap-3">
          <Link className="flex min-w-0 items-center gap-2 md:gap-3" href={`/s/${storeSlug}`}>
            {logoUrl ? (
              <SafeImage
                alt=""
                className="h-8 w-8 rounded-2xl object-cover md:h-10 md:w-10"
                fallbackClassName="h-8 w-8 rounded-2xl md:h-10 md:w-10"
                src={logoUrl}
              />
            ) : (
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-2xl bg-[#6f4e37] text-sm font-bold text-[#FFFDF8] md:h-10 md:w-10">
                {(storeName ?? storeSlug).slice(0, 1).toUpperCase()}
              </div>
            )}
            <div className="min-w-0">
              <p className="truncate text-xs font-bold leading-4 text-[#251F1A] sm:text-sm">{storeName ?? storeSlug}</p>
              <p className="hidden truncate text-xs text-[#6F6256] sm:block">{city ?? "Storefront publik"}</p>
            </div>
          </Link>

          <nav
            aria-label="Navigasi toko"
            className="hidden items-center gap-1 rounded-full border border-[#E3D2BC] bg-white/80 p-0.5 shadow-[0_8px_24px_rgba(89,63,38,0.06)] md:flex"
          >
            {navItems.map((item) => (
              <Link
                key={item.href}
                className="rounded-full px-3 py-1.5 text-sm font-semibold text-[#6F6256] transition hover:bg-[#F1E7D8] hover:text-[#251F1A]"
                href={item.href}
              >
                {item.label}
              </Link>
            ))}
          </nav>

          <div className="flex min-w-0 shrink-0 items-center justify-end gap-1.5 sm:gap-2">
            <form
              action={`/s/${storeSlug}/products`}
              className="flex h-9 w-[min(40vw,136px)] min-w-0 items-center rounded-2xl border border-[#E3D2BC] bg-white pl-2.5 shadow-sm sm:w-52 md:h-10 md:w-full md:max-w-xs md:pl-3"
              method="get"
            >
              <input
                aria-label="Cari produk toko"
                className="min-w-0 flex-1 bg-transparent text-sm text-[#251F1A] outline-none placeholder:text-[#9B8D7B]"
                name="q"
                placeholder="Cari..."
              />
              <button
                aria-label="Cari produk"
                className="flex h-full w-8 shrink-0 items-center justify-center rounded-r-2xl text-[#B96E45] transition hover:bg-[#F1E7D8] hover:text-[#7C3F25] md:w-10"
                type="submit"
              >
                <Search aria-hidden="true" className="h-4 w-4" />
              </button>
            </form>
            <StorefrontCartLink storeSlug={storeSlug} />
            <Badge className="hidden lg:inline-flex" tone="neutral">
              Storefront aktif
            </Badge>
          </div>
        </div>
      </div>
    </header>
  );
}
