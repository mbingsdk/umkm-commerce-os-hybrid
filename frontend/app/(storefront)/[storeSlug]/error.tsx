"use client";

import { ErrorState } from "@/components/feedback/error-state";

export default function StorefrontError({ reset }: { reset: () => void }) {
  return (
    <main className="mx-auto max-w-5xl px-4 py-12 sm:px-6 lg:px-8">
      <ErrorState
        title="Storefront gagal dimuat"
        description="Coba muat ulang halaman toko ini."
        onRetry={reset}
      />
    </main>
  );
}
