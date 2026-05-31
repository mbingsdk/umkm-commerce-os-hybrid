import Link from "next/link";
import { SafeImage } from "@/features/storefront/components/safe-image";
import type { DiscoveryProduct, DiscoveryStore } from "@/features/discovery/types";
import { formatRupiah } from "@/lib/format/money";

export function DiscoveryStoreCard({ store }: { store: DiscoveryStore }) {
  return (
    <article className="overflow-hidden rounded-[24px] border border-[#E3D2BC] bg-[#FFFDF8] shadow-[0_10px_30px_rgba(89,63,38,0.07)] transition hover:-translate-y-0.5 hover:shadow-[0_16px_42px_rgba(89,63,38,0.12)]">
      <Link href={store.storeUrl} className="block">
        <div className="relative h-28 bg-[#F1E7D8] sm:h-32">
          <SafeImage
            alt=""
            className="h-full w-full object-cover"
            fallbackClassName="h-full w-full"
            src={store.bannerUrl || store.logoUrl}
          />
          <div className="absolute left-3 top-3 flex h-12 w-12 items-center justify-center overflow-hidden rounded-2xl border border-white/80 bg-[#FFFDF8] text-base font-bold text-[#B96E45] shadow-soft">
            {store.logoUrl ? (
              <SafeImage alt="" className="h-full w-full object-cover" fallbackLabel={store.name.slice(0, 1)} src={store.logoUrl} />
            ) : (
              store.name.slice(0, 1).toUpperCase()
            )}
          </div>
        </div>

        <div className="space-y-2.5 p-3.5">
          <div>
            <div className="flex items-start justify-between gap-2">
              <h2 className="line-clamp-1 text-sm font-bold text-[#251F1A] sm:text-base">{store.name}</h2>
              <span className="shrink-0 rounded-full bg-[#F4F7ED] px-2 py-0.5 text-[11px] font-semibold text-[#6F7D55]">
                Storefront aktif
              </span>
            </div>
            <p className="mt-1 text-xs font-medium text-[#6F6256]">
              {[store.city, store.province].filter(Boolean).join(", ") || "Toko lokal"}
            </p>
          </div>
          <p className="line-clamp-2 text-xs leading-5 text-[#6F6256] sm:text-sm sm:leading-6">
            {store.description || "Storefront aktif untuk melihat katalog dan kontak toko."}
          </p>
          <span className="inline-flex min-h-8 items-center rounded-full bg-[#251F1A] px-3 text-xs font-semibold text-[#FFFDF8] sm:text-sm">
            Kunjungi toko
          </span>
        </div>
      </Link>
    </article>
  );
}

export function DiscoveryProductCard({ product }: { product: DiscoveryProduct }) {
  return (
    <article className="group overflow-hidden rounded-[24px] border border-[#E3D2BC] bg-[#FFFDF8] shadow-[0_10px_30px_rgba(89,63,38,0.07)] transition hover:-translate-y-0.5 hover:shadow-[0_16px_42px_rgba(89,63,38,0.12)]">
      <Link href={product.productUrl} className="block">
        <div className="relative aspect-square overflow-hidden bg-[#F1E7D8]">
          <SafeImage
            alt={product.name}
            className="h-full w-full object-cover transition duration-200 group-hover:scale-[1.02]"
            fallbackClassName="h-full w-full"
            fallbackLabel="Foto produk"
            src={product.primaryImageUrl}
          />
          <span className="absolute left-2 top-2 rounded-full bg-[#FFFDF8]/90 px-2 py-1 text-[11px] font-semibold text-[#6F7D55] shadow-sm">
            Cek stok di toko
          </span>
        </div>
      </Link>

      <div className="space-y-2.5 p-3.5">
        <Link href={product.productUrl} className="block space-y-1.5">
          {product.category ? <p className="text-[11px] font-semibold uppercase tracking-wide text-[#8b6f4e]">{product.category.name}</p> : null}
          <h2 className="line-clamp-2 min-h-10 text-sm font-bold leading-5 text-[#251F1A]">{product.name}</h2>
          <p className="text-base font-bold text-[#B96E45]">{formatRupiah(product.price)}</p>
        </Link>

        <Link href={product.storeUrl} className="block rounded-2xl border border-[#E3D2BC] bg-white/75 p-2.5 text-xs transition hover:bg-white">
          <p className="line-clamp-1 font-semibold text-[#251F1A]">{product.store.name}</p>
          <p className="mt-0.5 text-[#6F6256]">
            {[product.store.city, product.store.province].filter(Boolean).join(", ") || "Toko lokal"}
          </p>
        </Link>
        <Link
          href={product.productUrl}
          className="inline-flex min-h-8 w-full items-center justify-center rounded-full bg-[#F4F7ED] px-3 text-xs font-semibold text-[#6F7D55] transition hover:bg-[#E9F0DF] sm:text-sm"
        >
          Lihat produk
        </Link>
      </div>
    </article>
  );
}
