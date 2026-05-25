import type { Metadata } from "next";
import type { ReactNode } from "react";
import Link from "next/link";
import { EmptyState } from "@/components/feedback/empty-state";
import {
  listDiscoveryProducts,
  listDiscoveryStores
} from "@/features/discovery/api/discovery.api";
import {
  ProductSection,
  StoreSection
} from "@/features/discovery/components/discovery-sections";
import { publicPageMetadata } from "@/lib/seo/metadata";

type CityDiscoveryPageProps = {
  params: Promise<{ citySlug: string }>;
};

export async function generateMetadata({ params }: CityDiscoveryPageProps): Promise<Metadata> {
  const { citySlug } = await params;
  const cityName = cityFromSlug(citySlug);

  return publicPageMetadata({
    title: `UMKM ${cityName} - Toko dan Produk Lokal`,
    description: `Jelajahi toko dan produk UMKM publik dari ${cityName}. Semua hasil mengarah ke storefront tenant.`,
    path: `/city/${citySlug}`
  });
}

export default async function CityDiscoveryPage({ params }: CityDiscoveryPageProps) {
  const { citySlug } = await params;
  const cityName = cityFromSlug(citySlug);

  const [stores, products] = await Promise.all([
    listDiscoveryStores({ city: cityName, limit: 9 }),
    listDiscoveryProducts({ city: cityName, limit: 12 })
  ]);

  const hasResults = stores.items.length > 0 || products.items.length > 0;

  return (
    <main className="min-h-screen bg-neutral-50">
      <section className="border-b border-neutral-200 bg-gradient-to-br from-primary-50 via-white to-neutral-50">
        <div className="mx-auto space-y-5 px-4 py-10 sm:px-6 lg:max-w-6xl lg:px-8 lg:py-14">
          <p className="text-sm font-semibold text-primary-700">Discovery kota</p>
          <div className="max-w-3xl">
            <h1 className="text-3xl font-bold tracking-tight text-neutral-950 sm:text-5xl">UMKM {cityName}</h1>
            <p className="mt-4 text-sm leading-7 text-neutral-600 sm:text-base">
              Lihat toko dan produk UMKM publik dari {cityName}. Customer tetap diarahkan ke storefront tenant
              untuk melihat detail produk atau menghubungi toko.
            </p>
          </div>

          <div className="flex flex-col gap-3 sm:flex-row">
            <LinkButton href={`/stores?city=${encodeURIComponent(cityName)}`}>Lihat toko di {cityName}</LinkButton>
            <LinkButton href={`/products?city=${encodeURIComponent(cityName)}`}>Lihat produk di {cityName}</LinkButton>
            <LinkButton href={`/search?city=${encodeURIComponent(cityName)}`}>Cari di kota ini</LinkButton>
          </div>
        </div>
      </section>

      <section className="mx-auto space-y-10 px-4 py-10 sm:px-6 lg:max-w-6xl lg:px-8">
        {!hasResults ? (
          <EmptyState
            title="Belum ada hasil di kota ini"
            description="Toko atau produk dari kota ini akan muncul setelah memenuhi aturan discovery publik."
            action={<LinkButton href="/stores">Jelajahi semua toko</LinkButton>}
          />
        ) : null}

        <StoreSection
          title={`Toko di ${cityName}`}
          description="Toko published dari tenant aktif atau trialing."
          stores={stores.items}
          href={`/stores?city=${encodeURIComponent(cityName)}`}
        />

        <ProductSection
          title={`Produk dari ${cityName}`}
          description="Produk aktif dan discoverable dari toko di kota ini."
          products={products.items}
          href={`/products?city=${encodeURIComponent(cityName)}`}
        />
      </section>
    </main>
  );
}

function cityFromSlug(value: string) {
  return decodeURIComponent(value)
    .split("-")
    .filter(Boolean)
    .map((word) => word.slice(0, 1).toUpperCase() + word.slice(1))
    .join(" ");
}

function LinkButton({ href, children }: { href: string; children: ReactNode }) {
  return (
    <Link
      className="inline-flex min-h-11 items-center justify-center rounded-xl border border-neutral-300 bg-white px-4 py-2 text-sm font-semibold text-neutral-900 transition hover:bg-neutral-50"
      href={href}
    >
      {children}
    </Link>
  );
}
