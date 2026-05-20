import { CheckoutPage } from "@/features/storefront/components/checkout-page";

type PageProps = {
  params: Promise<{ storeSlug: string }>;
};

export default async function StorefrontCheckoutRoute({ params }: PageProps) {
  const { storeSlug } = await params;
  return <CheckoutPage storeSlug={storeSlug} />;
}
