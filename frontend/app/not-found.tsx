import Link from "next/link";

export default function NotFound() {
  return (
    <main className="flex min-h-screen items-center justify-center px-4">
      <section className="max-w-md rounded-3xl border border-neutral-200 bg-white p-8 text-center shadow-soft">
        <p className="text-sm font-semibold text-primary-700">404</p>
        <h1 className="mt-2 text-2xl font-bold text-neutral-950">Halaman tidak ditemukan</h1>
        <p className="mt-3 text-sm text-neutral-500">
          Link yang kamu buka tidak tersedia, sudah dipindahkan, atau tidak bisa diakses dari akun ini.
        </p>
        <div className="mt-6 flex flex-col gap-3 sm:flex-row sm:justify-center">
          <Link
            href="/"
            className="inline-flex rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-primary-700"
          >
            Ke beranda
          </Link>
          <Link
            href="/dashboard"
            className="inline-flex rounded-xl border border-neutral-300 bg-white px-4 py-2 text-sm font-semibold text-neutral-900 transition hover:bg-neutral-50"
          >
            Ke dashboard
          </Link>
        </div>
      </section>
    </main>
  );
}
