import Link from "next/link";
import { formatRupiah } from "@/lib/format/money";
import type { PublicProductListItem } from "@/features/storefront/types";
import { StockBadge } from "@/features/storefront/components/stock-badge";

type ProductCardProps = {
  storeSlug: string;
  product: PublicProductListItem;
  categoryName?: string;
};

export function ProductCard({ storeSlug, product, categoryName }: ProductCardProps) {
  return (
    <Link
      href={`/s/${storeSlug}/products/${product.slug}`}
      className="group overflow-hidden rounded-3xl border border-neutral-200 bg-white shadow-soft transition hover:-translate-y-0.5 hover:shadow-md"
    >
      <div className="aspect-square overflow-hidden bg-neutral-100">
        {product.primaryImageUrl ? (
          <img
            alt={product.name}
            className="h-full w-full object-cover transition duration-200 group-hover:scale-[1.02]"
            src={product.primaryImageUrl}
          />
        ) : (
          <div className="flex h-full items-center justify-center text-sm text-neutral-400">Belum ada foto</div>
        )}
      </div>

      <div className="space-y-3 p-4">
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
