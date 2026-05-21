import type { Metadata } from "next";
import Link from "next/link";
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
  title: "UMKM Commerce OS - Discovery Toko dan Produk Lokal",
  description: "Temukan toko UMKM, produk lokal, dan storefront publik dari tenant UMKM Commerce OS."
};

export default async function PlatformHomePage() {
  const [home, stores, products] = await Promise.all([
    getDiscoveryHome(),
    listDiscoveryStores({ limit: 6 }),
    listDiscoveryProducts({ limit: 8 })
  ]);

  const featuredStores = home.featuredStores.length > 0 ? home.featuredStores : stores.items;
  const featuredProducts = home.featuredProducts.length > 0 ? home.featuredProducts : products.items;

  return (
    <main className="min-h-screen bg-neutral-50">
      <PlatformHero />

      <section className="mx-auto space-y-10 px-4 py-10 sm:px-6 lg:max-w-6xl lg:px-8">
        <div className="grid gap-4 lg:grid-cols-2">
          <AggregateChips title="Kategori populer" items={home.popularCategories} type="category" />
          <AggregateChips title="Kota populer" items={home.popularCities} type="city" />
        </div>

        <StoreSection
          title="Toko pilihan"
          description="Storefront tenant yang sudah published dan bisa ditemukan customer."
          stores={featuredStores}
          href="/stores"
        />

        <ProductSection
          title="Produk pilihan"
          description="Produk aktif dan discoverable. Klik produk untuk masuk ke halaman toko tenant."
          products={featuredProducts}
          href="/products"
        />

        <div className="rounded-3xl border border-primary-100 bg-primary-700 p-6 text-white shadow-soft sm:p-8">
          <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center">
            <div>
              <h2 className="text-2xl font-bold tracking-tight">Punya toko UMKM?</h2>
              <p className="mt-2 max-w-2xl text-sm leading-6 text-primary-50">
                Buat storefront, kelola produk, pesanan, POS, inventori, finance dasar, dan mulai tampil di discovery.
              </p>
            </div>
            <Link
              className="inline-flex h-11 items-center justify-center rounded-xl bg-white px-5 text-sm font-semibold text-primary-800 transition hover:bg-primary-50"
              href="/register"
            >
              Daftar sebagai UMKM
            </Link>
          </div>
        </div>
      </section>
    </main>
  );
}
