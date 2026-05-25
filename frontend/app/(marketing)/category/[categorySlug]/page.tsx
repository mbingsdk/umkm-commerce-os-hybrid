import type { Metadata } from "next";
import type { ReactNode } from "react";
import Link from "next/link";
import { EmptyState } from "@/components/feedback/empty-state";
import {
  getDiscoveryHome,
  listDiscoveryProducts,
  listDiscoveryStores
} from "@/features/discovery/api/discovery.api";
import {
  ProductSection,
  StoreSection
} from "@/features/discovery/components/discovery-sections";
import { publicPageMetadata } from "@/lib/seo/metadata";

type CategoryDiscoveryPageProps = {
  params: Promise<{ categorySlug: string }>;
};

export async function generateMetadata({ params }: CategoryDiscoveryPageProps): Promise<Metadata> {
  const { categorySlug } = await params;
  const categoryName = await getCategoryName(categorySlug);

  return publicPageMetadata({
    title: `${categoryName} - Produk dan Toko UMKM`,
    description: `Jelajahi produk dan toko UMKM dalam kategori ${categoryName}. Semua hasil mengarah ke storefront tenant.`,
    path: `/category/${categorySlug}`
  });
}

export default async function CategoryDiscoveryPage({ params }: CategoryDiscoveryPageProps) {
  const { categorySlug } = await params;

  const [home, products, stores] = await Promise.all([
    getDiscoveryHome(),
    listDiscoveryProducts({ category: categorySlug, limit: 12 }),
    listDiscoveryStores({ category: categorySlug, limit: 6 })
  ]);

  const category = home.popularCategories.find((item) => item.slug === categorySlug);
  const categoryName = category?.name ?? titleFromSlug(categorySlug);
  const hasResults = products.items.length > 0 || stores.items.length > 0;

  return (
    <main className="min-h-screen bg-neutral-50">
      <section className="border-b border-neutral-200 bg-gradient-to-br from-primary-50 via-white to-neutral-50">
        <div className="mx-auto space-y-5 px-4 py-10 sm:px-6 lg:max-w-6xl lg:px-8 lg:py-14">
          <p className="text-sm font-semibold text-primary-700">Kategori discovery</p>
          <div className="max-w-3xl">
            <h1 className="text-3xl font-bold tracking-tight text-neutral-950 sm:text-5xl">{categoryName}</h1>
            <p className="mt-4 text-sm leading-7 text-neutral-600 sm:text-base">
              Temukan produk dan toko UMKM publik dalam kategori {categoryName}. Klik produk atau toko untuk
              masuk ke storefront tenant masing-masing.
            </p>
          </div>

          <div className="flex flex-col gap-3 sm:flex-row">
            <LinkButton href={`/products?category=${encodeURIComponent(categorySlug)}`}>Lihat produk</LinkButton>
            <LinkButton href={`/stores?category=${encodeURIComponent(categorySlug)}`}>Lihat toko</LinkButton>
            <LinkButton href={`/search?category=${encodeURIComponent(categorySlug)}`}>Cari di kategori ini</LinkButton>
          </div>
        </div>
      </section>

      <section className="mx-auto space-y-10 px-4 py-10 sm:px-6 lg:max-w-6xl lg:px-8">
        {!hasResults ? (
          <EmptyState
            title="Belum ada hasil di kategori ini"
            description="Produk atau toko yang memenuhi aturan discovery akan muncul di sini setelah tersedia."
            action={<LinkButton href="/products">Jelajahi semua produk</LinkButton>}
          />
        ) : null}

        <ProductSection
          title={`Produk ${categoryName}`}
          description="Produk aktif dan discoverable dari kategori ini."
          products={products.items}
          href={`/products?category=${encodeURIComponent(categorySlug)}`}
        />

        <StoreSection
          title={`Toko terkait ${categoryName}`}
          description="Toko publik yang memiliki produk atau kategori terkait."
          stores={stores.items}
          href={`/stores?category=${encodeURIComponent(categorySlug)}`}
        />
      </section>
    </main>
  );
}

async function getCategoryName(categorySlug: string) {
  try {
    const home = await getDiscoveryHome();
    return home.popularCategories.find((item) => item.slug === categorySlug)?.name ?? titleFromSlug(categorySlug);
  } catch {
    return titleFromSlug(categorySlug);
  }
}

function titleFromSlug(value: string) {
  return decodeURIComponent(value)
    .split("-")
    .filter(Boolean)
    .map((word) => word.slice(0, 1).toUpperCase() + word.slice(1))
    .join(" ");
}

function LinkButton({ href, children }: { href: string; children: ReactNode }) {
  return (
    <Link
      className="inline-flex min-h-11 items-center justify-center rounded-xl border border-neutral-300 bg-white px-4 py-2 text-sm font-semibold text-neutral-900 transition hover:bg-neutral-50"
      href={href}
    >
      {children}
    </Link>
  );
}
