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

type CategoryPageProps = {
  params: Promise<{
    storeSlug: string;
    categorySlug: string;
  }>;
  searchParams: Promise<{
    q?: string | string[];
    cursor?: string | string[];
  }>;
};

export async function generateMetadata({ params }: Pick<CategoryPageProps, "params">): Promise<Metadata> {
  const { storeSlug, categorySlug } = await params;

  try {
    const [store, categories] = await Promise.all([
      getPublicStoreBySlug(storeSlug),
      listPublicCategories(storeSlug)
    ]);
    const category = categories.find((item) => item.slug === categorySlug);

    if (!category) {
      return {
        title: "Kategori tidak ditemukan"
      };
    }

    const title = `${category.name} - ${store.name}`;
    const description =
      category.description ??
      `Belanja produk kategori ${category.name} dari ${store.name}${store.city ? ` di ${store.city}` : ""}.`;
    const image = toAbsoluteURL(category.imageUrl ?? store.bannerUrl ?? store.logoUrl);
    const canonicalURL = `${siteURL}/s/${store.slug}/categories/${category.slug}`;

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
      title: "Kategori tidak ditemukan"
    };
  }
}

export default async function CategoryPage({ params, searchParams }: CategoryPageProps) {
  const [{ storeSlug, categorySlug }, rawSearchParams] = await Promise.all([params, searchParams]);
  const query = firstParam(rawSearchParams.q);
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

  const category = categories.find((item) => item.slug === categorySlug);

  if (!category) {
    notFound();
  }

  return (
    <StorefrontProductListingView
      activeCategory={category}
      categories={categories}
      currentPath={`/s/${store.slug}/categories/${category.slug}`}
      description={
        category.description ??
        `Produk dalam kategori ${category.name} dari ${store.name}. Semua harga dan stok akhir tetap mengikuti data toko.`
      }
      products={products}
      query={query}
      store={store}
      title={category.name}
    />
  );
}

function firstParam(value?: string | string[]) {
  return Array.isArray(value) ? value[0] : value;
}
