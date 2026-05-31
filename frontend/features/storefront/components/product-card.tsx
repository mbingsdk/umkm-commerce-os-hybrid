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
        "group overflow-hidden rounded-[20px] border border-[#E3D2BC] bg-[#FFFDF8] shadow-[0_6px_20px_rgba(89,63,38,0.055)] transition hover:-translate-y-0.5 hover:shadow-[0_12px_32px_rgba(89,63,38,0.09)]",
        isUnavailable && "bg-[#F1E7D8]"
      )}
    >
      <Link className="block" href={`/s/${storeSlug}/products/${product.slug}`}>
        <div className="relative aspect-[5/4] overflow-hidden bg-[#F1E7D8] sm:aspect-[4/3]">
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
            <div className="absolute inset-0 flex items-end bg-white/45 p-1.5">
              <span className="rounded-full bg-[#251F1A]/85 px-2.5 py-1 text-[11px] font-semibold text-[#FFFDF8]">
                Stok habis
              </span>
            </div>
          ) : null}
        </div>
      </Link>

      <div className="space-y-1.5 p-2 sm:p-2.5">
        <Link className="block space-y-1.5" href={`/s/${storeSlug}/products/${product.slug}`}>
          <div className="space-y-0.5">
            {categoryName ? <p className="truncate text-[10px] font-semibold text-[#B96E45] sm:text-[11px]">{categoryName}</p> : null}
            <h2 className="line-clamp-2 text-[13px] font-bold leading-5 text-[#251F1A] sm:text-sm">{product.name}</h2>
          </div>

          <div>
            <p className="text-sm font-bold text-[#B96E45] sm:text-base">{formatRupiah(product.price)}</p>
            {product.compareAtPrice ? (
              <p className="text-[11px] text-neutral-400 line-through">{formatRupiah(product.compareAtPrice)}</p>
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
          size="sm"
          className="w-full bg-[#251F1A] text-[#FFFDF8] hover:bg-[#16110E]"
        />
      </div>
    </article>
  );
}
