type StorefrontPageProps = {
  params: Promise<{ storeSlug: string }>;
};

export default async function StorefrontHomePage({ params }: StorefrontPageProps) {
  const { storeSlug } = await params;

  return (
    <main className="mx-auto max-w-5xl px-4 py-12 sm:px-6 lg:px-8">
      <p className="text-sm font-semibold text-primary-700">Storefront foundation</p>
      <h1 className="mt-2 text-3xl font-bold text-neutral-950">{storeSlug}</h1>
      <p className="mt-3 max-w-2xl text-sm leading-6 text-neutral-500">
        Route storefront sudah disiapkan untuk rendering server-side. Data toko publik akan dihubungkan setelah API
        publik tersedia.
      </p>
    </main>
  );
}

