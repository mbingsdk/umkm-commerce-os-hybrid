import type { Metadata } from "next";
import Link from "next/link";
import { EmptyState } from "@/components/feedback/empty-state";
import { PublicPageIntro } from "@/components/public/public-ui";
import { Button } from "@/components/ui/button";
import { searchDiscovery } from "@/features/discovery/api/discovery.api";
import { DiscoveryProductCard, DiscoveryStoreCard } from "@/features/discovery/components/discovery-cards";
import { publicPageMetadata } from "@/lib/seo/metadata";

type SearchPageProps = {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
};

export const metadata: Metadata = publicPageMetadata({
  title: "Cari Produk dan Toko UMKM",
  description: "Cari produk dan toko publik di UMKM Commerce OS.",
  path: "/search"
});

export default async function SearchPage({ searchParams }: SearchPageProps) {
  const params = await searchParams;
  const query = firstParam(params.q) ?? "";
  const type = normalizeType(firstParam(params.type));
  const result = query
    ? await searchDiscovery({
        query,
        type,
        limit: 12
      })
    : { stores: [], products: [] };

  const showProducts = type === "all" || type === "products";
  const showStores = type === "all" || type === "stores";

  return (
    <main className="min-h-screen bg-[#F8F1E7]">
      <section className="mx-auto space-y-6 px-4 py-8 sm:px-6 lg:max-w-6xl lg:px-8">
        <PublicPageIntro
          compact
          eyebrow="Cari toko dan produk"
          title="Pencarian discovery"
          description="Cari produk atau toko dari UMKM publik. Hasil membawa customer ke storefront resmi tiap toko."
        />

        <form className="flex flex-col gap-3 rounded-[24px] border border-[#E3D2BC] bg-[#FFFDF8] p-4 shadow-[0_10px_28px_rgba(80,57,34,0.06)] sm:flex-row" action="/search">
          <input
            className="h-11 min-w-0 flex-1 rounded-xl border border-[#E3D2BC] bg-white px-3 text-sm text-[#251F1A] outline-none placeholder:text-[#9b8d7b] focus:border-[#B96E45] focus:ring-2 focus:ring-[#F1E7D8]"
            defaultValue={query}
            name="q"
            placeholder="Cari produk atau toko..."
          />
          <input name="type" type="hidden" value={type} />
          <Button className="h-11 bg-[#251F1A] hover:bg-[#16110E]" type="submit">Cari</Button>
        </form>

        <div className="flex flex-wrap gap-2">
          <TabLink active={type === "all"} href={buildSearchHref(query, "all")} label="Semua" />
          <TabLink active={type === "products"} href={buildSearchHref(query, "products")} label="Produk" />
          <TabLink active={type === "stores"} href={buildSearchHref(query, "stores")} label="Toko" />
        </div>

        {!query ? (
          <EmptyState
            title="Masukkan kata kunci"
            description="Contoh: bouquet, makanan, Makassar, atau nama toko."
          />
        ) : (
          <div className="space-y-10">
            {showProducts ? (
              <section className="space-y-4">
                <SectionTitle title="Produk" count={result.products.length} />
                {result.products.length === 0 ? (
                  <EmptyState title="Produk tidak ditemukan" description="Coba kata kunci lain atau lihat semua produk." />
                ) : (
                  <div className="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-4">
                    {result.products.map((product) => (
                      <DiscoveryProductCard key={product.id} product={product} />
                    ))}
                  </div>
                )}
              </section>
            ) : null}

            {showStores ? (
              <section className="space-y-4">
                <SectionTitle title="Toko" count={result.stores.length} />
                {result.stores.length === 0 ? (
                  <EmptyState title="Toko tidak ditemukan" description="Coba kata kunci lain atau lihat semua toko." />
                ) : (
                  <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                    {result.stores.map((store) => (
                      <DiscoveryStoreCard key={store.id} store={store} />
                    ))}
                  </div>
                )}
              </section>
            ) : null}
          </div>
        )}
      </section>
    </main>
  );
}

function TabLink({ active, href, label }: { active: boolean; href: string; label: string }) {
  return (
    <Link
      href={href}
      className={
        active
          ? "inline-flex min-h-11 items-center rounded-full bg-[#251F1A] px-4 py-2 text-sm font-semibold text-[#FFFDF8]"
          : "inline-flex min-h-11 items-center rounded-full border border-[#E3D2BC] bg-[#FFFDF8] px-4 py-2 text-sm font-semibold text-[#6F6256] transition hover:border-[#9a6a43] hover:text-[#B96E45]"
      }
    >
      {label}
    </Link>
  );
}

function SectionTitle({ title, count }: { title: string; count: number }) {
  return (
    <div>
      <h2 className="text-xl font-bold text-[#251F1A]">{title}</h2>
      <p className="mt-1 text-sm text-[#7a6a58]">{count} hasil ditemukan.</p>
    </div>
  );
}

function buildSearchHref(query: string, type: "all" | "products" | "stores") {
  const params = new URLSearchParams();
  if (query) {
    params.set("q", query);
  }
  params.set("type", type);
  return `/search?${params.toString()}`;
}

function normalizeType(value?: string): "all" | "products" | "stores" {
  if (value === "products" || value === "stores") {
    return value;
  }
  return "all";
}

function firstParam(value?: string | string[]) {
  return Array.isArray(value) ? value[0] : value;
}
