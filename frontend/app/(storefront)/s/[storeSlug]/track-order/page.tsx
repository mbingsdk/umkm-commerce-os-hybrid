import type { Metadata } from "next";
import { TrackOrderPage } from "@/features/storefront/components/track-order-page";

type TrackOrderRouteProps = {
  params: Promise<{ storeSlug: string }>;
};

export const metadata: Metadata = {
  title: "Lacak Pesanan",
  description: "Cek status pesanan dan pengiriman toko."
};

export default async function TrackOrderRoute({ params }: TrackOrderRouteProps) {
  const { storeSlug } = await params;

  return <TrackOrderPage storeSlug={storeSlug} />;
}
