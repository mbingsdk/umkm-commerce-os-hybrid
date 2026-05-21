import Link from "next/link";
import { EmptyState } from "@/components/feedback/empty-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { DiscoveryProductCard, DiscoveryStoreCard } from "@/features/discovery/components/discovery-cards";
import type { DiscoveryAggregate, DiscoveryProduct, DiscoveryStore, Pagination } from "@/features/discovery/types";

export function PlatformHero({ query }: { query?: string }) {
  return (
    <section className="border-b border-neutral-200 bg-gradient-to-br from-primary-50 via-white to-neutral-50">
      <div className="mx-auto grid max-w-6xl gap-8 px-4 py-12 sm:px-6 lg:grid-cols-[minmax(0,1fr)_360px] lg:px-8 lg:py-16">
        <div>
          <p className="text-sm font-semibold text-primary-700">Platform discovery UMKM Indonesia</p>
          <h1 className="mt-4 text-4xl font-bold tracking-tight text-neutral-950 sm:text-5xl">
            Temukan toko lokal, produk siap jual, dan storefront UMKM dalam satu tempat.
          </h1>
          <p className="mt-5 max-w-2xl text-base leading-7 text-neutral-600">
            UMKM Commerce OS Hybrid membantu toko punya storefront sendiri, sambil tetap bisa ditemukan customer
            lewat discovery platform.
          </p>

          <form className="mt-8 flex flex-col gap-3 rounded-3xl border border-neutral-200 bg-white p-3 shadow-soft sm:flex-row" action="/search">
            <input
              className="h-12 min-w-0 flex-1 rounded-2xl border border-neutral-200 px-4 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
              defaultValue={query}
              name="q"
              placeholder="Cari bouquet, makanan, toko Makassar..."
            />
            <Button type="submit" size="lg">
              Cari
            </Button>
          </form>

          <div className="mt-5 flex flex-wrap gap-3">
            <LinkButton href="/explore">Jelajahi platform</LinkButton>
            <LinkButton href="/register" variant="primary">Daftar UMKM</LinkButton>
          </div>
        </div>

        <Card className="self-start border-primary-100 bg-white/85">
          <CardHeader>
            <CardTitle>Kenapa discovery hybrid?</CardTitle>
            <CardDescription>Customer masuk dari platform, transaksi tetap diarahkan ke toko tenant.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 text-sm leading-6 text-neutral-600">
            <p>• Toko punya halaman publik SEO-friendly.</p>
            <p>• Produk discovery mengarah ke storefront tenant.</p>
            <p>• MVP tidak membuat marketplace checkout lintas toko.</p>
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
    <section className="space-y-4">
      <SectionHeader title={title} description={description} href={href} />
      {stores.length === 0 ? (
        <EmptyState title="Belum ada toko" description="Toko publik yang memenuhi filter akan muncul di sini." />
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
    <section className="space-y-4">
      <SectionHeader title={title} description={description} href={href} />
      {products.length === 0 ? (
        <EmptyState title="Belum ada produk" description="Produk aktif dan discoverable akan muncul di sini." />
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
    <Card className="shadow-none">
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-wrap gap-2">
        {items.map((item) => {
          const label = type === "city" ? item.city : item.name;
          const value = type === "city" ? item.city : item.slug;
          if (!label || !value) {
            return null;
          }

          const href = type === "city" ? `/stores?city=${encodeURIComponent(value)}` : `/products?category=${encodeURIComponent(value)}`;
          return (
            <Link
              key={`${type}-${value}`}
              href={href}
              className="rounded-full border border-neutral-200 bg-white px-4 py-2 text-sm font-semibold text-neutral-700 transition hover:border-primary-300 hover:text-primary-700"
            >
              {label} <span className="text-neutral-400">({item.count})</span>
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
    <div className="flex items-center justify-between gap-3">
      <LinkButton href={buildHref(basePath, { ...searchParams, cursor: undefined })}>Reset halaman</LinkButton>
      <p className="text-sm text-neutral-500">Menampilkan maksimal {pagination.limit} item.</p>
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
    <form className="grid gap-3 rounded-2xl border border-neutral-200 bg-white p-4 lg:grid-cols-[1.2fr_180px_180px_auto]" action={action}>
      <input
        className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
        defaultValue={query}
        name="q"
        placeholder="Cari nama, produk, atau toko..."
      />
      <select className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100" defaultValue={city} name="city">
        <option value="">Semua kota</option>
        {cities.map((item) =>
          item.city ? (
            <option key={item.city} value={item.city}>
              {item.city}
            </option>
          ) : null
        )}
      </select>
      <select className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100" defaultValue={category} name="category">
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
            className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            defaultValue={priceMin}
            min={0}
            name="price_min"
            placeholder="Harga min"
            type="number"
          />
          <input
            className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            defaultValue={priceMax}
            min={0}
            name="price_max"
            placeholder="Harga max"
            type="number"
          />
        </div>
      ) : null}
      <Button type="submit">Terapkan</Button>
    </form>
  );
}

function SectionHeader({ title, description, href }: { title: string; description: string; href?: string }) {
  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h2 className="text-xl font-semibold text-neutral-950">{title}</h2>
        <p className="mt-1 text-sm leading-6 text-neutral-500">{description}</p>
      </div>
      {href ? (
        <Link className="text-sm font-semibold text-primary-700 hover:text-primary-800" href={href}>
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
          ? "inline-flex h-10 items-center justify-center rounded-xl bg-primary-600 px-4 text-sm font-semibold text-white transition hover:bg-primary-700"
          : "inline-flex h-10 items-center justify-center rounded-xl border border-neutral-300 bg-white px-4 text-sm font-semibold text-neutral-900 transition hover:bg-neutral-50"
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
