import type { PublicProductDetail } from "@/features/storefront/types";

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
        <img
          alt={primaryImage.altText ?? product.name}
          className="aspect-square h-full w-full object-cover"
          src={primaryImage.url}
        />
      </div>

      {secondaryImages.length > 0 ? (
        <div className="grid grid-cols-4 gap-3">
          {secondaryImages.map((image, index) => (
            <div key={`${image.url}-${index}`} className="overflow-hidden rounded-2xl bg-neutral-100">
              <img
                alt={image.altText ?? product.name}
                className="aspect-square h-full w-full object-cover"
                src={image.url}
              />
            </div>
          ))}
        </div>
      ) : null}
    </div>
  );
}
