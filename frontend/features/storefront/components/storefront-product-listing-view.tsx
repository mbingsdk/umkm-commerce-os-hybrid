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
        <div className="mx-auto max-w-[1500px] px-4 py-4 sm:px-6 sm:py-5 lg:px-8">
          <div className="space-y-3 rounded-[28px] border border-[#E3D2BC] bg-[#FFFDF8] p-4 shadow-[0_14px_40px_rgba(89,63,38,0.07)] sm:p-5">
            <Link className="text-sm font-semibold text-[#B96E45] hover:text-[#7C3F25]" href={`/s/${store.slug}`}>
              Kembali ke toko
            </Link>
            <div className="max-w-3xl space-y-2">
              <p className="text-sm font-medium text-[#6F6256]">
                {store.name}
                {store.city ? ` - ${store.city}` : ""}
              </p>
              <h1 className="text-2xl font-bold tracking-tight text-[#251F1A] sm:text-3xl">{title}</h1>
              <p className="text-sm leading-6 text-[#6F6256]">{description}</p>
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-[1500px] space-y-4 px-4 py-4 sm:px-6 sm:py-6 lg:px-8">
        <div className="rounded-[20px] border border-[#E3D2BC] bg-[#FFFDF8] p-1.5 shadow-[0_8px_24px_rgba(89,63,38,0.05)]">
          <form action={currentPath} className="flex gap-2" method="get">
            <Input defaultValue={query} name="q" placeholder="Cari nama produk..." />
            <button className="h-10 shrink-0 rounded-xl bg-[#251F1A] px-3.5 text-sm font-semibold text-[#FFFDF8] transition hover:bg-[#16110E]">
              Cari
            </button>
          </form>
        </div>

        <div className="space-y-3">
          {categories.length === 0 ? (
            <EmptyState
              title="Kategori belum tersedia"
              description="Toko ini belum menambahkan kategori publik. Semua produk tetap bisa dilihat dari daftar produk."
            />
          ) : (
            <div className="flex gap-1.5 overflow-x-auto pb-1 sm:flex-wrap sm:overflow-visible">
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
              <div className="grid grid-cols-2 gap-2 sm:gap-2.5 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5">
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
                <div className="flex justify-center pt-1">
                  <Link
                    className="inline-flex h-9 items-center justify-center rounded-xl border border-[#E3D2BC] bg-[#FFFDF8] px-3.5 text-sm font-semibold text-[#7C3F25] transition hover:bg-[#F1E7D8]"
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
          ? "shrink-0 rounded-full bg-[#251F1A] px-3 py-1.5 text-xs font-semibold text-[#FFFDF8] sm:text-sm"
          : "shrink-0 rounded-full border border-[#E3D2BC] bg-[#FFFDF8] px-3 py-1.5 text-xs font-semibold text-[#6F6256] transition hover:border-[#B96E45] hover:text-[#7C3F25] sm:text-sm"
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
