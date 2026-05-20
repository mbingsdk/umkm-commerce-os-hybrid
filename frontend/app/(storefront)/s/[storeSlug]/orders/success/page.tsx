import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { formatRupiah } from "@/lib/format/money";

type PageProps = {
  params: Promise<{ storeSlug: string }>;
  searchParams: Promise<{
    order_number?: string | string[];
    grand_total?: string | string[];
    payment_message?: string | string[];
  }>;
};

export default async function OrderSuccessPage({ params, searchParams }: PageProps) {
  const [{ storeSlug }, rawSearchParams] = await Promise.all([params, searchParams]);
  const orderNumber = firstParam(rawSearchParams.order_number);
  const paymentMessage = firstParam(rawSearchParams.payment_message);
  const grandTotal = Number(firstParam(rawSearchParams.grand_total) ?? 0);

  return (
    <main className="mx-auto max-w-3xl px-4 py-10 sm:px-6 lg:px-8">
      <Card className="overflow-hidden">
        <div className="bg-gradient-to-br from-primary-600 to-neutral-950 p-6 text-white sm:p-8">
          <Badge className="bg-white/15 text-white" tone="neutral">
            Pesanan dibuat
          </Badge>
          <h1 className="mt-4 text-2xl font-bold tracking-tight sm:text-3xl">Pesanan berhasil dibuat</h1>
          <p className="mt-3 max-w-2xl text-sm leading-7 text-primary-50">
            Simpan nomor order ini untuk komunikasi dengan toko. Pembayaran masih diproses manual oleh penjual.
          </p>
        </div>

        <CardHeader>
          <CardTitle>Detail pesanan</CardTitle>
          <CardDescription>Toko akan menghubungi kamu untuk instruksi pembayaran dan konfirmasi pesanan.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-5">
          <div className="rounded-2xl bg-neutral-50 p-4">
            <p className="text-xs font-semibold uppercase tracking-wide text-neutral-500">Nomor order</p>
            <p className="mt-1 text-xl font-bold text-neutral-950">{orderNumber ?? "Belum tersedia"}</p>
          </div>

          {grandTotal > 0 ? (
            <div className="flex items-center justify-between rounded-2xl border border-neutral-200 p-4">
              <span className="text-sm text-neutral-500">Total final dari backend</span>
              <span className="text-lg font-bold text-primary-700">{formatRupiah(grandTotal)}</span>
            </div>
          ) : null}

          <div className="rounded-2xl border border-primary-100 bg-primary-50 p-4 text-sm leading-6 text-primary-950">
            {paymentMessage ??
              "Silakan tunggu instruksi pembayaran manual dari toko. Jika perlu lebih cepat, hubungi toko melalui kontak yang tersedia."}
          </div>

          <div className="flex flex-col gap-3 sm:flex-row">
            <Link
              className="inline-flex h-11 flex-1 items-center justify-center rounded-xl bg-primary-600 px-4 text-sm font-semibold text-white transition hover:bg-primary-700"
              href={`/s/${storeSlug}`}
            >
              Lanjut Belanja
            </Link>
            <Link
              className="inline-flex h-11 flex-1 items-center justify-center rounded-xl border border-neutral-300 bg-white px-4 text-sm font-semibold text-neutral-900 transition hover:bg-neutral-50"
              href={`/s/${storeSlug}/cart`}
            >
              Lihat Keranjang
            </Link>
          </div>
        </CardContent>
      </Card>
    </main>
  );
}

function firstParam(value?: string | string[]) {
  return Array.isArray(value) ? value[0] : value;
}
