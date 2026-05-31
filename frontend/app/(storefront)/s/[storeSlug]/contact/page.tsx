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
    <main className="mx-auto max-w-[1500px] px-4 py-4 sm:px-6 sm:py-6 lg:px-8">
      <section className="rounded-[28px] border border-[#E3D2BC] bg-[#FFFDF8] p-5 shadow-[0_12px_36px_rgba(89,63,38,0.07)]">
        <p className="text-sm font-semibold text-[#B96E45]">Kontak toko</p>
        <h1 className="mt-2 text-2xl font-bold tracking-tight text-[#251F1A] sm:text-3xl">{store.name}</h1>
        <p className="mt-2 max-w-3xl text-sm leading-6 text-[#6F6256]">
          Gunakan kontak resmi toko untuk menanyakan produk, stok, atau detail pemesanan. Pembayaran otomatis belum
          tersedia di MVP; konfirmasi pembayaran tetap ditinjau oleh pemilik toko.
        </p>

        <div className="mt-4 flex flex-col gap-2 sm:flex-row">
          {whatsappHref ? (
            <ExternalButton href={whatsappHref}>Chat WhatsApp</ExternalButton>
          ) : null}
          <LinkButton href={`/s/${store.slug}/products`} variant="outline">
            Lihat produk
          </LinkButton>
        </div>
      </section>

      <section className="mt-4 grid gap-3 md:grid-cols-3">
        <ContactCard title="WhatsApp" value={store.whatsapp ?? "Belum diisi"} helper="Kontak cepat untuk tanya produk." />
        <ContactCard title="Telepon" value={store.phone ?? "Belum diisi"} helper="Nomor telepon toko jika tersedia." />
        <ContactCard
          title="Lokasi"
          value={[store.city, store.province].filter(Boolean).join(", ") || "Belum diisi"}
          helper="Alamat detail hanya ditampilkan jika toko menyediakannya."
        />
      </section>

      {!hasContact ? (
        <div className="mt-4">
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
    <article className="rounded-[24px] border border-[#E3D2BC] bg-[#FFFDF8] p-4 shadow-[0_8px_24px_rgba(89,63,38,0.055)]">
      <p className="text-sm font-medium text-[#6F6256]">{title}</p>
      <p className="mt-1.5 break-words text-base font-semibold text-[#251F1A]">{value}</p>
      <p className="mt-1.5 text-sm leading-6 text-[#6F6256]">{helper}</p>
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
          ? "inline-flex h-10 items-center justify-center rounded-xl border border-[#E3D2BC] bg-white px-4 text-sm font-semibold text-[#7C3F25] transition hover:bg-[#F8F1E7]"
          : "inline-flex h-10 items-center justify-center rounded-xl bg-[#251F1A] px-4 text-sm font-semibold text-[#FFFDF8] transition hover:bg-[#16110E]"
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
      className="inline-flex h-10 items-center justify-center rounded-xl bg-[#251F1A] px-4 text-sm font-semibold text-[#FFFDF8] transition hover:bg-[#16110E]"
    >
      {children}
    </a>
  );
}
