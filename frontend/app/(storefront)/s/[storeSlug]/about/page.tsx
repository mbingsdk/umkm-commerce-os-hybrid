import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { EmptyState } from "@/components/feedback/empty-state";
import { getPublicStoreBySlug, isPublicNotFoundError } from "@/features/storefront/api/storefront.api";
import { SafeImage } from "@/features/storefront/components/safe-image";
import { buildStoreJsonLd, getSiteURL, serializeJsonLd, toAbsoluteURL } from "@/features/storefront/seo";
import type { PublicBusinessHour, PublicStore } from "@/features/storefront/types";

type AboutPageProps = {
  params: Promise<{ storeSlug: string }>;
};

const siteURL = getSiteURL();

export async function generateMetadata({ params }: AboutPageProps): Promise<Metadata> {
  const { storeSlug } = await params;

  try {
    const store = await getPublicStoreBySlug(storeSlug);
    const title = `Tentang ${store.name}`;
    const description =
      store.description ??
      `Kenali ${store.name}${store.city ? ` di ${store.city}` : ""}, toko UMKM di UMKM Commerce OS.`;
    const canonicalURL = `${siteURL}/s/${store.slug}/about`;
    const image = toAbsoluteURL(store.bannerUrl ?? store.logoUrl);

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
      title: "Tentang toko tidak ditemukan"
    };
  }
}

export default async function StorefrontAboutPage({ params }: AboutPageProps) {
  const { storeSlug } = await params;
  const store = await loadStore(storeSlug);
  const hasBusinessHours = store.businessHours.length > 0;

  return (
    <main>
      <script
        dangerouslySetInnerHTML={{ __html: serializeJsonLd(buildStoreJsonLd(store)) }}
        type="application/ld+json"
      />

      <section className="border-b border-neutral-200 bg-white">
        <div className="mx-auto grid max-w-6xl gap-6 px-4 py-8 sm:px-6 lg:grid-cols-[minmax(0,1fr)_320px] lg:px-8">
          <div className="space-y-5">
            <p className="text-sm font-semibold text-primary-700">Tentang toko</p>
            <div>
              <h1 className="text-3xl font-bold tracking-tight text-neutral-950 sm:text-5xl">{store.name}</h1>
              <p className="mt-4 max-w-3xl text-sm leading-7 text-neutral-600 sm:text-base">
                {store.description ??
                  "Toko ini belum menambahkan cerita singkat. Kamu tetap bisa melihat katalog produk atau menghubungi toko untuk informasi terbaru."}
              </p>
            </div>
            <div className="flex flex-col gap-3 sm:flex-row">
              <LinkButton href={`/s/${store.slug}/products`}>Lihat produk</LinkButton>
              <LinkButton href={`/s/${store.slug}/contact`} variant="outline">
                Hubungi toko
              </LinkButton>
            </div>
          </div>

          <div className="overflow-hidden rounded-3xl border border-neutral-200 bg-neutral-100 shadow-soft">
            <SafeImage
              alt=""
              className="h-56 w-full object-cover"
              fallbackClassName="h-56 w-full"
              fallbackLabel={store.name.slice(0, 1)}
              src={store.bannerUrl ?? store.logoUrl}
            />
          </div>
        </div>
      </section>

      <section className="mx-auto grid max-w-6xl gap-6 px-4 py-8 sm:px-6 lg:grid-cols-[1fr_1fr] lg:px-8">
        <article className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-soft">
          <h2 className="text-lg font-semibold text-neutral-950">Informasi dasar</h2>
          <dl className="mt-4 space-y-3 text-sm text-neutral-600">
            <InfoRow label="Lokasi" value={[store.city, store.province].filter(Boolean).join(", ") || "Belum diisi"} />
            <InfoRow label="WhatsApp" value={store.whatsapp ?? "Belum diisi"} />
            <InfoRow label="Telepon" value={store.phone ?? "Belum diisi"} />
          </dl>
        </article>

        <article className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-soft">
          <h2 className="text-lg font-semibold text-neutral-950">Jam operasional</h2>
          {hasBusinessHours ? (
            <div className="mt-4 space-y-2">
              {sortBusinessHours(store.businessHours).map((hour) => (
                <div key={hour.dayOfWeek} className="flex items-center justify-between gap-4 text-sm">
                  <span className="font-medium text-neutral-950">{dayLabel(hour.dayOfWeek)}</span>
                  <span className="text-neutral-600">
                    {hour.isClosed ? "Tutup" : `${hour.openTime ?? "--:--"} - ${hour.closeTime ?? "--:--"}`}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <EmptyState
              title="Jam operasional belum tersedia"
              description="Tanyakan langsung ke toko untuk memastikan waktu layanan."
            />
          )}
        </article>
      </section>
    </main>
  );
}

async function loadStore(storeSlug: string): Promise<PublicStore> {
  try {
    return await getPublicStoreBySlug(storeSlug);
  } catch (error) {
    if (isPublicNotFoundError(error)) {
      notFound();
    }
    throw error;
  }
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col gap-1 sm:flex-row sm:justify-between">
      <dt className="font-medium text-neutral-950">{label}</dt>
      <dd>{value}</dd>
    </div>
  );
}

function sortBusinessHours(hours: PublicBusinessHour[]) {
  return [...hours].sort((a, b) => a.dayOfWeek - b.dayOfWeek);
}

function dayLabel(day: number) {
  const labels: Record<number, string> = {
    1: "Senin",
    2: "Selasa",
    3: "Rabu",
    4: "Kamis",
    5: "Jumat",
    6: "Sabtu",
    7: "Minggu"
  };
  return labels[day] ?? `Hari ${day}`;
}

function LinkButton({
  href,
  children,
  variant = "primary"
}: {
  href: string;
  children: string;
  variant?: "primary" | "outline";
}) {
  return (
    <Link
      href={href}
      className={
        variant === "outline"
          ? "inline-flex min-h-12 items-center justify-center rounded-xl border border-neutral-300 bg-white px-5 text-base font-semibold text-neutral-900 transition hover:bg-neutral-50"
          : "inline-flex min-h-12 items-center justify-center rounded-xl bg-primary-600 px-5 text-base font-semibold text-white transition hover:bg-primary-700"
      }
    >
      {children}
    </Link>
  );
}
