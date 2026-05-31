"use client";

import { ErrorState } from "@/components/feedback/error-state";

export default function StorefrontError({ reset }: { reset: () => void }) {
  return (
    <main className="mx-auto max-w-[1500px] px-4 py-6 sm:px-6 lg:px-8">
      <ErrorState
        title="Storefront gagal dimuat"
        description="Koneksi sedang bermasalah. Coba muat ulang halaman toko ini."
        onRetry={reset}
      />
    </main>
  );
}
