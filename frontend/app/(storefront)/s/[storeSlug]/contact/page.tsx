import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { EmptyState } from "@/components/feedback/empty-state";
import { getPublicStoreBySlug, isPublicNotFoundError } from "@/features/storefront/api/storefront.api";
import { getSiteURL, toAbsoluteURL } from "@/features/storefront/seo";
import type { PublicStore } from "@/features/storefront/types";

type ContactPageProps = {
  params: Promise<{ storeSlug: string }>;
};

const siteURL = getSiteURL();

export async function generateMetadata({ params }: ContactPageProps): Promise<Metadata> {
  const { storeSlug } = await params;

  try {
    const store = await getPublicStoreBySlug(storeSlug);
    const title = `Kontak ${store.name}`;
    const description = `Hubungi ${store.name}${store.city ? ` di ${store.city}` : ""} untuk tanya produk, stok, atau pemesanan.`;
    const canonicalURL = `${siteURL}/s/${store.slug}/contact`;
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
      title: "Kontak toko tidak ditemukan"
    };
  }
}

export default async function StorefrontContactPage({ params }: ContactPageProps) {
  const { storeSlug } = await params;
  const store = await loadStore(storeSlug);
  const whatsappHref = buildWhatsappHref(store.whatsapp);
  const hasContact = Boolean(store.whatsapp || store.phone);

  return (
    <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
      <section className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-soft sm:p-8">
        <p className="text-sm font-semibold text-primary-700">Kontak toko</p>
        <h1 className="mt-3 text-3xl font-bold tracking-tight text-neutral-950 sm:text-5xl">{store.name}</h1>
        <p className="mt-4 max-w-3xl text-sm leading-7 text-neutral-600 sm:text-base">
          Gunakan kontak resmi toko untuk menanyakan produk, stok, atau detail pemesanan. Pembayaran otomatis belum
          tersedia di MVP; konfirmasi pembayaran tetap ditinjau oleh pemilik toko.
        </p>

        <div className="mt-6 flex flex-col gap-3 sm:flex-row">
          {whatsappHref ? (
            <ExternalButton href={whatsappHref}>Chat WhatsApp</ExternalButton>
          ) : null}
          <LinkButton href={`/s/${store.slug}/products`} variant="outline">
            Lihat produk
          </LinkButton>
        </div>
      </section>

      <section className="mt-6 grid gap-4 md:grid-cols-3">
        <ContactCard title="WhatsApp" value={store.whatsapp ?? "Belum diisi"} helper="Kontak cepat untuk tanya produk." />
        <ContactCard title="Telepon" value={store.phone ?? "Belum diisi"} helper="Nomor telepon toko jika tersedia." />
        <ContactCard
          title="Lokasi"
          value={[store.city, store.province].filter(Boolean).join(", ") || "Belum diisi"}
          helper="Alamat detail hanya ditampilkan jika toko menyediakannya."
        />
      </section>

      {!hasContact ? (
        <div className="mt-6">
          <EmptyState
            title="Kontak belum tersedia"
            description="Toko belum menambahkan nomor WhatsApp atau telepon publik. Coba cek lagi nanti atau lihat katalog produk lebih dulu."
            action={
              <LinkButton href={`/s/${store.slug}/products`} variant="outline">
                Lihat produk
              </LinkButton>
            }
          />
        </div>
      ) : null}
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

function ContactCard({ title, value, helper }: { title: string; value: string; helper: string }) {
  return (
    <article className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-soft">
      <p className="text-sm font-medium text-neutral-500">{title}</p>
      <p className="mt-2 break-words text-lg font-semibold text-neutral-950">{value}</p>
      <p className="mt-2 text-sm leading-6 text-neutral-500">{helper}</p>
    </article>
  );
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

function ExternalButton({ href, children }: { href: string; children: string }) {
  return (
    <a
      href={href}
      rel="noopener noreferrer"
      target="_blank"
      className="inline-flex min-h-12 items-center justify-center rounded-xl bg-primary-600 px-5 text-base font-semibold text-white transition hover:bg-primary-700"
    >
      {children}
    </a>
  );
}
