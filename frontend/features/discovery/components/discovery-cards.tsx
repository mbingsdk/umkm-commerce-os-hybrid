import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { SafeImage } from "@/features/storefront/components/safe-image";
import type { DiscoveryProduct, DiscoveryStore } from "@/features/discovery/types";
import { formatRupiah } from "@/lib/format/money";

export function DiscoveryStoreCard({ store }: { store: DiscoveryStore }) {
  return (
    <article className="overflow-hidden rounded-[28px] border border-[#eadfce] bg-[#fffaf2] shadow-[0_14px_40px_rgba(89,63,38,0.07)] transition hover:-translate-y-0.5 hover:shadow-[0_18px_50px_rgba(89,63,38,0.11)]">
      <Link href={store.storeUrl} className="block">
        <div className="relative h-36 bg-[#efe2cf]">
          <SafeImage
            alt=""
            className="h-full w-full object-cover"
            fallbackClassName="h-full w-full"
            src={store.bannerUrl || store.logoUrl}
          />
          <div className="absolute left-4 top-4 flex h-14 w-14 items-center justify-center overflow-hidden rounded-2xl border border-white/80 bg-[#fffaf2] text-lg font-bold text-[#7a4f2f] shadow-soft">
            {store.logoUrl ? (
              <SafeImage alt="" className="h-full w-full object-cover" fallbackLabel={store.name.slice(0, 1)} src={store.logoUrl} />
            ) : (
              store.name.slice(0, 1).toUpperCase()
            )}
          </div>
        </div>

        <div className="space-y-3 p-4">
          <div>
            <h2 className="line-clamp-1 text-base font-bold text-[#241c16]">{store.name}</h2>
            <p className="mt-1 text-sm text-[#7a6a58]">
              {[store.city, store.province].filter(Boolean).join(", ") || "Toko lokal"}
            </p>
          </div>
          <p className="line-clamp-2 text-sm leading-6 text-[#6d5e4e]">
            {store.description || "Toko ini belum menambahkan deskripsi singkat."}
          </p>
          <span className="inline-flex text-sm font-semibold text-[#7a4f2f]">Kunjungi toko →</span>
        </div>
      </Link>
    </article>
  );
}

export function DiscoveryProductCard({ product }: { product: DiscoveryProduct }) {
  return (
    <article className="group overflow-hidden rounded-[28px] border border-[#eadfce] bg-[#fffaf2] shadow-[0_14px_40px_rgba(89,63,38,0.07)] transition hover:-translate-y-0.5 hover:shadow-[0_18px_50px_rgba(89,63,38,0.11)]">
      <Link href={product.productUrl} className="block">
        <div className="aspect-square overflow-hidden bg-[#efe2cf]">
          <SafeImage
            alt={product.name}
            className="h-full w-full object-cover transition duration-200 group-hover:scale-[1.02]"
            fallbackClassName="h-full w-full"
            fallbackLabel="Belum ada foto"
            src={product.primaryImageUrl}
          />
        </div>
      </Link>

      <div className="space-y-3 p-4">
        <Link href={product.productUrl} className="block space-y-2">
          {product.category ? <Badge tone="primary">{product.category.name}</Badge> : null}
          <h2 className="line-clamp-2 text-sm font-bold text-[#241c16]">{product.name}</h2>
          <p className="text-base font-bold text-[#7a4f2f]">{formatRupiah(product.price)}</p>
        </Link>

        <Link href={product.storeUrl} className="block rounded-2xl border border-[#eadfce] bg-white/70 p-3 text-sm">
          <p className="font-semibold text-[#241c16]">{product.store.name}</p>
          <p className="mt-1 text-xs text-[#7a6a58]">
            {[product.store.city, product.store.province].filter(Boolean).join(", ") || "Toko lokal"}
          </p>
        </Link>
      </div>
    </article>
  );
}
