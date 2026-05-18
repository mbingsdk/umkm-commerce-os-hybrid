import type { ReactNode } from "react";
import { StorefrontHeader } from "@/components/layout/storefront-header";

type StorefrontLayoutProps = {
  children: ReactNode;
  params: Promise<{ storeSlug: string }>;
};

export default async function StorefrontLayout({ children, params }: StorefrontLayoutProps) {
  const { storeSlug } = await params;

  return (
    <div className="min-h-screen bg-neutral-50">
      <StorefrontHeader storeSlug={storeSlug} />
      {children}
    </div>
  );
}
