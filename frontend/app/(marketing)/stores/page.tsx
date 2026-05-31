import type { Metadata } from "next";
import {
  getDiscoveryHome,
  listDiscoveryStores
} from "@/features/discovery/api/discovery.api";
import {
  FilterBar,
  PaginationLinks,
  StoreSection
} from "@/features/discovery/components/discovery-sections";
import { PublicPageIntro } from "@/components/public/public-ui";
import { publicPageMetadata } from "@/lib/seo/metadata";

type StoresPageProps = {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
};

export const metadata: Metadata = publicPageMetadata({
  title: "Daftar Toko UMKM",
  description: "Cari toko UMKM berdasarkan nama, kota, dan kategori bisnis.",
  path: "/stores"
});

export default async function StoresPage({ searchParams }: StoresPageProps) {
  const params = await searchParams;
  const query = firstParam(params.q);
  const city = firstParam(params.city);
  const category = firstParam(params.category);
  const cursor = firstParam(params.cursor);

  const [home, stores] = await Promise.all([
    getDiscoveryHome(),
    listDiscoveryStores({
      query,
      city,
      category,
      cursor,
      limit: 12
    })
  ]);

  return (
    <main className="min-h-screen bg-[#F8F1E7]">
      <section className="mx-auto space-y-6 px-4 py-8 sm:px-6 lg:max-w-6xl lg:px-8">
        <PublicPageIntro
          compact
          eyebrow="Direktori toko publik"
          title="Toko UMKM"
          description="Temukan storefront UMKM yang sudah published. Klik kartu toko untuk melihat katalog, kontak, dan cara belanja."
        />

        <FilterBar
          action="/stores"
          query={query}
          city={city}
          category={category}
          categories={home.popularCategories}
          cities={home.popularCities}
        />

        <StoreSection
          title="Hasil toko"
          description={query || city || category ? "Toko yang cocok dengan filtermu." : "Semua toko publik terbaru."}
          stores={stores.items}
        />

        <PaginationLinks
          basePath="/stores"
          searchParams={{ q: query, city, category, cursor }}
          pagination={stores.pagination}
        />
      </section>
    </main>
  );
}

function firstParam(value?: string | string[]) {
  return Array.isArray(value) ? value[0] : value;
}
