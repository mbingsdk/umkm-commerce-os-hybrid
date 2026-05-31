import type { Metadata } from "next";
import {
  getDiscoveryHome,
  listDiscoveryProducts
} from "@/features/discovery/api/discovery.api";
import {
  FilterBar,
  PaginationLinks,
  ProductSection
} from "@/features/discovery/components/discovery-sections";
import { PublicPageIntro } from "@/components/public/public-ui";
import { publicPageMetadata } from "@/lib/seo/metadata";

type ProductsPageProps = {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
};

export const metadata: Metadata = publicPageMetadata({
  title: "Produk UMKM",
  description: "Cari produk aktif dari storefront UMKM yang tampil di discovery platform.",
  path: "/products"
});

export default async function ProductsPage({ searchParams }: ProductsPageProps) {
  const params = await searchParams;
  const query = firstParam(params.q);
  const city = firstParam(params.city);
  const category = firstParam(params.category);
  const priceMin = firstParam(params.price_min);
  const priceMax = firstParam(params.price_max);
  const cursor = firstParam(params.cursor);

  const [home, products] = await Promise.all([
    getDiscoveryHome(),
    listDiscoveryProducts({
      query,
      city,
      category,
      priceMin,
      priceMax,
      cursor,
      limit: 16
    })
  ]);

  return (
    <main className="min-h-screen bg-[#F8F1E7]">
      <section className="mx-auto space-y-6 px-4 py-8 sm:px-6 lg:max-w-6xl lg:px-8">
        <PublicPageIntro
          compact
          eyebrow="Etalase produk"
          title="Produk UMKM"
          description="Jelajahi produk aktif dari toko publik. Setiap produk mengarah ke storefront resmi toko masing-masing."
        />

        <FilterBar
          action="/products"
          query={query}
          city={city}
          category={category}
          priceMin={priceMin}
          priceMax={priceMax}
          showPrice
          categories={home.popularCategories}
          cities={home.popularCities}
        />

        <ProductSection
          title="Hasil produk"
          description={query || city || category ? "Produk yang cocok dengan filtermu." : "Semua produk publik terbaru."}
          products={products.items}
        />

        <PaginationLinks
          basePath="/products"
          searchParams={{ q: query, city, category, price_min: priceMin, price_max: priceMax, cursor }}
          pagination={products.pagination}
        />
      </section>
    </main>
  );
}

function firstParam(value?: string | string[]) {
  return Array.isArray(value) ? value[0] : value;
}
