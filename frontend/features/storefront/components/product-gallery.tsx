import type { PublicProductDetail } from "@/features/storefront/types";
import { SafeImage } from "@/features/storefront/components/safe-image";

export function ProductGallery({ product }: { product: PublicProductDetail }) {
  if (product.images.length === 0) {
    return (
      <div className="flex aspect-[5/4] items-center justify-center rounded-[24px] bg-[#F1E7D8] text-sm text-[#9B8D7B] sm:aspect-square">
        Belum ada foto produk
      </div>
    );
  }

  const [primaryImage, ...secondaryImages] = product.images;

  return (
    <div className="space-y-2.5">
      <div className="overflow-hidden rounded-[24px] bg-[#F1E7D8]">
        <SafeImage
          alt={primaryImage.altText ?? product.name}
          className="aspect-[5/4] h-full w-full object-cover sm:aspect-square"
          fallbackClassName="aspect-[5/4] h-full w-full rounded-[24px] sm:aspect-square"
          fallbackLabel="Foto produk belum tersedia"
          src={primaryImage.url}
        />
      </div>

      {secondaryImages.length > 0 ? (
        <div className="grid grid-cols-4 gap-2">
          {secondaryImages.map((image, index) => (
            <div key={`${image.url}-${index}`} className="overflow-hidden rounded-2xl bg-[#F1E7D8]">
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
