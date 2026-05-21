import type { Metadata } from "next";
import Link from "next/link";
import { EmptyState } from "@/components/feedback/empty-state";
import { Button } from "@/components/ui/button";
import { searchDiscovery } from "@/features/discovery/api/discovery.api";
import { DiscoveryProductCard, DiscoveryStoreCard } from "@/features/discovery/components/discovery-cards";

type SearchPageProps = {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
};

export const metadata: Metadata = {
  title: "Cari Produk dan Toko UMKM",
  description: "Cari produk dan toko publik di UMKM Commerce OS."
};

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
    <main className="min-h-screen bg-neutral-50">
      <section className="mx-auto space-y-6 px-4 py-8 sm:px-6 lg:max-w-6xl lg:px-8">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-neutral-950">Pencarian discovery</h1>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-neutral-500">
            Cari produk atau toko. Hasil produk dan toko akan membawa customer ke storefront tenant.
          </p>
        </div>

        <form className="flex flex-col gap-3 rounded-2xl border border-neutral-200 bg-white p-4 sm:flex-row" action="/search">
          <input
            className="h-10 min-w-0 flex-1 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            defaultValue={query}
            name="q"
            placeholder="Cari produk atau toko..."
          />
          <input name="type" type="hidden" value={type} />
          <Button type="submit">Cari</Button>
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
          ? "rounded-full bg-primary-600 px-4 py-2 text-sm font-semibold text-white"
          : "rounded-full border border-neutral-200 bg-white px-4 py-2 text-sm font-semibold text-neutral-700 transition hover:border-primary-300 hover:text-primary-700"
      }
    >
      {label}
    </Link>
  );
}

function SectionTitle({ title, count }: { title: string; count: number }) {
  return (
    <div>
      <h2 className="text-xl font-semibold text-neutral-950">{title}</h2>
      <p className="mt-1 text-sm text-neutral-500">{count} hasil ditemukan.</p>
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
