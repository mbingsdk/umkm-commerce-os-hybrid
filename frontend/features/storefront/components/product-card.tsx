import Link from "next/link";
import { formatRupiah } from "@/lib/format/money";
import { cn } from "@/lib/utils/cn";
import { AddToCartButton } from "@/features/storefront/components/add-to-cart-button";
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
    <article
      className={cn(
        "group overflow-hidden rounded-[28px] border border-[#E3D2BC] bg-[#FFFDF8] shadow-[0_14px_40px_rgba(89,63,38,0.07)] transition hover:-translate-y-0.5 hover:shadow-[0_18px_50px_rgba(89,63,38,0.11)]",
        isUnavailable && "bg-[#F1E7D8]"
      )}
    >
      <Link className="block" href={`/s/${storeSlug}/products/${product.slug}`}>
        <div className="relative aspect-square overflow-hidden bg-[#F1E7D8]">
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
              <span className="rounded-full bg-[#251F1A]/85 px-3 py-1 text-xs font-semibold text-[#FFFDF8]">
                Stok habis
              </span>
            </div>
          ) : null}
        </div>
      </Link>

      <div className="space-y-3 p-3 sm:p-4">
        <Link className="block space-y-3" href={`/s/${storeSlug}/products/${product.slug}`}>
          <div className="space-y-1">
            {categoryName ? <p className="text-xs font-semibold text-[#B96E45]">{categoryName}</p> : null}
            <h2 className="line-clamp-2 text-sm font-bold text-[#251F1A]">{product.name}</h2>
          </div>

          <div>
            <p className="text-base font-bold text-[#B96E45]">{formatRupiah(product.price)}</p>
            {product.compareAtPrice ? (
              <p className="text-xs text-neutral-400 line-through">{formatRupiah(product.compareAtPrice)}</p>
            ) : null}
          </div>

          <StockBadge status={product.stockStatus} />
        </Link>

        <AddToCartButton
          disabled={isUnavailable}
          item={{
            productId: product.id,
            storeSlug,
            name: product.name,
            slug: product.slug,
            imageUrl: product.primaryImageUrl,
            displayPrice: product.price,
            quantity: 1
          }}
          label="Tambah"
          size="md"
          className="w-full"
        />
      </div>
    </article>
  );
}
