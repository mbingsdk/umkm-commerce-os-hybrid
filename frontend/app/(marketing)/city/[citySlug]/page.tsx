import type { Metadata } from "next";
import { EmptyState } from "@/components/feedback/empty-state";
import { PublicLinkButton, PublicPageIntro } from "@/components/public/public-ui";
import {
  listDiscoveryProducts,
  listDiscoveryStores
} from "@/features/discovery/api/discovery.api";
import {
  ProductSection,
  StoreSection
} from "@/features/discovery/components/discovery-sections";
import { publicPageMetadata } from "@/lib/seo/metadata";

type CityDiscoveryPageProps = {
  params: Promise<{ citySlug: string }>;
};

export async function generateMetadata({ params }: CityDiscoveryPageProps): Promise<Metadata> {
  const { citySlug } = await params;
  const cityName = cityFromSlug(citySlug);

  return publicPageMetadata({
    title: `UMKM ${cityName} - Toko dan Produk Lokal`,
    description: `Jelajahi toko dan produk UMKM publik dari ${cityName}. Semua hasil mengarah ke storefront tenant.`,
    path: `/city/${citySlug}`
  });
}

export default async function CityDiscoveryPage({ params }: CityDiscoveryPageProps) {
  const { citySlug } = await params;
  const cityName = cityFromSlug(citySlug);

  const [stores, products] = await Promise.all([
    listDiscoveryStores({ city: cityName, limit: 9 }),
    listDiscoveryProducts({ city: cityName, limit: 12 })
  ]);

  const hasResults = stores.items.length > 0 || products.items.length > 0;

  return (
    <main className="min-h-screen bg-[#F8F1E7]">
      <section>
        <div className="mx-auto space-y-5 px-4 py-8 sm:px-6 lg:max-w-6xl lg:px-8">
          <PublicPageIntro
            compact
            eyebrow="Discovery kota"
            title={`UMKM ${cityName}`}
            description={`Lihat toko dan produk UMKM publik dari ${cityName}. Customer tetap diarahkan ke storefront tenant untuk melihat detail produk atau menghubungi toko.`}
          >
            <div className="flex flex-col gap-3 sm:flex-row">
              <PublicLinkButton href={`/stores?city=${encodeURIComponent(cityName)}`} variant="outline">Lihat toko di {cityName}</PublicLinkButton>
              <PublicLinkButton href={`/products?city=${encodeURIComponent(cityName)}`} variant="outline">Lihat produk di {cityName}</PublicLinkButton>
              <PublicLinkButton href={`/search?city=${encodeURIComponent(cityName)}`} variant="outline">Cari di kota ini</PublicLinkButton>
            </div>
          </PublicPageIntro>
        </div>
      </section>

      <section className="mx-auto space-y-10 px-4 pb-10 sm:px-6 lg:max-w-6xl lg:px-8">
        {!hasResults ? (
          <EmptyState
            title="Belum ada hasil di kota ini"
            description="Toko atau produk dari kota ini akan muncul setelah memenuhi aturan discovery publik."
            action={<PublicLinkButton href="/stores" variant="outline">Jelajahi semua toko</PublicLinkButton>}
          />
        ) : null}

        <StoreSection
          title={`Toko di ${cityName}`}
          description="Toko published dari tenant aktif atau trialing."
          stores={stores.items}
          href={`/stores?city=${encodeURIComponent(cityName)}`}
        />

        <ProductSection
          title={`Produk dari ${cityName}`}
          description="Produk aktif dan discoverable dari toko di kota ini."
          products={products.items}
          href={`/products?city=${encodeURIComponent(cityName)}`}
        />
      </section>
    </main>
  );
}

function cityFromSlug(value: string) {
  return decodeURIComponent(value)
    .split("-")
    .filter(Boolean)
    .map((word) => word.slice(0, 1).toUpperCase() + word.slice(1))
    .join(" ");
}
