import { Badge } from "@/components/ui/badge";

type StorefrontHeaderProps = {
  storeSlug: string;
};

export function StorefrontHeader({ storeSlug }: StorefrontHeaderProps) {
  return (
    <header className="border-b border-neutral-200 bg-white">
      <div className="mx-auto flex max-w-5xl items-center justify-between gap-4 px-4 py-4 sm:px-6 lg:px-8">
        <div>
          <p className="text-sm font-semibold text-neutral-950">{storeSlug}</p>
          <p className="text-xs text-neutral-500">Storefront shell</p>
        </div>
        <Badge tone="neutral">Publik</Badge>
      </div>
    </header>
  );
}
