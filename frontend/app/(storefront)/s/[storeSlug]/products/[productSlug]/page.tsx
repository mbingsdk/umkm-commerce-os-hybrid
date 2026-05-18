import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { formatRupiah } from "@/lib/format/money";
import { getPublicProductDetail, isPublicNotFoundError } from "@/features/storefront/api/storefront.api";
import { ProductGallery } from "@/features/storefront/components/product-gallery";
import { StockBadge } from "@/features/storefront/components/stock-badge";
import type { PublicProductDetail } from "@/features/storefront/types";

const siteURL = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";

type ProductDetailPageProps = {
  params: Promise<{
    storeSlug: string;
    productSlug: string;
  }>;
};

export async function generateMetadata({ params }: ProductDetailPageProps): Promise<Metadata> {
  const { storeSlug, productSlug } = await params;

  try {
    const product = await getPublicProductDetail(storeSlug, productSlug);
    const title = product.seo?.title ?? `${product.name} - ${product.store.name}`;
    const description =
      product.seo?.description ??
      product.description ??
      `Beli ${product.name} dari ${product.store.name}${product.store.city ? ` di ${product.store.city}` : ""}.`;
    const image = product.images[0]?.url;

    return {
      title,
      description,
      alternates: {
        canonical: `${siteURL}/s/${product.store.slug}/products/${product.slug}`
      },
      openGraph: {
        title,
        description,
        images: image ? [image] : undefined
      }
    };
  } catch {
    return {
      title: "Produk tidak ditemukan"
    };
  }
}

export default async function ProductDetailPage({ params }: ProductDetailPageProps) {
  const { storeSlug, productSlug } = await params;
  let product: PublicProductDetail;

  try {
    product = await getPublicProductDetail(storeSlug, productSlug);
  } catch (error) {
    if (isPublicNotFoundError(error)) {
      notFound();
    }
    throw error;
  }

  return (
    <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-6">
        <Link href={`/s/${product.store.slug}`} className="text-sm font-semibold text-primary-700 hover:text-primary-800">
          Kembali ke toko
        </Link>
      </div>

      <section className="grid gap-8 lg:grid-cols-2">
        <ProductGallery product={product} />

        <div className="space-y-6">
          <div className="space-y-3">
            {product.category ? <p className="text-sm font-medium text-primary-700">{product.category.name}</p> : null}
            <div>
              <h1 className="text-3xl font-bold tracking-tight text-neutral-950">{product.name}</h1>
              <p className="mt-3 text-2xl font-bold text-primary-700">{formatRupiah(product.price)}</p>
              {product.compareAtPrice ? (
                <p className="mt-1 text-sm text-neutral-400 line-through">{formatRupiah(product.compareAtPrice)}</p>
              ) : null}
            </div>
            <StockBadge status={product.stock.stockStatus} />
          </div>

          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-soft">
            <h2 className="text-lg font-semibold text-neutral-950">Deskripsi produk</h2>
            <p className="mt-3 whitespace-pre-line text-sm leading-7 text-neutral-600">
              {product.description ?? "Deskripsi produk belum tersedia."}
            </p>
          </div>

          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-soft">
            <h2 className="text-lg font-semibold text-neutral-950">Informasi toko</h2>
            <p className="mt-2 text-sm text-neutral-600">
              {product.store.name}
              {product.store.city ? ` - ${product.store.city}` : ""}
            </p>
            <button
              className="mt-4 h-11 rounded-xl border border-neutral-300 bg-neutral-50 px-4 text-sm font-semibold text-neutral-500"
              disabled
              type="button"
            >
              Pesan via WhatsApp - segera hadir
            </button>
          </div>
        </div>
      </section>
    </main>
  );
}
