import { Badge } from "@/components/ui/badge";
import { SafeImage } from "@/features/storefront/components/safe-image";
import { StorefrontCartLink } from "@/features/storefront/components/storefront-cart-link";

type StorefrontHeaderProps = {
  storeSlug: string;
  storeName?: string;
  logoUrl?: string;
  city?: string;
};

export function StorefrontHeader({ storeSlug, storeName, logoUrl, city }: StorefrontHeaderProps) {
  return (
    <header className="border-b border-neutral-200 bg-white">
      <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-4 sm:px-6 lg:px-8">
        <div className="flex items-center gap-3">
          {logoUrl ? (
            <SafeImage
              alt=""
              className="h-10 w-10 rounded-2xl object-cover"
              fallbackClassName="h-10 w-10 rounded-2xl"
              src={logoUrl}
            />
          ) : (
            <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-primary-100 text-sm font-bold text-primary-800">
              {(storeName ?? storeSlug).slice(0, 1).toUpperCase()}
            </div>
          )}
          <div>
            <p className="text-sm font-semibold text-neutral-950">{storeName ?? storeSlug}</p>
            <p className="text-xs text-neutral-500">{city ?? "Storefront publik"}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <StorefrontCartLink storeSlug={storeSlug} />
          <Badge className="hidden sm:inline-flex" tone="neutral">
            Publik
          </Badge>
        </div>
      </div>
    </header>
  );
}
