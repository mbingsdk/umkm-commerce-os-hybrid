import type { Metadata } from "next";
import {
  getDiscoveryHome,
  listDiscoveryProducts,
  listDiscoveryStores
} from "@/features/discovery/api/discovery.api";
import {
  AggregateChips,
  PlatformHero,
  ProductSection,
  StoreSection
} from "@/features/discovery/components/discovery-sections";

export const metadata: Metadata = {
  title: "Explore UMKM - Toko dan Produk Lokal",
  description: "Jelajahi toko dan produk publik dari tenant UMKM Commerce OS."
};

export default async function ExplorePage() {
  const [home, stores, products] = await Promise.all([
    getDiscoveryHome(),
    listDiscoveryStores({ limit: 9 }),
    listDiscoveryProducts({ limit: 12 })
  ]);

  return (
    <main className="min-h-screen bg-neutral-50">
      <PlatformHero />
      <section className="mx-auto space-y-10 px-4 py-10 sm:px-6 lg:max-w-6xl lg:px-8">
        <div className="grid gap-4 lg:grid-cols-2">
          <AggregateChips title="Kategori populer" items={home.popularCategories} type="category" />
          <AggregateChips title="Kota populer" items={home.popularCities} type="city" />
        </div>

        <StoreSection
          title="Toko terbaru"
          description="Toko lokal yang sudah published dan memenuhi aturan discovery."
          stores={stores.items}
          href="/stores"
        />

        <ProductSection
          title="Produk terbaru"
          description="Produk discovery terbaru dari berbagai storefront tenant."
          products={products.items}
          href="/products"
        />
      </section>
    </main>
  );
}
