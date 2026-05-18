import type { PublicProductDetail } from "@/features/storefront/types";
import { SafeImage } from "@/features/storefront/components/safe-image";

export function ProductGallery({ product }: { product: PublicProductDetail }) {
  if (product.images.length === 0) {
    return (
      <div className="flex aspect-square items-center justify-center rounded-3xl bg-neutral-100 text-sm text-neutral-400">
        Belum ada foto produk
      </div>
    );
  }

  const [primaryImage, ...secondaryImages] = product.images;

  return (
    <div className="space-y-4">
      <div className="overflow-hidden rounded-3xl bg-neutral-100">
        <SafeImage
          alt={primaryImage.altText ?? product.name}
          className="aspect-square h-full w-full object-cover"
          fallbackClassName="aspect-square h-full w-full rounded-3xl"
          fallbackLabel="Foto produk belum tersedia"
          src={primaryImage.url}
        />
      </div>

      {secondaryImages.length > 0 ? (
        <div className="grid grid-cols-4 gap-3">
          {secondaryImages.map((image, index) => (
            <div key={`${image.url}-${index}`} className="overflow-hidden rounded-2xl bg-neutral-100">
              <SafeImage
                alt={image.altText ?? product.name}
                className="aspect-square h-full w-full object-cover"
                fallbackClassName="aspect-square h-full w-full rounded-2xl text-xs"
                fallbackLabel="Foto"
                src={image.url}
              />
            </div>
          ))}
        </div>
      ) : null}
    </div>
  );
}
