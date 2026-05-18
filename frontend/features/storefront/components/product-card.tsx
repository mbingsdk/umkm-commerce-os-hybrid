import Link from "next/link";
import { formatRupiah } from "@/lib/format/money";
import { cn } from "@/lib/utils/cn";
import { SafeImage } from "@/features/storefront/components/safe-image";
import type { PublicProductListItem } from "@/features/storefront/types";
import { StockBadge } from "@/features/storefront/components/stock-badge";

type ProductCardProps = {
  storeSlug: string;
  product: PublicProductListItem;
  categoryName?: string;
};

export function ProductCard({ storeSlug, product, categoryName }: ProductCardProps) {
  const isUnavailable = product.stockStatus === "out_of_stock";

  return (
    <Link
      href={`/s/${storeSlug}/products/${product.slug}`}
      className={cn(
        "group overflow-hidden rounded-3xl border border-neutral-200 bg-white shadow-soft transition hover:-translate-y-0.5 hover:shadow-md",
        isUnavailable && "bg-neutral-50"
      )}
    >
      <div className="relative aspect-square overflow-hidden bg-neutral-100">
        <SafeImage
          alt={product.name}
          className={cn(
            "h-full w-full object-cover transition duration-200 group-hover:scale-[1.02]",
            isUnavailable && "grayscale"
          )}
          fallbackClassName="h-full w-full"
          fallbackLabel="Belum ada foto"
          src={product.primaryImageUrl}
        />
        {isUnavailable ? (
          <div className="absolute inset-0 flex items-end bg-white/45 p-3">
            <span className="rounded-full bg-neutral-950/80 px-3 py-1 text-xs font-semibold text-white">
              Stok habis
            </span>
          </div>
        ) : null}
      </div>

      <div className="space-y-3 p-3 sm:p-4">
        <div className="space-y-1">
          {categoryName ? <p className="text-xs font-medium text-primary-700">{categoryName}</p> : null}
          <h2 className="line-clamp-2 text-sm font-semibold text-neutral-950">{product.name}</h2>
        </div>

        <div>
          <p className="text-base font-bold text-primary-700">{formatRupiah(product.price)}</p>
          {product.compareAtPrice ? (
            <p className="text-xs text-neutral-400 line-through">{formatRupiah(product.compareAtPrice)}</p>
          ) : null}
        </div>

        <StockBadge status={product.stockStatus} />
      </div>
    </Link>
  );
}
