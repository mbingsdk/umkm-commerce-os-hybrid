import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { formatRupiah } from "@/lib/format/money";
import { getPublicProductDetail, isPublicNotFoundError } from "@/features/storefront/api/storefront.api";
import { AddToCartButton } from "@/features/storefront/components/add-to-cart-button";
import { ProductGallery } from "@/features/storefront/components/product-gallery";
import { ShareLinkButton } from "@/features/storefront/components/share-link-button";
import { StockBadge } from "@/features/storefront/components/stock-badge";
import { buildProductJsonLd, getSiteURL, serializeJsonLd, toAbsoluteURL } from "@/features/storefront/seo";
import type { PublicProductDetail } from "@/features/storefront/types";

const siteURL = getSiteURL();

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
    const image = toAbsoluteURL(product.images[0]?.url);
    const canonicalURL = `${siteURL}/s/${product.store.slug}/products/${product.slug}`;

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

  const isOutOfStock = product.stock.stockStatus === "out_of_stock";
  const productJsonLd = buildProductJsonLd(product);

  return (
    <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
      <script
        dangerouslySetInnerHTML={{ __html: serializeJsonLd(productJsonLd) }}
        type="application/ld+json"
      />

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
              <h1 className="text-2xl font-bold tracking-tight text-neutral-950 sm:text-3xl">{product.name}</h1>
              <p className="mt-3 text-2xl font-bold text-primary-700">{formatRupiah(product.price)}</p>
              {product.compareAtPrice ? (
                <p className="mt-1 text-sm text-neutral-400 line-through">{formatRupiah(product.compareAtPrice)}</p>
              ) : null}
            </div>
            <StockBadge status={product.stock.stockStatus} />
          </div>

          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-soft">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <div>
                <h2 className="text-lg font-semibold text-neutral-950">Belanja produk ini</h2>
                <p className="mt-1 text-sm leading-6 text-neutral-600">
                  Total akhir akan dihitung ulang oleh toko saat checkout.
                </p>
              </div>
              <AddToCartButton
                disabled={isOutOfStock}
                item={{
                  productId: product.id,
                  storeSlug: product.store.slug,
                  name: product.name,
                  slug: product.slug,
                  imageUrl: product.images[0]?.url,
                  displayPrice: product.price,
                  quantity: 1
                }}
                size="lg"
              />
            </div>
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
            <p className="mt-4 text-sm leading-6 text-neutral-600">
              {isOutOfStock
                ? "Stok sedang habis. Lihat kontak toko untuk menanyakan restok via WhatsApp bila tersedia."
                : "Produk tersedia. Lihat kontak toko untuk tanya detail dan pemesanan via WhatsApp bila tersedia."}
            </p>
            <div className="mt-4 flex flex-wrap gap-2">
              <Link
                className="inline-flex h-10 items-center justify-center rounded-xl bg-primary-600 px-4 text-sm font-semibold text-white transition hover:bg-primary-700"
                href={`/s/${product.store.slug}`}
              >
                Lihat kontak toko
              </Link>
              <ShareLinkButton label="Bagikan produk" />
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
