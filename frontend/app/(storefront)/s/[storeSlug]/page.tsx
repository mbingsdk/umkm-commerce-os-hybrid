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
import { getSiteURL, toAbsoluteURL } from "@/features/storefront/seo";
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

  return (
    <main>
      <section className="border-b border-neutral-200 bg-white">
        <div className="mx-auto max-w-6xl px-4 py-6 sm:px-6 sm:py-8 lg:px-8">
          <div className="overflow-hidden rounded-3xl bg-gradient-to-br from-primary-700 via-primary-700 to-neutral-950 text-white shadow-soft">
            {store.bannerUrl ? (
              <SafeImage
                alt=""
                className="h-40 w-full object-cover opacity-80 sm:h-52"
                fallbackClassName="h-40 w-full bg-transparent sm:h-52"
                src={store.bannerUrl}
              />
            ) : null}
            <div className="grid gap-5 p-5 sm:p-8 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
              <div className="space-y-3">
                <p className="text-sm text-primary-100">
                  {store.city ?? "Toko lokal"}
                  {store.province ? `, ${store.province}` : ""}
                </p>
                <h1 className="text-2xl font-bold tracking-tight sm:text-4xl">{store.name}</h1>
                <p className="max-w-2xl text-sm leading-7 text-primary-50">
                  {store.description ?? "Toko ini belum menambahkan deskripsi."}
                </p>
              </div>

              <div className="space-y-3">
                <p className="max-w-sm text-sm leading-6 text-primary-50">
                  {whatsappHref
                    ? "Hubungi toko via WhatsApp untuk tanya stok dan pemesanan."
                    : "Kontak WhatsApp belum tersedia. Lihat informasi toko untuk kontak lain."}
                </p>
                <div className="flex flex-wrap gap-2">
                  {whatsappHref ? (
                    <a
                      className="inline-flex h-9 items-center justify-center rounded-xl bg-white px-4 text-sm font-semibold text-primary-800 transition hover:bg-primary-50"
                      href={whatsappHref}
                      rel="noopener noreferrer"
                      target="_blank"
                    >
                      Chat WhatsApp
                    </a>
                  ) : null}
                  <div className="[&_button]:border-white/30 [&_button]:bg-white/10 [&_button]:text-white [&_button:hover]:bg-white/20 [&_span]:text-primary-50">
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
          <div className="space-y-4 rounded-3xl border border-neutral-200 bg-white p-5 shadow-soft">
            <div>
              <h2 className="text-lg font-semibold text-neutral-950">Cari produk</h2>
              <p className="mt-1 text-sm text-neutral-500">Temukan produk dari toko ini dengan cepat.</p>
            </div>

            <form className="flex flex-col gap-3 sm:flex-row" method="get">
              <Input defaultValue={query} name="q" placeholder="Cari nama produk..." />
              {categorySlug ? <input name="category" type="hidden" value={categorySlug} /> : null}
              <button className="h-10 rounded-xl bg-primary-600 px-4 text-sm font-semibold text-white transition hover:bg-primary-700">
                Cari
              </button>
            </form>
          </div>

          <div className="space-y-4">
            <div>
              <h2 className="text-xl font-semibold text-neutral-950">Produk</h2>
              <p className="text-sm text-neutral-500">
                {selectedCategory ? `Kategori ${selectedCategory.name}` : "Semua produk aktif dari toko ini."}
              </p>
            </div>

            {categories.length === 0 ? (
              <EmptyState
                title="Kategori belum tersedia"
                description="Toko ini belum menambahkan kategori publik. Semua produk tetap ditampilkan."
              />
            ) : (
              <div className="flex flex-wrap gap-2">
                <CategoryLink active={!selectedCategory} href={buildStoreHref(store.slug, { q: query })} label="Semua" />
                {categories.map((category) => (
                  <CategoryLink
                    key={category.id}
                    active={category.slug === selectedCategory?.slug}
                    href={buildStoreHref(store.slug, { q: query, category: category.slug })}
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

        <aside className="space-y-4 rounded-3xl border border-neutral-200 bg-white p-5 shadow-soft lg:sticky lg:top-6">
          <div>
            <h2 className="text-lg font-semibold text-neutral-950">Tentang toko</h2>
            <p className="mt-2 text-sm leading-6 text-neutral-600">
              {store.description ?? "Deskripsi toko belum tersedia."}
            </p>
          </div>

          <div className="space-y-2 text-sm text-neutral-600">
            <p>
              <span className="font-medium text-neutral-950">Lokasi:</span>{" "}
              {[store.city, store.province].filter(Boolean).join(", ") || "Belum diisi"}
            </p>
            <p>
              <span className="font-medium text-neutral-950">WhatsApp:</span> {store.whatsapp ?? "Belum diisi"}
            </p>
            <p>
              <span className="font-medium text-neutral-950">Telepon:</span> {store.phone ?? "Belum diisi"}
            </p>
          </div>

          <p className="rounded-2xl bg-primary-50 p-4 text-sm leading-6 text-primary-900">
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
          ? "rounded-full bg-primary-600 px-4 py-2 text-sm font-semibold text-white"
          : "rounded-full border border-neutral-200 bg-white px-4 py-2 text-sm font-semibold text-neutral-700 transition hover:border-primary-300 hover:text-primary-700"
      }
    >
      {label}
    </Link>
  );
}

function firstParam(value?: string | string[]) {
  return Array.isArray(value) ? value[0] : value;
}

function buildStoreHref(storeSlug: string, params: { q?: string; category?: string }) {
  const searchParams = new URLSearchParams();

  if (params.q) {
    searchParams.set("q", params.q);
  }
  if (params.category) {
    searchParams.set("category", params.category);
  }

  const suffix = searchParams.size > 0 ? `?${searchParams.toString()}` : "";
  return `/s/${storeSlug}${suffix}`;
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
