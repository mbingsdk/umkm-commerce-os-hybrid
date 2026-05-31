import type { Metadata } from "next";
import {
  getDiscoveryHome,
  listDiscoveryProducts,
  listDiscoveryStores
} from "@/features/discovery/api/discovery.api";
import {
  ProductDiscoveryPanel,
  ProductSection,
  StoreSection
} from "@/features/discovery/components/discovery-sections";
import { publicPageMetadata } from "@/lib/seo/metadata";

export const metadata: Metadata = publicPageMetadata({
  title: "UMKM Commerce OS - Discovery Toko dan Produk Lokal",
  description: "Temukan toko UMKM, produk lokal, dan storefront publik dari tenant UMKM Commerce OS.",
  path: "/"
});

type PlatformHomePageProps = {
  searchParams: Promise<{
    q?: string | string[];
    category?: string | string[];
    city?: string | string[];
    price_min?: string | string[];
    price_max?: string | string[];
  }>;
};

export default async function PlatformHomePage({ searchParams }: PlatformHomePageProps) {
  const rawSearchParams = await searchParams;
  const query = firstParam(rawSearchParams.q);
  const category = firstParam(rawSearchParams.category);
  const city = firstParam(rawSearchParams.city);
  const priceMin = firstParam(rawSearchParams.price_min);
  const priceMax = firstParam(rawSearchParams.price_max);

  const [home, stores, products] = await Promise.all([
    getDiscoveryHome(),
    listDiscoveryStores({ limit: 3 }),
    listDiscoveryProducts({
      query,
      category,
      city,
      priceMin,
      priceMax,
      limit: 12
    })
  ]);

  const featuredStores = home.featuredStores.length > 0 ? home.featuredStores : stores.items;

  return (
    <main className="min-h-screen bg-[#F8F1E7]">
      <ProductDiscoveryPanel
        categories={home.popularCategories}
        category={category}
        cities={home.popularCities}
        city={city}
        priceMax={priceMax}
        priceMin={priceMin}
        query={query}
      />

      <section className="mx-auto space-y-5 px-4 pb-8 pt-2 sm:px-6 sm:pb-10 lg:max-w-6xl lg:px-8">
        <ProductSection
          title="Produk UMKM terbaru"
          description="Produk aktif dari UMKM lokal, langsung menuju halaman toko resminya."
          products={products.items}
          href="/products"
        />

        <StoreSection
          title="Toko lokal pilihan"
          description="Storefront publik sebagai konteks toko, setelah kamu melihat produk yang tersedia."
          stores={featuredStores.slice(0, 3)}
          href="/stores"
        />
      </section>
    </main>
  );
}

function firstParam(value?: string | string[]) {
  return Array.isArray(value) ? value[0] : value;
}
