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
        <Link href={`/s/${product.store.slug}`} className="text-sm font-semibold text-[#7a4f2f] hover:text-[#3b2f24]">
          ? Kembali ke toko
        </Link>
      </div>

      <section className="grid gap-8 lg:grid-cols-[minmax(0,1.05fr)_minmax(360px,0.95fr)]">
        <div className="rounded-[32px] border border-[#eadfce] bg-[#fffaf2] p-3 shadow-[0_18px_50px_rgba(89,63,38,0.08)]">
          <ProductGallery product={product} />
        </div>

        <div className="space-y-5">
          <div className="rounded-[32px] border border-[#eadfce] bg-white/85 p-5 shadow-[0_12px_35px_rgba(89,63,38,0.06)] sm:p-6">
            <div className="space-y-3">
              {product.category ? <p className="text-sm font-semibold text-[#7a4f2f]">{product.category.name}</p> : null}
              <div>
                <h1 className="text-2xl font-bold tracking-tight text-[#241c16] sm:text-3xl">{product.name}</h1>
                <p className="mt-3 text-2xl font-bold text-[#7a4f2f]">{formatRupiah(product.price)}</p>
                {product.compareAtPrice ? (
                  <p className="mt-1 text-sm text-neutral-400 line-through">{formatRupiah(product.compareAtPrice)}</p>
                ) : null}
              </div>
              <StockBadge status={product.stock.stockStatus} />
            </div>
          </div>

          <div className="rounded-[28px] border border-[#eadfce] bg-[#fffaf2] p-5 shadow-[0_12px_35px_rgba(89,63,38,0.06)]">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
              <div>
                <h2 className="text-lg font-semibold text-[#241c16]">Belanja produk ini</h2>
                <p className="mt-1 text-sm leading-6 text-[#665746]">
                  Harga akhir dan stok akan dicek ulang saat checkout agar pesanan tetap akurat.
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

          <div className="rounded-[28px] border border-[#eadfce] bg-white/85 p-5 shadow-[0_12px_35px_rgba(89,63,38,0.06)]">
            <h2 className="text-lg font-semibold text-[#241c16]">Deskripsi produk</h2>
            <p className="mt-3 whitespace-pre-line text-sm leading-7 text-[#665746]">
              {product.description ?? "Deskripsi produk belum tersedia. Hubungi toko untuk menanyakan detail produk ini."}
            </p>
          </div>

          <div className="rounded-[28px] border border-[#eadfce] bg-[#fffaf2] p-5 shadow-[0_12px_35px_rgba(89,63,38,0.06)]">
            <h2 className="text-lg font-semibold text-[#241c16]">Informasi toko</h2>
            <p className="mt-2 text-sm font-medium text-[#665746]">
              {product.store.name}
              {product.store.city ? ` ? ${product.store.city}` : ""}
            </p>
            <p className="mt-4 text-sm leading-6 text-[#665746]">
              {isOutOfStock
                ? "Stok sedang habis. Kamu bisa menghubungi toko untuk menanyakan restok atau produk pengganti."
                : "Produk tersedia. Jika perlu detail tambahan, hubungi toko sebelum checkout."}
            </p>
            <div className="mt-4 flex flex-wrap gap-2">
              <Link
                className="inline-flex h-10 items-center justify-center rounded-xl bg-[#2f2923] px-4 text-sm font-semibold text-[#fffaf2] transition hover:bg-[#1f1a16]"
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
