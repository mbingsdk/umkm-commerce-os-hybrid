import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { EmptyState } from "@/components/feedback/empty-state";
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

      <section className="bg-[#F8F1E7]">
        <div className="mx-auto max-w-[1500px] px-4 py-3 sm:px-6 sm:py-5 lg:px-8">
          <div className="overflow-hidden rounded-[28px] border border-[#E3D2BC] bg-[#FFFDF8] shadow-[0_14px_42px_rgba(89,63,38,0.09)]">
            <div className="relative h-28 overflow-hidden bg-[radial-gradient(circle_at_20%_20%,#D99A3D_0,#D99A3D_18%,transparent_19%),linear-gradient(135deg,#6F7D55,#B96E45_48%,#251F1A)] sm:h-40 lg:h-44">
              {store.bannerUrl ? (
                <SafeImage
                  alt=""
                  className="h-full w-full object-cover"
                  fallbackClassName="h-full w-full bg-transparent"
                  src={store.bannerUrl}
                />
              ) : (
                <div className="absolute inset-0 bg-[linear-gradient(135deg,rgba(255,253,248,0.18),transparent_36%),radial-gradient(circle_at_82%_18%,rgba(255,245,222,0.28),transparent_28%)]" />
              )}
              <div className="absolute inset-0 bg-gradient-to-t from-[#251F1A]/65 via-[#251F1A]/10 to-transparent" />
              <Link
                className="absolute left-3 top-3 inline-flex h-8 items-center rounded-full border border-white/30 bg-[#251F1A]/55 px-2.5 text-xs font-semibold text-[#FFFDF8] shadow-sm backdrop-blur transition hover:bg-[#251F1A]/75 sm:left-4 sm:top-4"
                href="/"
              >
                &larr; Jelajah platform
              </Link>
            </div>

            <div className="relative grid gap-4 px-4 pb-4 pt-10 sm:px-5 sm:pb-5 sm:pt-12 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end lg:px-6">
              <div className="absolute left-4 top-0 -translate-y-1/2 sm:left-5 lg:left-6">
                {store.logoUrl ? (
                  <SafeImage
                    alt=""
                    className="h-16 w-16 rounded-[22px] border-4 border-[#FFFDF8] bg-[#FFFDF8] object-cover shadow-[0_10px_24px_rgba(37,31,26,0.16)] sm:h-20 sm:w-20"
                    fallbackClassName="h-16 w-16 rounded-[22px] border-4 border-[#FFFDF8] sm:h-20 sm:w-20"
                    src={store.logoUrl}
                  />
                ) : (
                  <div className="flex h-16 w-16 items-center justify-center rounded-[22px] border-4 border-[#FFFDF8] bg-[#6f4e37] text-xl font-bold text-[#FFFDF8] shadow-[0_10px_24px_rgba(37,31,26,0.16)] sm:h-20 sm:w-20 sm:text-2xl">
                    {store.name.slice(0, 1).toUpperCase()}
                  </div>
                )}
              </div>

              <div className="space-y-2">
                <div className="flex flex-wrap items-center gap-1.5 text-xs font-semibold">
                  <span className="rounded-full bg-[#F1E7D8] px-2.5 py-0.5 text-[#7C3F25]">Storefront aktif</span>
                  <span className="rounded-full border border-[#E3D2BC] bg-white px-2.5 py-0.5 text-[#6F6256]">
                    {[store.city, store.province].filter(Boolean).join(", ") || "Toko lokal"}
                  </span>
                </div>
                <div>
                  <h1 className="text-2xl font-bold tracking-tight text-[#251F1A] sm:text-3xl">{store.name}</h1>
                  <p className="mt-1 max-w-2xl text-sm leading-6 text-[#6F6256]">
                    {store.description ?? "Toko ini belum menambahkan deskripsi. Jelajahi katalog produk yang tersedia di bawah ini."}
                  </p>
                </div>
                <div className="flex flex-wrap gap-1.5 text-xs font-semibold text-[#6F6256]">
                  <span className="rounded-full bg-[#FFF5DE] px-2.5 py-0.5 text-[#7A4D1D]">Checkout ke toko</span>
                  <span className="rounded-full bg-[#EEF4EA] px-2.5 py-0.5 text-[#52613C]">Katalog publik</span>
                  <span className="rounded-full bg-[#EAF4F3] px-2.5 py-0.5 text-[#2F7C78]">Kontak jelas</span>
                </div>
              </div>

              <div className="space-y-2 rounded-[20px] border border-[#E3D2BC] bg-[#FFFDF8] p-3 shadow-[0_8px_24px_rgba(89,63,38,0.05)] lg:w-72">
                <p className="text-sm leading-5 text-[#6F6256]">
                  {whatsappHref
                    ? "Butuh tanya stok atau varian? Hubungi toko langsung sebelum checkout."
                    : "Kontak WhatsApp belum tersedia. Gunakan informasi kontak toko jika sudah diisi."}
                </p>
                <div className="flex flex-wrap gap-1.5">
                  {whatsappHref ? (
                    <a
                      className="inline-flex h-9 items-center justify-center rounded-xl bg-[#251F1A] px-4 text-sm font-semibold text-[#FFFDF8] transition hover:bg-[#16110E]"
                      href={whatsappHref}
                      rel="noopener noreferrer"
                      target="_blank"
                    >
                      Chat WhatsApp
                    </a>
                  ) : null}
                  <ShareLinkButton />
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto grid max-w-[1500px] gap-4 px-4 py-4 sm:px-6 sm:py-6 lg:grid-cols-[minmax(0,1fr)_300px] lg:items-start lg:px-8">
        <div className="space-y-4">
          <div className="space-y-3">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
              <div>
                <h2 className="text-xl font-bold text-[#251F1A] sm:text-2xl">Produk</h2>
                <p className="text-sm text-[#6F6256]">
                  {selectedCategory ? `Kategori ${selectedCategory.name}` : "Semua produk aktif dari toko ini."}
                </p>
              </div>
              <Link
                className="text-sm font-semibold text-[#B96E45] transition hover:text-[#7C3F25]"
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
              <div className="flex gap-1.5 overflow-x-auto pb-1 sm:flex-wrap sm:overflow-visible">
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
              <div className="grid grid-cols-2 gap-2 sm:gap-2.5 md:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
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

        <aside className="space-y-3 rounded-[24px] border border-[#E3D2BC] bg-[#FFFDF8] p-4 shadow-[0_10px_28px_rgba(89,63,38,0.06)] lg:sticky lg:top-20">
          <div>
            <h2 className="text-lg font-bold text-[#251F1A]">Tentang toko</h2>
            <p className="mt-1.5 text-sm leading-6 text-[#6F6256]">
              {store.description ?? "Deskripsi toko belum tersedia."}
            </p>
          </div>

          <div className="space-y-2 text-sm text-[#6F6256]">
            <p>
              <span className="font-medium text-[#251F1A]">Lokasi:</span>{" "}
              {[store.city, store.province].filter(Boolean).join(", ") || "Belum diisi"}
            </p>
            <p>
              <span className="font-medium text-[#251F1A]">WhatsApp:</span> {store.whatsapp ?? "Belum diisi"}
            </p>
            <p>
              <span className="font-medium text-[#251F1A]">Telepon:</span> {store.phone ?? "Belum diisi"}
            </p>
          </div>

          <p className="rounded-2xl border border-[#E8D2AA] bg-[#FFF5DE] p-3 text-sm leading-6 text-[#7A4D1D]">
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
          ? "shrink-0 rounded-full bg-[#251F1A] px-3 py-1.5 text-xs font-semibold text-[#FFFDF8] sm:text-sm"
          : "shrink-0 rounded-full border border-[#E3D2BC] bg-white px-3 py-1.5 text-xs font-semibold text-[#6F6256] transition hover:border-[#B96E45] hover:text-[#B96E45] sm:text-sm"
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
