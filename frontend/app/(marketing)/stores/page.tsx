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
    <main className="min-h-screen bg-[#f7f1e8]">
      <section className="mx-auto space-y-6 px-4 py-8 sm:px-6 lg:max-w-6xl lg:px-8">
        <div className="rounded-[28px] border border-[#eadfce] bg-[#fffaf2] p-5 shadow-[0_14px_40px_rgba(89,63,38,0.07)] sm:p-7">
          <p className="text-sm font-semibold text-[#7a4f2f]">Direktori toko publik</p>
          <h1 className="mt-3 text-3xl font-bold tracking-tight text-[#241c16] sm:text-5xl">Toko UMKM</h1>
          <p className="mt-3 max-w-2xl text-sm leading-7 text-[#6d5e4e]">
            Temukan storefront UMKM yang sudah published. Klik kartu toko untuk melihat produk, kontak, dan cara belanja
            langsung dari toko tersebut.
          </p>
        </div>

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
