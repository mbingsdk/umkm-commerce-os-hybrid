import Link from "next/link";
import { EmptyState } from "@/components/feedback/empty-state";
import { Input } from "@/components/ui/input";
import { ProductCard } from "@/features/storefront/components/product-card";
import type {
  PublicCategory,
  PublicProductListResult,
  PublicStore
} from "@/features/storefront/types";

type StorefrontProductListingViewProps = {
  store: PublicStore;
  categories: PublicCategory[];
  products: PublicProductListResult;
  title: string;
  description: string;
  query?: string;
  currentPath: string;
  activeCategory?: PublicCategory;
};

export function StorefrontProductListingView({
  store,
  categories,
  products,
  title,
  description,
  query,
  currentPath,
  activeCategory
}: StorefrontProductListingViewProps) {
  const hasProducts = products.items.length > 0;

  return (
    <main>
      <section className="bg-[#F8F1E7]">
        <div className="mx-auto max-w-6xl px-4 py-6 sm:px-6 sm:py-8 lg:px-8">
          <div className="space-y-4 rounded-[32px] border border-[#E3D2BC] bg-[#FFFDF8] p-5 shadow-[0_18px_50px_rgba(89,63,38,0.08)] sm:p-7">
            <Link className="text-sm font-semibold text-[#B96E45] hover:text-[#7C3F25]" href={`/s/${store.slug}`}>
              ? Kembali ke toko
            </Link>
            <div className="max-w-3xl space-y-2">
              <p className="text-sm font-medium text-[#6F6256]">
                {store.name}
                {store.city ? ` ? ${store.city}` : ""}
              </p>
              <h1 className="text-2xl font-bold tracking-tight text-[#251F1A] sm:text-4xl">{title}</h1>
              <p className="text-sm leading-7 text-[#6F6256] sm:text-base">{description}</p>
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-6xl space-y-6 px-4 py-6 sm:px-6 sm:py-8 lg:px-8">
        <div className="space-y-4 rounded-[28px] border border-[#E3D2BC] bg-white/85 p-5 shadow-[0_12px_35px_rgba(89,63,38,0.06)]">
          <div>
            <h2 className="text-lg font-semibold text-[#251F1A]">Cari produk</h2>
            <p className="mt-1 text-sm text-[#6F6256]">
              Cari produk aktif dari toko ini berdasarkan nama produk.
            </p>
          </div>

          <form action={currentPath} className="flex flex-col gap-3 sm:flex-row" method="get">
            <Input defaultValue={query} name="q" placeholder="Cari nama produk..." />
            <button className="h-10 rounded-xl bg-[#251F1A] px-4 text-sm font-semibold text-[#FFFDF8] transition hover:bg-[#16110E]">
              Cari
            </button>
          </form>
        </div>

        <div className="space-y-4">
          {categories.length === 0 ? (
            <EmptyState
              title="Kategori belum tersedia"
              description="Toko ini belum menambahkan kategori publik. Semua produk tetap bisa dilihat dari daftar produk."
            />
          ) : (
            <div className="flex gap-2 overflow-x-auto pb-1 sm:flex-wrap sm:overflow-visible">
              <CategoryLink
                active={!activeCategory}
                href={buildListingHref(`/s/${store.slug}/products`, { q: query })}
                label="Semua produk"
              />
              {categories.map((category) => (
                <CategoryLink
                  key={category.id}
                  active={category.slug === activeCategory?.slug}
                  href={buildListingHref(`/s/${store.slug}/categories/${category.slug}`, { q: query })}
                  label={category.name}
                />
              ))}
            </div>
          )}

          {hasProducts ? (
            <>
              <div className="grid grid-cols-2 gap-3 sm:gap-4 md:grid-cols-3 xl:grid-cols-4">
                {products.items.map((product) => (
                  <ProductCard
                    key={product.id}
                    categoryName={activeCategory?.name}
                    product={product}
                    storeSlug={store.slug}
                  />
                ))}
              </div>

              {products.pagination?.hasMore && products.pagination.nextCursor ? (
                <div className="flex justify-center pt-2">
                  <Link
                    className="inline-flex h-10 items-center justify-center rounded-xl border border-[#E3D2BC] bg-[#FFFDF8] px-4 text-sm font-semibold text-[#7C3F25] transition hover:bg-[#F1E7D8]"
                    href={buildListingHref(currentPath, {
                      q: query,
                      cursor: products.pagination.nextCursor
                    })}
                  >
                    Lihat produk berikutnya
                  </Link>
                </div>
              ) : null}
            </>
          ) : (
            <EmptyState
              title="Produk belum ditemukan"
              description={
                query || activeCategory
                  ? "Coba ubah kata kunci atau pilih kategori lain."
                  : "Toko ini belum menampilkan produk. Hubungi toko melalui WhatsApp untuk info terbaru."
              }
              action={
                query || activeCategory ? (
                  <Link
                    className="inline-flex h-10 items-center justify-center rounded-xl bg-[#251F1A] px-4 text-sm font-semibold text-[#FFFDF8] transition hover:bg-[#16110E]"
                    href={`/s/${store.slug}/products`}
                  >
                    Lihat semua produk
                  </Link>
                ) : null
              }
            />
          )}
        </div>
      </section>
    </main>
  );
}

function CategoryLink({ active, href, label }: { active: boolean; href: string; label: string }) {
  return (
    <Link
      className={
        active
          ? "shrink-0 rounded-full bg-[#251F1A] px-4 py-2 text-sm font-semibold text-[#FFFDF8]"
          : "shrink-0 rounded-full border border-[#E3D2BC] bg-[#FFFDF8] px-4 py-2 text-sm font-semibold text-[#6F6256] transition hover:border-[#B96E45] hover:text-[#7C3F25]"
      }
      href={href}
    >
      {label}
    </Link>
  );
}

function buildListingHref(path: string, params: { q?: string; cursor?: string }) {
  const searchParams = new URLSearchParams();

  if (params.q) {
    searchParams.set("q", params.q);
  }
  if (params.cursor) {
    searchParams.set("cursor", params.cursor);
  }

  const suffix = searchParams.size > 0 ? `?${searchParams.toString()}` : "";
  return `${path}${suffix}`;
}
