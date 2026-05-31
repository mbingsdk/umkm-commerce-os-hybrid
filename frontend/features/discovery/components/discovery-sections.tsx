import Link from "next/link";
import type { ReactNode } from "react";
import { EmptyState } from "@/components/feedback/empty-state";
import { PublicSectionHeader, publicTheme } from "@/components/public/public-ui";
import { Button } from "@/components/ui/button";
import { DiscoveryProductCard, DiscoveryStoreCard } from "@/features/discovery/components/discovery-cards";
import type { DiscoveryAggregate, DiscoveryProduct, DiscoveryStore, Pagination } from "@/features/discovery/types";
import { cn } from "@/lib/utils/cn";

export function PlatformHero({ query }: { query?: string }) {
  return (
    <section className={publicTheme.bg}>
      <div className="mx-auto max-w-6xl px-4 pb-3 pt-4 sm:px-6 sm:pb-5 sm:pt-6 lg:px-8">
        <div className={cn(publicTheme.card, "p-4 sm:p-5")}>
          <div className="grid gap-4 lg:grid-cols-[minmax(0,0.92fr)_minmax(360px,1fr)] lg:items-end">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.18em] text-[#B96E45]">Marketplace UMKM lokal</p>
              <h1 className="mt-2 max-w-2xl text-2xl font-bold tracking-tight text-[#251F1A] sm:text-4xl">
                Temukan produk UMKM lokal tanpa ribet.
              </h1>
              <p className="mt-2 max-w-xl text-sm leading-6 text-[#6F6256] sm:text-base">
                Cari toko, lihat katalog, lalu checkout langsung di storefront resmi masing-masing UMKM.
              </p>
            </div>

            <div className="space-y-2">
              <form action="/search" className="flex gap-2 rounded-2xl border border-[#E3D2BC] bg-white p-2 shadow-sm">
                <input
                  className="h-10 min-w-0 flex-1 rounded-xl bg-transparent px-2 text-sm text-[#251F1A] outline-none placeholder:text-[#9b8d7b]"
                  defaultValue={query}
                  name="q"
                  placeholder="Cari bouquet, hampers, makanan, toko di Makassar..."
                />
                <Button className="h-10 shrink-0 bg-[#251F1A] px-4 hover:bg-[#16110E]" type="submit">
                  Cari
                </Button>
              </form>
              <div className="flex gap-2 overflow-x-auto pb-1">
                <QuickChip href="/search?q=bouquet">Bouquet</QuickChip>
                <QuickChip href="/search?q=hampers">Hampers</QuickChip>
                <QuickChip href="/search?q=makanan">Makanan</QuickChip>
                <QuickChip href="/city/makassar">Makassar</QuickChip>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

export function ProductDiscoveryPanel({
  query,
  category,
  city,
  priceMin,
  priceMax,
  categories,
  cities
}: {
  query?: string;
  category?: string;
  city?: string;
  priceMin?: string;
  priceMax?: string;
  categories: DiscoveryAggregate[];
  cities: DiscoveryAggregate[];
}) {
  const visibleCategories = categories.filter((item) => item.slug && item.name).slice(0, 8);
  const visibleCities = cities.filter((item) => item.city).slice(0, 6);

  return (
    <section className={publicTheme.bg}>
      <div className="mx-auto max-w-6xl space-y-3 px-4 pb-3 pt-3 sm:px-6 sm:pt-4 lg:px-8">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-[#251F1A] sm:text-3xl">
              Temukan produk UMKM lokal.
            </h1>
            <p className="mt-1 max-w-2xl text-sm leading-6 text-[#6F6256]">
              Cari produk, bandingkan toko, lalu checkout langsung di storefront resmi masing-masing UMKM.
            </p>
          </div>
          <Link className="hidden text-sm font-semibold text-[#B96E45] hover:text-[#7C3F25] sm:inline-flex" href="/products">
            Lihat semua produk
          </Link>
        </div>

        <form action="/" className="flex gap-2 rounded-2xl border border-[#E3D2BC] bg-[#FFFDF8] p-2 shadow-sm">
          <input type="hidden" name="category" value={category ?? ""} />
          <input type="hidden" name="city" value={city ?? ""} />
          <input type="hidden" name="price_min" value={priceMin ?? ""} />
          <input type="hidden" name="price_max" value={priceMax ?? ""} />
          <input
            className="h-10 min-w-0 flex-1 rounded-xl bg-transparent px-2 text-sm text-[#251F1A] outline-none placeholder:text-[#9b8d7b]"
            defaultValue={query}
            name="q"
            placeholder="Cari produk, toko, kategori, atau kota..."
          />
          <Button className="h-10 shrink-0 bg-[#251F1A] px-4 hover:bg-[#16110E]" type="submit">
            Cari
          </Button>
        </form>

        <div className="flex items-center gap-2">
          <div className="flex min-w-0 flex-1 gap-2 overflow-x-auto pb-1">
            <FilterChip active={!query && !category && !city && !priceMin && !priceMax} href="/">
              Semua
            </FilterChip>
            <FilterChip active={query === "bouquet"} href="/?q=bouquet">
              Bouquet
            </FilterChip>
            <FilterChip active={query === "hampers"} href="/?q=hampers">
              Hampers
            </FilterChip>
            <FilterChip active={query === "makanan"} href="/?q=makanan">
              Makanan
            </FilterChip>
            {visibleCategories.map((item) => (
              <FilterChip
                active={category === item.slug}
                href={`/?category=${encodeURIComponent(item.slug ?? "")}`}
                key={`category-${item.slug}`}
              >
                {item.name} ({item.count})
              </FilterChip>
            ))}
            {visibleCities.map((item) => (
              <FilterChip
                active={city === item.city}
                href={`/?city=${encodeURIComponent(item.city ?? "")}`}
                key={`city-${item.city}`}
              >
                {item.city} ({item.count})
              </FilterChip>
            ))}
          </div>

          <details className="group relative shrink-0">
            <summary className="inline-flex h-9 cursor-pointer list-none items-center rounded-full border border-[#E3D2BC] bg-[#FFFDF8] px-3 text-sm font-semibold text-[#6F6256] shadow-sm marker:hidden hover:text-[#251F1A]">
              Filter
              <span className="ml-2 text-xs text-[#B96E45] group-open:hidden">Buka</span>
              <span className="ml-2 hidden text-xs text-[#B96E45] group-open:inline">Tutup</span>
            </summary>
            <div className="absolute right-0 top-11 z-30 w-[min(92vw,680px)] rounded-3xl border border-[#E3D2BC] bg-[#FFFDF8] p-3 shadow-[0_18px_50px_rgba(80,57,34,0.16)]">
              <form action="/" className="grid gap-2 sm:grid-cols-2 lg:grid-cols-[1fr_1fr_130px_130px_auto]">
                <input type="hidden" name="q" value={query ?? ""} />
                <select
                  className="h-10 rounded-xl border border-[#E3D2BC] bg-white px-3 text-sm text-[#251F1A] outline-none focus:border-[#B96E45] focus:ring-2 focus:ring-[#E8D2AA]"
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
                <select
                  className="h-10 rounded-xl border border-[#E3D2BC] bg-white px-3 text-sm text-[#251F1A] outline-none focus:border-[#B96E45] focus:ring-2 focus:ring-[#E8D2AA]"
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
                <input
                  className="h-10 rounded-xl border border-[#E3D2BC] bg-white px-3 text-sm text-[#251F1A] outline-none placeholder:text-[#9b8d7b] focus:border-[#B96E45] focus:ring-2 focus:ring-[#E8D2AA]"
                  defaultValue={priceMin}
                  min={0}
                  name="price_min"
                  placeholder="Harga min"
                  type="number"
                />
                <input
                  className="h-10 rounded-xl border border-[#E3D2BC] bg-white px-3 text-sm text-[#251F1A] outline-none placeholder:text-[#9b8d7b] focus:border-[#B96E45] focus:ring-2 focus:ring-[#E8D2AA]"
                  defaultValue={priceMax}
                  min={0}
                  name="price_max"
                  placeholder="Harga max"
                  type="number"
                />
                <Button className="h-10 w-full bg-[#B96E45] hover:bg-[#7C3F25]" type="submit">
                  Terapkan
                </Button>
              </form>
              {query || category || city || priceMin || priceMax ? (
                <Link className="mt-2 inline-flex text-sm font-semibold text-[#B96E45] hover:text-[#7C3F25]" href="/">
                  Reset filter
                </Link>
              ) : null}
            </div>
          </details>
        </div>
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
    <section className="space-y-3">
      <PublicSectionHeader title={title} description={description} href={href} />
      {stores.length === 0 ? (
        <EmptyState
          title="Toko pertama sedang disiapkan"
          description="Coba cari produk atau jelajahi kategori yang tersedia. Storefront publik akan tampil saat toko sudah dipublikasikan."
        />
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
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
    <section className="space-y-3">
      <PublicSectionHeader title={title} description={description} href={href} />
      {products.length === 0 ? (
        <EmptyState
          title="Produk baru akan muncul di sini"
          description="Produk baru akan muncul di sini saat toko menambahkan katalog."
        />
      ) : (
        <div className="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4">
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
    <section className="space-y-2">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-base font-bold text-[#251F1A]">{title}</h2>
        <span className="text-xs font-semibold text-[#B96E45]">Jelajah cepat</span>
      </div>
      <div className="flex gap-2 overflow-x-auto pb-1">
        {items.map((item, index) => {
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
              className={`inline-flex min-h-10 shrink-0 items-center rounded-2xl border px-3.5 py-2 text-sm font-semibold shadow-sm transition hover:-translate-y-0.5 ${chipTone(index)}`}
            >
              {label} <span className="ml-1 opacity-70">({item.count})</span>
            </Link>
          );
        })}
      </div>
    </section>
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
    <div className="flex flex-col gap-3 rounded-3xl border border-[#E3D2BC] bg-[#FFFDF8] p-4 sm:flex-row sm:items-center sm:justify-between">
      <LinkButton href={buildHref(basePath, { ...searchParams, cursor: undefined })}>Reset halaman</LinkButton>
      <p className="text-sm text-[#6F6256]">Menampilkan maksimal {pagination.limit} item.</p>
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
      className="grid gap-3 rounded-[28px] border border-[#E3D2BC] bg-[#FFFDF8] p-4 shadow-[0_14px_40px_rgba(89,63,38,0.07)] lg:grid-cols-[1.2fr_180px_180px_auto]"
      action={action}
    >
      <input
        className="h-11 rounded-xl border border-[#E3D2BC] bg-white px-3 text-sm text-[#251F1A] outline-none placeholder:text-[#9B8D7B] focus:border-[#B96E45] focus:ring-2 focus:ring-[#E8D2AA]"
        defaultValue={query}
        name="q"
        placeholder="Cari nama, produk, atau toko..."
      />
      <select
        className="h-11 rounded-xl border border-[#E3D2BC] bg-white px-3 text-sm text-[#251F1A] outline-none focus:border-[#B96E45] focus:ring-2 focus:ring-[#E8D2AA]"
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
        className="h-11 rounded-xl border border-[#E3D2BC] bg-white px-3 text-sm text-[#251F1A] outline-none focus:border-[#B96E45] focus:ring-2 focus:ring-[#E8D2AA]"
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
            className="h-11 rounded-xl border border-[#E3D2BC] bg-white px-3 text-sm text-[#251F1A] outline-none placeholder:text-[#9B8D7B] focus:border-[#B96E45] focus:ring-2 focus:ring-[#E8D2AA]"
            defaultValue={priceMin}
            min={0}
            name="price_min"
            placeholder="Harga min"
            type="number"
          />
          <input
            className="h-11 rounded-xl border border-[#E3D2BC] bg-white px-3 text-sm text-[#251F1A] outline-none placeholder:text-[#9B8D7B] focus:border-[#B96E45] focus:ring-2 focus:ring-[#E8D2AA]"
            defaultValue={priceMax}
            min={0}
            name="price_max"
            placeholder="Harga max"
            type="number"
          />
        </div>
      ) : null}
      <Button className="h-11 w-full bg-[#251F1A] hover:bg-[#16110E] lg:w-auto" type="submit">
        Terapkan
      </Button>
    </form>
  );
}

function QuickChip({ href, children }: { href: string; children: string }) {
  return (
    <Link
      href={href}
      className="inline-flex min-h-8 shrink-0 items-center rounded-full border border-[#E3D2BC] bg-white px-3 text-xs font-semibold text-[#6F6256] transition hover:border-[#B96E45] hover:text-[#7C3F25]"
    >
      {children}
    </Link>
  );
}

function FilterChip({
  href,
  children,
  active = false
}: {
  href: string;
  children: ReactNode;
  active?: boolean;
}) {
  return (
    <Link
      href={href}
      className={cn(
        "inline-flex min-h-9 shrink-0 items-center rounded-full border px-3.5 text-sm font-semibold transition",
        active
          ? "border-[#251F1A] bg-[#251F1A] text-[#FFFDF8]"
          : "border-[#E3D2BC] bg-white text-[#6F6256] hover:border-[#B96E45] hover:text-[#7C3F25]"
      )}
    >
      {children}
    </Link>
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
          ? "inline-flex h-11 items-center justify-center rounded-xl bg-[#251F1A] px-4 text-sm font-semibold text-[#FFFDF8] transition hover:bg-[#16110E]"
          : "inline-flex h-11 items-center justify-center rounded-xl border border-[#E3D2BC] bg-white px-4 text-sm font-semibold text-[#251F1A] transition hover:bg-[#F1E7D8]"
      }
    >
      {children}
    </Link>
  );
}

function chipTone(index: number) {
  const tones = [
    "border-[#E3D2BC] bg-[#FFFDF8] text-[#6F6256]",
    "border-[#d7dec5] bg-[#f4f8ed] text-[#4f613d]",
    "border-[#e8d3ae] bg-[#FFF5DE] text-[#7A4D1D]",
    "border-[#cfe0dc] bg-[#eff7f4] text-[#315f58]"
  ];
  return tones[index % tones.length];
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
