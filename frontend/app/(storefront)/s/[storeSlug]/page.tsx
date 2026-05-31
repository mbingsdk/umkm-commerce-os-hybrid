import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { EmptyState } from "@/components/feedback/empty-state";
import { Input } from "@/components/ui/input";
import {
  getPublicStoreBySlug,
  isPublicNotFoundError,
  listPublicCategories,
  listPublicProducts
} from "@/features/storefront/api/storefront.api";
import { ProductCard } from "@/features/storefront/components/product-card";
import { SafeImage } from "@/features/storefront/components/safe-image";
import { ShareLinkButton } from "@/features/storefront/components/share-link-button";
import { buildStoreJsonLd, getSiteURL, serializeJsonLd, toAbsoluteURL } from "@/features/storefront/seo";
import type {
  PublicCategory,
  PublicProductListResult,
  PublicStore
} from "@/features/storefront/types";

const siteURL = getSiteURL();

type StorefrontPageProps = {
  params: Promise<{ storeSlug: string }>;
  searchParams: Promise<{
    q?: string | string[];
    category?: string | string[];
  }>;
};

export async function generateMetadata({ params }: Pick<StorefrontPageProps, "params">): Promise<Metadata> {
  const { storeSlug } = await params;

  try {
    const store = await getPublicStoreBySlug(storeSlug);
    const title = store.seo?.title ?? `${store.name} - Toko Online`;
    const description =
      store.seo?.description ??
      store.description ??
      `Belanja produk dari ${store.name}${store.city ? ` di ${store.city}` : ""}.`;
    const image = toAbsoluteURL(store.bannerUrl ?? store.logoUrl);
    const canonicalURL = `${siteURL}/s/${store.slug}`;

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
      title: "Toko tidak ditemukan"
    };
  }
}

export default async function StorefrontPage({ params, searchParams }: StorefrontPageProps) {
  const [{ storeSlug }, rawSearchParams] = await Promise.all([params, searchParams]);
  const query = firstParam(rawSearchParams.q);
  const categorySlug = firstParam(rawSearchParams.category);

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
        limit: 24
      })
    ]);
  } catch (error) {
    if (isPublicNotFoundError(error)) {
      notFound();
    }
    throw error;
  }

  const selectedCategory = categories.find((category) => category.slug === categorySlug);
  const whatsappHref = buildWhatsappHref(store.whatsapp);
  const storeJsonLd = buildStoreJsonLd(store);

  return (
    <main>
      <script
        dangerouslySetInnerHTML={{ __html: serializeJsonLd(storeJsonLd) }}
        type="application/ld+json"
      />

      <section className="bg-[#f7f1e8]">
        <div className="mx-auto max-w-6xl px-4 py-6 sm:px-6 sm:py-8 lg:px-8">
          <div className="overflow-hidden rounded-[32px] border border-[#d6b98e] bg-[#2f2923] text-[#fffaf2] shadow-[0_22px_70px_rgba(47,41,35,0.20)]">
            {store.bannerUrl ? (
              <SafeImage
                alt=""
                className="h-40 w-full object-cover opacity-75 sm:h-52"
                fallbackClassName="h-40 w-full bg-transparent sm:h-52"
                src={store.bannerUrl}
              />
            ) : null}
            <div className="grid gap-5 p-5 sm:p-8 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
              <div className="space-y-3">
                <p className="text-sm text-[#eadfce]">
                  {store.city ?? "Toko lokal"}
                  {store.province ? `, ${store.province}` : ""}
                </p>
                <h1 className="text-2xl font-bold tracking-tight sm:text-4xl">{store.name}</h1>
                <p className="max-w-2xl text-sm leading-7 text-[#f4eadb]">
                  {store.description ?? "Toko ini belum menambahkan deskripsi."}
                </p>
              </div>

              <div className="space-y-3">
                <p className="max-w-sm text-sm leading-6 text-[#f4eadb]">
                  {whatsappHref
                    ? "Hubungi toko via WhatsApp untuk tanya stok dan pemesanan."
                    : "Kontak WhatsApp belum tersedia. Lihat informasi toko untuk kontak lain."}
                </p>
                <div className="flex flex-wrap gap-2">
                  {whatsappHref ? (
                    <a
                      className="inline-flex h-9 items-center justify-center rounded-xl bg-[#fffaf2] px-4 text-sm font-semibold text-[#2f2923] transition hover:bg-[#f4eadb]"
                      href={whatsappHref}
                      rel="noopener noreferrer"
                      target="_blank"
                    >
                      Chat WhatsApp
                    </a>
                  ) : null}
                  <div className="[&_button]:border-white/30 [&_button]:bg-white/10 [&_button]:text-white [&_button:hover]:bg-white/20 [&_span]:text-[#f4eadb]">
                    <ShareLinkButton />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto grid max-w-6xl gap-6 px-4 py-6 sm:px-6 sm:py-8 lg:grid-cols-[minmax(0,1fr)_320px] lg:items-start lg:px-8">
        <div className="space-y-6">
          <div className="space-y-4 rounded-[28px] border border-[#eadfce] bg-[#fffaf2] p-5 shadow-[0_14px_40px_rgba(89,63,38,0.07)]">
            <div>
              <h2 className="text-lg font-bold text-[#241c16]">Cari produk</h2>
              <p className="mt-1 text-sm text-[#7a6a58]">Temukan produk dari toko ini dengan cepat.</p>
            </div>

            <form className="flex flex-col gap-3 sm:flex-row" method="get">
              <Input defaultValue={query} name="q" placeholder="Cari nama produk..." />
              {categorySlug ? <input name="category" type="hidden" value={categorySlug} /> : null}
              <button className="h-10 rounded-xl bg-[#2f2923] px-4 text-sm font-semibold text-[#fffaf2] transition hover:bg-[#1f1a16]">
                Cari
              </button>
            </form>
          </div>

          <div className="space-y-4">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
              <div>
                <h2 className="text-2xl font-bold text-[#241c16]">Produk</h2>
                <p className="text-sm text-[#7a6a58]">
                  {selectedCategory ? `Kategori ${selectedCategory.name}` : "Semua produk aktif dari toko ini."}
                </p>
              </div>
              <Link
                className="text-sm font-semibold text-[#7a4f2f] transition hover:text-[#4e321f]"
                href={`/s/${store.slug}/products`}
              >
                Lihat semua produk
              </Link>
            </div>

            {categories.length === 0 ? (
              <EmptyState
                title="Kategori belum tersedia"
                description="Toko ini belum menambahkan kategori publik. Semua produk tetap ditampilkan."
              />
            ) : (
              <div className="flex flex-wrap gap-2">
                <CategoryLink
                  active={!selectedCategory}
                  href={buildProductListingHref(store.slug, { q: query })}
                  label="Semua"
                />
                {categories.map((category) => (
                  <CategoryLink
                    key={category.id}
                    active={category.slug === selectedCategory?.slug}
                    href={buildCategoryHref(store.slug, category.slug, { q: query })}
                    label={category.name}
                  />
                ))}
              </div>
            )}

            {products.items.length === 0 ? (
              <EmptyState
                title="Produk belum ditemukan"
                description={
                  query || selectedCategory
                    ? "Coba ubah kata kunci atau pilih kategori lain."
                    : "Toko ini belum menampilkan produk. Hubungi toko melalui WhatsApp untuk info terbaru."
                }
              />
            ) : (
              <div className="grid grid-cols-2 gap-3 sm:gap-4 md:grid-cols-3 xl:grid-cols-4">
                {products.items.map((product) => (
                  <ProductCard
                    key={product.id}
                    categoryName={selectedCategory?.name}
                    product={product}
                    storeSlug={store.slug}
                  />
                ))}
              </div>
            )}
          </div>
        </div>

        <aside className="space-y-4 rounded-[28px] border border-[#eadfce] bg-[#fffaf2] p-5 shadow-[0_14px_40px_rgba(89,63,38,0.07)] lg:sticky lg:top-24">
          <div>
            <h2 className="text-lg font-bold text-[#241c16]">Tentang toko</h2>
            <p className="mt-2 text-sm leading-6 text-[#6d5e4e]">
              {store.description ?? "Deskripsi toko belum tersedia."}
            </p>
          </div>

          <div className="space-y-2 text-sm text-[#6d5e4e]">
            <p>
              <span className="font-medium text-[#241c16]">Lokasi:</span>{" "}
              {[store.city, store.province].filter(Boolean).join(", ") || "Belum diisi"}
            </p>
            <p>
              <span className="font-medium text-[#241c16]">WhatsApp:</span> {store.whatsapp ?? "Belum diisi"}
            </p>
            <p>
              <span className="font-medium text-[#241c16]">Telepon:</span> {store.phone ?? "Belum diisi"}
            </p>
          </div>

          <p className="rounded-2xl border border-[#ead7bd] bg-[#fff4d8] p-4 text-sm leading-6 text-[#72512f]">
            {whatsappHref
              ? "Gunakan WhatsApp untuk menanyakan ketersediaan, varian, atau cara pemesanan langsung ke toko."
              : "Kontak WhatsApp belum tersedia. Gunakan nomor telepon jika toko sudah mengisinya."}
          </p>
        </aside>
      </section>
    </main>
  );
}

function CategoryLink({ active, href, label }: { active: boolean; href: string; label: string }) {
  return (
    <Link
      href={href}
      className={
        active
          ? "rounded-full bg-[#2f2923] px-4 py-2 text-sm font-semibold text-[#fffaf2]"
          : "rounded-full border border-[#eadfce] bg-white px-4 py-2 text-sm font-semibold text-[#5f5042] transition hover:border-[#9a6a43] hover:text-[#7a4f2f]"
      }
    >
      {label}
    </Link>
  );
}

function firstParam(value?: string | string[]) {
  return Array.isArray(value) ? value[0] : value;
}

function buildProductListingHref(storeSlug: string, params: { q?: string }) {
  const searchParams = new URLSearchParams();

  if (params.q) {
    searchParams.set("q", params.q);
  }

  const suffix = searchParams.size > 0 ? `?${searchParams.toString()}` : "";
  return `/s/${storeSlug}/products${suffix}`;
}

function buildCategoryHref(storeSlug: string, categorySlug: string, params: { q?: string }) {
  const searchParams = new URLSearchParams();

  if (params.q) {
    searchParams.set("q", params.q);
  }

  const suffix = searchParams.size > 0 ? `?${searchParams.toString()}` : "";
  return `/s/${storeSlug}/categories/${categorySlug}${suffix}`;
}

function buildWhatsappHref(value?: string) {
  if (!value) {
    return undefined;
  }

  const digits = value.replace(/\D/g, "");
  if (!digits) {
    return undefined;
  }

  const normalized = digits.startsWith("0") ? `62${digits.slice(1)}` : digits;
  return `https://wa.me/${normalized}`;
}
