import Link from "next/link";
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
    { label: "Lacak order", href: `/s/${storeSlug}/track-order` }
  ];

  return (
    <header className="sticky top-0 z-40 border-b border-[#eadfce] bg-[#fffaf2]/92 backdrop-blur">
      <div className="mx-auto max-w-6xl px-4 py-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between gap-4">
          <Link className="flex min-w-0 items-center gap-3" href={`/s/${storeSlug}`}>
            {logoUrl ? (
              <SafeImage
                alt=""
                className="h-10 w-10 rounded-2xl object-cover"
                fallbackClassName="h-10 w-10 rounded-2xl"
                src={logoUrl}
              />
            ) : (
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-[#6f4e37] text-sm font-bold text-[#fffaf2]">
                {(storeName ?? storeSlug).slice(0, 1).toUpperCase()}
              </div>
            )}
            <div className="min-w-0">
              <p className="truncate text-sm font-bold text-[#241c16]">{storeName ?? storeSlug}</p>
              <p className="truncate text-xs text-[#7a6a58]">{city ?? "Storefront publik"}</p>
            </div>
          </Link>

          <div className="flex shrink-0 items-center gap-2">
            <StorefrontCartLink storeSlug={storeSlug} />
            <Badge className="hidden sm:inline-flex" tone="neutral">
              Publik
            </Badge>
          </div>
        </div>

        <nav
          aria-label="Navigasi toko"
          className="mt-3 flex gap-2 overflow-x-auto pb-1 text-sm text-neutral-700"
        >
          {navItems.map((item) => (
            <Link
              key={item.href}
              className="shrink-0 rounded-full border border-[#eadfce] bg-white/75 px-3 py-2 font-semibold text-[#5f5042] transition hover:border-[#9a6a43] hover:text-[#7a4f2f]"
              href={item.href}
            >
              {item.label}
            </Link>
          ))}
        </nav>
      </div>
    </header>
  );
}
