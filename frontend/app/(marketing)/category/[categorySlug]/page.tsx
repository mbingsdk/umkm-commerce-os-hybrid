import type { Metadata } from "next";
import { EmptyState } from "@/components/feedback/empty-state";
import { PublicLinkButton, PublicPageIntro } from "@/components/public/public-ui";
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
    <main className="min-h-screen bg-[#F8F1E7]">
      <section>
        <div className="mx-auto space-y-5 px-4 py-8 sm:px-6 lg:max-w-6xl lg:px-8">
          <PublicPageIntro
            compact
            eyebrow="Kategori discovery"
            title={categoryName}
            description={`Temukan produk dan toko UMKM publik dalam kategori ${categoryName}. Klik produk atau toko untuk masuk ke storefront resmi masing-masing.`}
          >
            <div className="flex flex-col gap-3 sm:flex-row">
              <PublicLinkButton href={`/products?category=${encodeURIComponent(categorySlug)}`} variant="outline">Lihat produk</PublicLinkButton>
              <PublicLinkButton href={`/stores?category=${encodeURIComponent(categorySlug)}`} variant="outline">Lihat toko</PublicLinkButton>
              <PublicLinkButton href={`/search?category=${encodeURIComponent(categorySlug)}`} variant="outline">Cari di kategori ini</PublicLinkButton>
            </div>
          </PublicPageIntro>
        </div>
      </section>

      <section className="mx-auto space-y-10 px-4 pb-10 sm:px-6 lg:max-w-6xl lg:px-8">
        {!hasResults ? (
          <EmptyState
            title="Belum ada hasil di kategori ini"
            description="Produk atau toko yang memenuhi aturan discovery akan muncul di sini setelah tersedia."
            action={<PublicLinkButton href="/products" variant="outline">Jelajahi semua produk</PublicLinkButton>}
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
