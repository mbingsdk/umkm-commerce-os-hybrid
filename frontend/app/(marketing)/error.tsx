"use client";

import { ErrorState } from "@/components/feedback/error-state";

export default function MarketingError({ reset }: { reset: () => void }) {
  return (
    <main className="mx-auto flex min-h-screen max-w-3xl items-center px-4 py-10 sm:px-6 lg:px-8">
      <ErrorState
        title="Halaman publik gagal dimuat"
        description="Coba muat ulang. Jika masih gagal, kemungkinan layanan discovery sedang tidak tersedia."
        onRetry={reset}
      />
    </main>
  );
}
