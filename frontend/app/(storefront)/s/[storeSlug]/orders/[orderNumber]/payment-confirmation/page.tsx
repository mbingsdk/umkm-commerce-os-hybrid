import type { Metadata } from "next";
import { PaymentConfirmationPage } from "@/features/storefront/components/payment-confirmation-page";

type PageProps = {
  params: Promise<{
    storeSlug: string;
    orderNumber: string;
  }>;
};

export const metadata: Metadata = {
  title: "Konfirmasi pembayaran",
  description: "Kirim konfirmasi pembayaran manual untuk pesanan toko."
};

export default async function StorefrontPaymentConfirmationRoute({ params }: PageProps) {
  const { storeSlug, orderNumber } = await params;

  return <PaymentConfirmationPage orderNumber={orderNumber} storeSlug={storeSlug} />;
}
