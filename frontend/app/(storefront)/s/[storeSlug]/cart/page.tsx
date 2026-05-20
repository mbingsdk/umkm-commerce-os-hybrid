import { CartPage } from "@/features/storefront/components/cart-page";

type PageProps = {
  params: Promise<{ storeSlug: string }>;
};

export default async function StorefrontCartRoute({ params }: PageProps) {
  const { storeSlug } = await params;
  return <CartPage storeSlug={storeSlug} />;
}
