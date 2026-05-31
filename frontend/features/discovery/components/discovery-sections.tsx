import Link from "next/link";
import { EmptyState } from "@/components/feedback/empty-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { DiscoveryProductCard, DiscoveryStoreCard } from "@/features/discovery/components/discovery-cards";
import type { DiscoveryAggregate, DiscoveryProduct, DiscoveryStore, Pagination } from "@/features/discovery/types";

export function PlatformHero({ query }: { query?: string }) {
  return (
    <section className="bg-[#f7f1e8]">
      <div className="mx-auto grid max-w-6xl gap-8 px-4 py-10 sm:px-6 lg:grid-cols-[minmax(0,1fr)_380px] lg:px-8 lg:py-16">
        <div>
          <p className="inline-flex rounded-full border border-[#e2d4bf] bg-[#fffaf2] px-3 py-1 text-sm font-semibold text-[#7a4f2f]">
            Platform discovery UMKM Indonesia
          </p>
          <h1 className="mt-5 max-w-4xl text-4xl font-bold tracking-tight text-[#241c16] sm:text-6xl">
            Temukan toko lokal dan produk siap beli dari UMKM sekitar.
          </h1>
          <p className="mt-5 max-w-2xl text-base leading-8 text-[#6d5e4e] sm:text-lg">
            Customer bisa menemukan produk dari platform, lalu masuk ke storefront masing-masing toko untuk melihat detail,
            checkout, dan konfirmasi pembayaran manual dengan jelas.
          </p>

          <form
            className="mt-8 flex flex-col gap-3 rounded-[28px] border border-[#eadfce] bg-[#fffaf2] p-3 shadow-[0_20px_60px_rgba(89,63,38,0.10)] sm:flex-row"
            action="/search"
          >
            <input
              className="h-12 min-w-0 flex-1 rounded-2xl border border-[#eadfce] bg-white px-4 text-sm text-[#241c16] outline-none placeholder:text-[#a0917f] focus:border-[#9a6a43] focus:ring-2 focus:ring-[#ead7bd]"
              defaultValue={query}
              name="q"
              placeholder="Cari bouquet, makanan, toko Makassar..."
            />
            <Button className="bg-[#2f2923] hover:bg-[#1f1a16]" type="submit" size="lg">
              Cari
            </Button>
          </form>

          <div className="mt-5 flex flex-wrap gap-3">
            <LinkButton href="/explore">Jelajahi platform</LinkButton>
            <LinkButton href="/register" variant="primary">
              Daftar UMKM
            </LinkButton>
          </div>
        </div>

        <Card className="self-start border-[#eadfce] bg-[#fffaf2]/90 shadow-[0_18px_50px_rgba(89,63,38,0.08)]">
          <CardHeader>
            <CardTitle className="text-[#241c16]">Belanja tetap dekat dengan tokonya.</CardTitle>
            <CardDescription className="text-[#7a6a58]">
              Discovery membantu customer menemukan, storefront membantu toko melayani.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 text-sm leading-6 text-[#6d5e4e]">
            {[
              "Tiap toko punya etalase publik sendiri.",
              "Produk discovery mengarah ke halaman produk toko.",
              "Checkout tetap satu toko per transaksi, lebih sederhana untuk pilot."
            ].map((item) => (
              <div key={item} className="flex gap-3 rounded-2xl bg-white/70 p-3">
                <span className="mt-1 h-2 w-2 rounded-full bg-[#9a6a43]" />
                <p>{item}</p>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </section>
  );
}

export function StoreSection({
  title,
  description,
  stores,
  href
}: {
  title: string;
  description: string;
  stores: DiscoveryStore[];
  href?: string;
}) {
  return (
    <section className="space-y-5">
      <SectionHeader title={title} description={description} href={href} />
      {stores.length === 0 ? (
        <EmptyState title="Belum ada toko yang cocok" description="Toko publik yang memenuhi filter akan tampil di sini setelah tersedia." />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {stores.map((store) => (
            <DiscoveryStoreCard key={store.id} store={store} />
          ))}
        </div>
      )}
    </section>
  );
}

export function ProductSection({
  title,
  description,
  products,
  href
}: {
  title: string;
  description: string;
  products: DiscoveryProduct[];
  href?: string;
}) {
  return (
    <section className="space-y-5">
      <SectionHeader title={title} description={description} href={href} />
      {products.length === 0 ? (
        <EmptyState title="Belum ada produk yang cocok" description="Produk aktif dan discoverable akan tampil di sini setelah tersedia." />
      ) : (
        <div className="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-4">
          {products.map((product) => (
            <DiscoveryProductCard key={product.id} product={product} />
          ))}
        </div>
      )}
    </section>
  );
}

export function AggregateChips({
  title,
  items,
  type
}: {
  title: string;
  items: DiscoveryAggregate[];
  type: "category" | "city";
}) {
  if (items.length === 0) {
    return null;
  }

  return (
    <Card className="border-[#eadfce] bg-[#fffaf2] shadow-none">
      <CardHeader>
        <CardTitle className="text-[#241c16]">{title}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-wrap gap-2">
        {items.map((item) => {
          const label = type === "city" ? item.city : item.name;
          const value = type === "city" ? item.city : item.slug;
          if (!label || !value) {
            return null;
          }

          const href = type === "city" ? `/city/${slugifyCity(value)}` : `/category/${encodeURIComponent(value)}`;
          return (
            <Link
              key={`${type}-${value}`}
              href={href}
              className="inline-flex min-h-11 items-center rounded-full border border-[#e2d4bf] bg-white px-4 py-2 text-sm font-semibold text-[#5f5042] transition hover:border-[#9a6a43] hover:text-[#7a4f2f]"
            >
              {label} <span className="text-[#a0917f]">({item.count})</span>
            </Link>
          );
        })}
      </CardContent>
    </Card>
  );
}

export function PaginationLinks({
  basePath,
  searchParams,
  pagination
}: {
  basePath: string;
  searchParams: Record<string, string | undefined>;
  pagination: Pagination;
}) {
  if (!pagination.hasMore && !searchParams.cursor) {
    return null;
  }

  return (
    <div className="flex flex-col gap-3 rounded-3xl border border-[#eadfce] bg-[#fffaf2] p-4 sm:flex-row sm:items-center sm:justify-between">
      <LinkButton href={buildHref(basePath, { ...searchParams, cursor: undefined })}>Reset halaman</LinkButton>
      <p className="text-sm text-[#7a6a58]">Menampilkan maksimal {pagination.limit} item.</p>
      {pagination.hasMore && pagination.nextCursor ? (
        <LinkButton href={buildHref(basePath, { ...searchParams, cursor: pagination.nextCursor })}>Berikutnya</LinkButton>
      ) : (
        <Button type="button" variant="outline" disabled>
          Berikutnya
        </Button>
      )}
    </div>
  );
}

export function FilterBar({
  action,
  query,
  city,
  category,
  priceMin,
  priceMax,
  showPrice = false,
  categories = [],
  cities = []
}: {
  action: string;
  query?: string;
  city?: string;
  category?: string;
  priceMin?: string;
  priceMax?: string;
  showPrice?: boolean;
  categories?: DiscoveryAggregate[];
  cities?: DiscoveryAggregate[];
}) {
  return (
    <form
      className="grid gap-3 rounded-[28px] border border-[#eadfce] bg-[#fffaf2] p-4 shadow-[0_14px_40px_rgba(89,63,38,0.07)] lg:grid-cols-[1.2fr_180px_180px_auto]"
      action={action}
    >
      <input
        className="h-11 rounded-xl border border-[#e2d4bf] bg-white px-3 text-sm text-[#241c16] outline-none placeholder:text-[#a0917f] focus:border-[#9a6a43] focus:ring-2 focus:ring-[#ead7bd]"
        defaultValue={query}
        name="q"
        placeholder="Cari nama, produk, atau toko..."
      />
      <select
        className="h-11 rounded-xl border border-[#e2d4bf] bg-white px-3 text-sm text-[#241c16] outline-none focus:border-[#9a6a43] focus:ring-2 focus:ring-[#ead7bd]"
        defaultValue={city}
        name="city"
      >
        <option value="">Semua kota</option>
        {cities.map((item) =>
          item.city ? (
            <option key={item.city} value={item.city}>
              {item.city}
            </option>
          ) : null
        )}
      </select>
      <select
        className="h-11 rounded-xl border border-[#e2d4bf] bg-white px-3 text-sm text-[#241c16] outline-none focus:border-[#9a6a43] focus:ring-2 focus:ring-[#ead7bd]"
        defaultValue={category}
        name="category"
      >
        <option value="">Semua kategori</option>
        {categories.map((item) =>
          item.slug ? (
            <option key={item.slug} value={item.slug}>
              {item.name ?? item.slug}
            </option>
          ) : null
        )}
      </select>
      {showPrice ? (
        <div className="grid gap-3 sm:grid-cols-2 lg:col-span-3">
          <input
            className="h-11 rounded-xl border border-[#e2d4bf] bg-white px-3 text-sm text-[#241c16] outline-none placeholder:text-[#a0917f] focus:border-[#9a6a43] focus:ring-2 focus:ring-[#ead7bd]"
            defaultValue={priceMin}
            min={0}
            name="price_min"
            placeholder="Harga min"
            type="number"
          />
          <input
            className="h-11 rounded-xl border border-[#e2d4bf] bg-white px-3 text-sm text-[#241c16] outline-none placeholder:text-[#a0917f] focus:border-[#9a6a43] focus:ring-2 focus:ring-[#ead7bd]"
            defaultValue={priceMax}
            min={0}
            name="price_max"
            placeholder="Harga max"
            type="number"
          />
        </div>
      ) : null}
      <Button className="h-11 w-full bg-[#2f2923] hover:bg-[#1f1a16] lg:w-auto" type="submit">
        Terapkan
      </Button>
    </form>
  );
}

function SectionHeader({ title, description, href }: { title: string; description: string; href?: string }) {
  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h2 className="text-2xl font-bold tracking-tight text-[#241c16]">{title}</h2>
        <p className="mt-1 max-w-2xl text-sm leading-6 text-[#6d5e4e]">{description}</p>
      </div>
      {href ? (
        <Link className="inline-flex min-h-11 items-center text-sm font-semibold text-[#7a4f2f] hover:text-[#4e321f]" href={href}>
          Lihat semua →
        </Link>
      ) : null}
    </div>
  );
}

function LinkButton({
  href,
  children,
  variant = "outline"
}: {
  href: string;
  children: string;
  variant?: "outline" | "primary";
}) {
  return (
    <Link
      href={href}
      className={
        variant === "primary"
          ? "inline-flex h-11 items-center justify-center rounded-xl bg-[#2f2923] px-4 text-sm font-semibold text-[#fffaf2] transition hover:bg-[#1f1a16]"
          : "inline-flex h-11 items-center justify-center rounded-xl border border-[#d9c8af] bg-white px-4 text-sm font-semibold text-[#3d3128] transition hover:bg-[#f4eadb]"
      }
    >
      {children}
    </Link>
  );
}

function buildHref(basePath: string, params: Record<string, string | undefined>) {
  const searchParams = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value) {
      searchParams.set(key, value);
    }
  });

  const suffix = searchParams.size > 0 ? `?${searchParams.toString()}` : "";
  return `${basePath}${suffix}`;
}

function slugifyCity(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}
