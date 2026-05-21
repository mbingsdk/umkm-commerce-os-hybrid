import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { SafeImage } from "@/features/storefront/components/safe-image";
import type { DiscoveryProduct, DiscoveryStore } from "@/features/discovery/types";
import { formatRupiah } from "@/lib/format/money";

export function DiscoveryStoreCard({ store }: { store: DiscoveryStore }) {
  return (
    <article className="overflow-hidden rounded-3xl border border-neutral-200 bg-white shadow-soft transition hover:-translate-y-0.5 hover:shadow-md">
      <Link href={store.storeUrl} className="block">
        <div className="relative h-32 bg-neutral-100">
          <SafeImage
            alt=""
            className="h-full w-full object-cover"
            fallbackClassName="h-full w-full"
            src={store.bannerUrl || store.logoUrl}
          />
          <div className="absolute left-4 top-4 flex h-14 w-14 items-center justify-center overflow-hidden rounded-2xl border border-white/70 bg-white text-lg font-bold text-primary-700 shadow-soft">
            {store.logoUrl ? (
              <SafeImage alt="" className="h-full w-full object-cover" fallbackLabel={store.name.slice(0, 1)} src={store.logoUrl} />
            ) : (
              store.name.slice(0, 1).toUpperCase()
            )}
          </div>
        </div>

        <div className="space-y-3 p-4">
          <div>
            <h2 className="line-clamp-1 text-base font-semibold text-neutral-950">{store.name}</h2>
            <p className="mt-1 text-sm text-neutral-500">
              {[store.city, store.province].filter(Boolean).join(", ") || "Toko lokal"}
            </p>
          </div>
          <p className="line-clamp-2 text-sm leading-6 text-neutral-600">
            {store.description || "Toko ini belum menambahkan deskripsi singkat."}
          </p>
          <span className="inline-flex text-sm font-semibold text-primary-700">Kunjungi toko →</span>
        </div>
      </Link>
    </article>
  );
}

export function DiscoveryProductCard({ product }: { product: DiscoveryProduct }) {
  return (
    <article className="group overflow-hidden rounded-3xl border border-neutral-200 bg-white shadow-soft transition hover:-translate-y-0.5 hover:shadow-md">
      <Link href={product.productUrl} className="block">
        <div className="aspect-square overflow-hidden bg-neutral-100">
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
          <h2 className="line-clamp-2 text-sm font-semibold text-neutral-950">{product.name}</h2>
          <p className="text-base font-bold text-primary-700">{formatRupiah(product.price)}</p>
        </Link>

        <Link href={product.storeUrl} className="block rounded-2xl bg-neutral-50 p-3 text-sm">
          <p className="font-semibold text-neutral-950">{product.store.name}</p>
          <p className="mt-1 text-xs text-neutral-500">
            {[product.store.city, product.store.province].filter(Boolean).join(", ") || "Toko lokal"}
          </p>
        </Link>
      </div>
    </article>
  );
}
