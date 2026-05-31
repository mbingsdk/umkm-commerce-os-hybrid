import type { ReactNode } from "react";
import { notFound } from "next/navigation";
import { StorefrontHeader } from "@/components/layout/storefront-header";
import { getPublicStoreBySlug, isPublicNotFoundError } from "@/features/storefront/api/storefront.api";

type StorefrontLayoutProps = {
  children: ReactNode;
  params: Promise<{ storeSlug: string }>;
};

export default async function StorefrontLayout({ children, params }: StorefrontLayoutProps) {
  const { storeSlug } = await params;

  let store;
  try {
    store = await getPublicStoreBySlug(storeSlug);
  } catch (error) {
    if (isPublicNotFoundError(error)) {
      notFound();
    }
    throw error;
  }

  return (
    <div className="min-h-screen bg-[#f7f1e8]">
      <StorefrontHeader
        city={store.city}
        logoUrl={store.logoUrl}
        storeName={store.name}
        storeSlug={store.slug}
      />
      {children}
    </div>
  );
}
