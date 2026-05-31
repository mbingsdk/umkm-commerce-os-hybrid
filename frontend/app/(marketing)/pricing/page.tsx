import type { Metadata } from "next";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatRupiah } from "@/lib/format/money";
import { publicPageMetadata } from "@/lib/seo/metadata";

export function generateMetadata(): Metadata {
  return publicPageMetadata({
    title: "Paket Harga UMKM Commerce OS",
    description:
      "Pilihan paket UMKM Commerce OS untuk storefront, katalog, inventory, POS online-first, order, finance dasar, courier lokal, dan discovery.",
    path: "/pricing"
  });
}

type Plan = {
  name: string;
  priceMonthly: number;
  audience: string;
  highlight?: string;
  badge?: string;
  benefits: string[];
  limits: string[];
  ctaLabel: string;
  ctaHref: string;
};

const plans: Plan[] = [
  {
    name: "Starter",
    priceMonthly: 99000,
    audience: "Untuk toko kecil yang ingin mulai jualan online dengan stok dan kasir sederhana.",
    benefits: [
      "Storefront publik SEO-friendly",
      "Katalog produk dan kategori",
      "Checkout manual payment",
      "Inventory dan riwayat stok dasar",
      "POS online-first"
    ],
    limits: ["Maks. 100 produk", "Maks. 2 staff", "Courier lokal tidak termasuk di MVP"],
    ctaLabel: "Mulai dari Starter",
    ctaHref: "/register"
  },
  {
    name: "Growth",
    priceMonthly: 199000,
    audience: "Untuk toko yang mulai ramai dan butuh staff, kurir lokal, discovery, dan limit produk lebih besar.",
    highlight: "Paling cocok untuk pilot aktif",
    badge: "Rekomendasi",
    benefits: [
      "Semua fitur Starter",
      "Limit produk lebih besar",
      "Courier zone lokal",
      "Discovery platform",
      "Finance summary dan expense tracking"
    ],
    limits: ["Maks. 1.000 produk", "Maks. 5 staff", "Custom domain belum tersedia di MVP"],
    ctaLabel: "Pilih Growth",
    ctaHref: "/register"
  },
  {
    name: "Business",
    priceMonthly: 499000,
    audience: "Untuk bisnis dengan banyak produk, beberapa staff, dan kebutuhan operasional lebih lengkap.",
    benefits: [
      "Semua fitur Growth",
      "Limit produk tinggi",
      "Limit staff lebih besar",
      "Discovery priority eligible",
      "Dukungan operasional pilot lebih prioritas"
    ],
    limits: ["Maks. 10.000 produk", "Maks. 20 staff", "Advanced accounting belum termasuk"],
    ctaLabel: "Diskusikan Business",
    ctaHref: "/register"
  }
];

const featureRows = [
  ["Storefront publik", "Ya", "Ya", "Ya"],
  ["Product limit", "100", "1.000", "10.000"],
  ["Staff limit", "2", "5", "20"],
  ["Katalog dan kategori", "Ya", "Ya", "Ya"],
  ["Inventory tracking", "Ya", "Ya", "Ya"],
  ["Checkout online", "Manual payment", "Manual payment", "Manual payment"],
  ["POS", "Online-first", "Online-first", "Online-first"],
  ["Finance summary", "Basic", "Ya", "Ya"],
  ["Expense tracking", "Ya", "Ya", "Ya"],
  ["Courier zone lokal", "Tidak termasuk", "Ya", "Ya"],
  ["Discovery platform", "Ya", "Ya", "Priority eligible"],
  ["Payment gateway", "Belum tersedia", "Belum tersedia", "Belum tersedia"],
  ["Offline POS", "Belum tersedia", "Belum tersedia", "Belum tersedia"]
];

const faqs = [
  {
    question: "Apakah harga ini sudah final?",
    answer:
      "Harga ini adalah rekomendasi awal untuk validasi pilot. Paket dan angka bisa berubah setelah feedback UMKM nyata terkumpul."
  },
  {
    question: "Apakah pembayaran customer otomatis terkonfirmasi?",
    answer:
      "Belum. MVP memakai manual payment confirmation: customer mengirim konfirmasi pembayaran, lalu owner atau manager meninjau dari dashboard."
  },
  {
    question: "Apakah ada payment gateway atau biaya transaksi?",
    answer:
      "Belum ada payment gateway dan belum ada transaction fee di MVP. Fokus pilot adalah validasi storefront, order, stok, POS, dan operasional toko."
  },
  {
    question: "Apakah POS bisa dipakai offline?",
    answer:
      "Belum. POS MVP bersifat online-first agar stok, order, dan finance tetap konsisten."
  },
  {
    question: "Apakah laporan finance sudah menghitung laba bersih akuntansi?",
    answer:
      "Belum sepenuhnya. Net estimate belum menghitung HPP/modal produk secara detail, jadi gunakan sebagai estimasi operasional awal."
  }
];

export default function PricingPage() {
  return (
    <main className="min-h-screen bg-neutral-50">
      <section className="border-b border-neutral-200 bg-gradient-to-br from-primary-50 via-white to-neutral-50">
        <div className="mx-auto grid max-w-6xl gap-8 px-4 py-12 sm:px-6 lg:grid-cols-[minmax(0,1fr)_360px] lg:px-8 lg:py-16">
          <div>
            <p className="text-sm font-semibold text-primary-700">Pricing pilot UMKM</p>
            <h1 className="mt-4 text-4xl font-bold tracking-tight text-neutral-950 sm:text-5xl">
              Paket sederhana untuk mulai jualan, mengelola stok, dan menjalankan kasir.
            </h1>
            <p className="mt-5 max-w-2xl text-base leading-7 text-neutral-600">
              Mulai dari storefront publik, katalog produk, checkout manual, inventory, POS online-first,
              finance dasar, sampai courier dan discovery untuk toko yang mulai tumbuh.
            </p>
            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <Link
                className="inline-flex h-12 items-center justify-center rounded-xl bg-primary-600 px-5 text-base font-semibold text-white transition hover:bg-primary-700"
                href="/register"
              >
                Mulai buat toko
              </Link>
              <Link
                className="inline-flex h-12 items-center justify-center rounded-xl border border-neutral-300 bg-white px-5 text-base font-semibold text-neutral-900 transition hover:bg-neutral-50"
                href="/stores"
              >
                Lihat toko publik
              </Link>
            </div>
          </div>

          <Card className="self-start border-primary-100 bg-white/90">
            <CardHeader>
              <CardTitle>Pilot Partner</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm leading-6 text-neutral-600">
              <p>
                Untuk pilot awal, tenant terpilih bisa memakai paket Pilot Partner selama 30-90 hari atau
                Rp0 sampai produk stabil, dengan feedback aktif sebagai bagian dari program.
              </p>
              <p className="rounded-2xl bg-amber-50 p-3 text-amber-800">
                Paket pilot diaktifkan manual oleh tim platform. Billing berlangganan dan payment gateway
                subscription belum tersedia di MVP.
              </p>
            </CardContent>
          </Card>
        </div>
      </section>

      <section className="mx-auto space-y-8 px-4 py-10 sm:px-6 lg:max-w-6xl lg:px-8">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-neutral-950">Pilih paket sesuai tahap toko</h2>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-neutral-500">
            Tiga paket publik ini mengikuti strategi MVP: sederhana, mudah dijelaskan, dan limit berdasarkan
            kebutuhan operasional nyata.
          </p>
        </div>

        <div className="grid gap-4 lg:grid-cols-3">
          {plans.map((plan) => (
            <PlanCard key={plan.name} plan={plan} />
          ))}
        </div>
      </section>

      <section className="mx-auto space-y-6 px-4 pb-10 sm:px-6 lg:max-w-6xl lg:px-8">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-neutral-950">Perbandingan fitur</h2>
          <p className="mt-2 text-sm leading-6 text-neutral-500">
            Fitur yang belum tersedia di MVP ditulis apa adanya supaya ekspektasi tetap jernih.
          </p>
        </div>

        <div className="overflow-hidden rounded-3xl border border-neutral-200 bg-white shadow-soft">
          <div className="overflow-x-auto">
            <table className="min-w-[760px] w-full text-left text-sm">
              <thead className="bg-neutral-50 text-neutral-600">
                <tr>
                  <th className="px-5 py-4 font-semibold">Fitur</th>
                  <th className="px-5 py-4 font-semibold">Starter</th>
                  <th className="px-5 py-4 font-semibold">Growth</th>
                  <th className="px-5 py-4 font-semibold">Business</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-neutral-100 text-neutral-700">
                {featureRows.map(([feature, starter, growth, business]) => (
                  <tr key={feature}>
                    <td className="px-5 py-4 font-medium text-neutral-950">{feature}</td>
                    <td className="px-5 py-4">{starter}</td>
                    <td className="px-5 py-4">{growth}</td>
                    <td className="px-5 py-4">{business}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <section className="mx-auto grid gap-6 px-4 pb-10 sm:px-6 lg:max-w-6xl lg:grid-cols-[minmax(0,1fr)_360px] lg:px-8">
        <div className="space-y-4">
          <h2 className="text-2xl font-bold tracking-tight text-neutral-950">FAQ</h2>
          <div className="space-y-3">
            {faqs.map((faq) => (
              <Card key={faq.question} className="shadow-none">
                <CardContent className="space-y-2">
                  <h3 className="font-semibold text-neutral-950">{faq.question}</h3>
                  <p className="text-sm leading-6 text-neutral-600">{faq.answer}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>

        <Card className="self-start border-primary-100 bg-primary-700 text-white">
          <CardContent className="space-y-4">
            <Badge className="bg-white/15 text-white" tone="neutral">
              Siap pilot?
            </Badge>
            <h2 className="text-2xl font-bold tracking-tight">Buat toko pertamamu hari ini.</h2>
            <p className="text-sm leading-6 text-primary-50">
              Daftar, buat store, tambah produk, lalu publish storefront. Untuk pilot, proses billing dan aktivasi
              paket masih dibantu secara manual.
            </p>
            <Link
              className="inline-flex h-11 w-full items-center justify-center rounded-xl bg-white px-5 text-sm font-semibold text-primary-800 transition hover:bg-primary-50"
              href="/register"
            >
              Daftar dan buat toko
            </Link>
          </CardContent>
        </Card>
      </section>
    </main>
  );
}

function PlanCard({ plan }: { plan: Plan }) {
  return (
    <Card className={plan.badge ? "relative border-primary-300 shadow-md" : ""}>
      {plan.badge ? (
        <div className="absolute right-5 top-5">
          <Badge tone="primary">{plan.badge}</Badge>
        </div>
      ) : null}
      <CardHeader className="pr-28">
        <p className="text-sm font-semibold text-primary-700">{plan.name}</p>
        <CardTitle className="text-2xl">{formatRupiah(plan.priceMonthly)}</CardTitle>
        <p className="text-sm text-neutral-500">per bulan</p>
      </CardHeader>
      <CardContent className="space-y-5">
        <p className="text-sm leading-6 text-neutral-600">{plan.audience}</p>
        {plan.highlight ? (
          <p className="rounded-2xl bg-primary-50 p-3 text-sm font-medium text-primary-900">{plan.highlight}</p>
        ) : null}

        <div>
          <h3 className="text-sm font-semibold text-neutral-950">Termasuk</h3>
          <ul className="mt-3 space-y-2 text-sm leading-6 text-neutral-600">
            {plan.benefits.map((item) => (
              <li key={item} className="flex gap-2">
                <span aria-hidden="true" className="text-primary-700">
                  &#10003;
                </span>
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </div>

        <div>
          <h3 className="text-sm font-semibold text-neutral-950">Limit utama</h3>
          <ul className="mt-3 space-y-2 text-sm leading-6 text-neutral-600">
            {plan.limits.map((item) => (
              <li key={item} className="flex gap-2">
                <span className="text-neutral-400">-</span>
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </div>

        <Link
          className={
            plan.badge
              ? "inline-flex h-11 w-full items-center justify-center rounded-xl bg-primary-600 px-4 text-sm font-semibold text-white transition hover:bg-primary-700"
              : "inline-flex h-11 w-full items-center justify-center rounded-xl border border-neutral-300 bg-white px-4 text-sm font-semibold text-neutral-900 transition hover:bg-neutral-50"
          }
          href={plan.ctaHref}
        >
          {plan.ctaLabel}
        </Link>
      </CardContent>
    </Card>
  );
}
