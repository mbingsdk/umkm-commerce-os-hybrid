import Link from "next/link";

export default function StoreNotFound() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-[#F8F1E7] px-4">
      <section className="max-w-md rounded-[28px] border border-[#E3D2BC] bg-[#FFFDF8] p-6 text-center shadow-[0_14px_42px_rgba(89,63,38,0.09)]">
        <p className="text-sm font-semibold text-[#B96E45]">404</p>
        <h1 className="mt-2 text-2xl font-bold text-[#251F1A]">Toko tidak ditemukan</h1>
        <p className="mt-2 text-sm leading-6 text-[#6F6256]">
          Toko ini mungkin belum dipublikasikan atau sudah tidak tersedia.
        </p>
        <Link
          href="/"
          className="mt-5 inline-flex rounded-xl bg-[#251F1A] px-4 py-2 text-sm font-semibold text-[#FFFDF8] transition hover:bg-[#16110E]"
        >
          Jelajah toko lain
        </Link>
      </section>
    </main>
  );
}
