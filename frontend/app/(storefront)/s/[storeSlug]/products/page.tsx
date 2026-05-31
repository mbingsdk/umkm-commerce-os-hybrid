import type { Metadata } from "next";
import { notFound } from "next/navigation";
import {
  getPublicStoreBySlug,
  isPublicNotFoundError,
  listPublicCategories,
  listPublicProducts
} from "@/features/storefront/api/storefront.api";
import { StorefrontProductListingView } from "@/features/storefront/components/storefront-product-listing-view";
import { getSiteURL, toAbsoluteURL } from "@/features/storefront/seo";
import type {
  PublicCategory,
  PublicProductListResult,
  PublicStore
} from "@/features/storefront/types";

const siteURL = getSiteURL();

type ProductListingPageProps = {
  params: Promise<{ storeSlug: string }>;
  searchParams: Promise<{
    q?: string | string[];
    category?: string | string[];
    cursor?: string | string[];
  }>;
};

export async function generateMetadata({ params }: Pick<ProductListingPageProps, "params">): Promise<Metadata> {
  const { storeSlug } = await params;

  try {
    const store = await getPublicStoreBySlug(storeSlug);
    const title = `Produk ${store.name}`;
    const description = `Lihat katalog produk aktif dari ${store.name}${store.city ? ` di ${store.city}` : ""}.`;
    const image = toAbsoluteURL(store.bannerUrl ?? store.logoUrl);
    const canonicalURL = `${siteURL}/s/${store.slug}/products`;

    return {
      title,
      description,
      alternates: {
        canonical: canonicalURL
      },
      openGraph: {
        title,
        description,
        locale: "id_ID",
        type: "website",
        url: canonicalURL,
        images: image ? [image] : undefined
      }
    };
  } catch {
    return {
      title: "Produk toko tidak ditemukan"
    };
  }
}

export default async function ProductListingPage({ params, searchParams }: ProductListingPageProps) {
  const [{ storeSlug }, rawSearchParams] = await Promise.all([params, searchParams]);
  const query = firstParam(rawSearchParams.q);
  const categorySlug = firstParam(rawSearchParams.category);
  const cursor = firstParam(rawSearchParams.cursor);

  let store: PublicStore;
  let categories: PublicCategory[];
  let products: PublicProductListResult;

  try {
    [store, categories, products] = await Promise.all([
      getPublicStoreBySlug(storeSlug),
      listPublicCategories(storeSlug),
      listPublicProducts(storeSlug, {
        query,
        categorySlug,
        cursor,
        limit: 24
      })
    ]);
  } catch (error) {
    if (isPublicNotFoundError(error)) {
      notFound();
    }
    throw error;
  }

  const activeCategory = categories.find((category) => category.slug === categorySlug);

  return (
    <StorefrontProductListingView
      activeCategory={activeCategory}
      categories={categories}
      currentPath={`/s/${store.slug}/products`}
      description={`Jelajahi produk aktif dari ${store.name}. Gunakan pencarian dan kategori untuk menemukan produk yang kamu butuhkan.`}
      products={products}
      query={query}
      store={store}
      title="Semua Produk"
    />
  );
}

function firstParam(value?: string | string[]) {
  return Array.isArray(value) ? value[0] : value;
}
