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
      <Card className="overflow-hidden border-[#eadfce] bg-white/90 shadow-[0_18px_55px_rgba(89,63,38,0.1)]">
        <div className="bg-[#2f2923] p-6 text-[#fffaf2] sm:p-8">
          <Badge className="bg-white/15 text-[#fffaf2]" tone="neutral">
            Pesanan dibuat
          </Badge>
          <h1 className="mt-4 text-2xl font-bold tracking-tight sm:text-3xl">Pesanan berhasil dibuat</h1>
          <p className="mt-3 max-w-2xl text-sm leading-7 text-[#eadfce]">
            Simpan nomor order ini untuk komunikasi dengan toko. Pembayaran dilakukan manual dan akan ditinjau oleh penjual.
          </p>
        </div>

        <CardHeader>
          <CardTitle>Detail pesanan</CardTitle>
          <CardDescription>Toko akan menghubungi kamu untuk instruksi pembayaran dan konfirmasi pesanan.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-5">
          <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf2] p-4">
            <p className="text-xs font-semibold uppercase tracking-wide text-[#7b6a58]">Nomor order</p>
            <p className="mt-1 text-xl font-bold text-[#241c16]">{orderNumber ?? "Belum tersedia"}</p>
          </div>

          {grandTotal > 0 ? (
            <div className="flex items-center justify-between rounded-2xl border border-[#eadfce] p-4">
              <span className="text-sm text-[#7b6a58]">Total final dari backend</span>
              <span className="text-lg font-bold text-[#7a4f2f]">{formatRupiah(grandTotal)}</span>
            </div>
          ) : null}

          <div className="rounded-2xl border border-[#d8c7ad] bg-[#fff6df] p-4 text-sm leading-6 text-[#4b3a29]">
            {paymentMessage ??
              "Silakan ikuti instruksi pembayaran dari toko. Setelah transfer, kirim konfirmasi pembayaran agar penjual bisa meninjau pesananmu."}
          </div>

          <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap">
            <Link
              className="inline-flex h-11 flex-1 items-center justify-center rounded-xl bg-[#2f2923] px-4 text-sm font-semibold text-[#fffaf2] transition hover:bg-[#1f1a16]"
              href={`/s/${storeSlug}`}
            >
              Lanjut Belanja
            </Link>
            {orderNumber ? (
              <Link
                className="inline-flex h-11 flex-1 items-center justify-center rounded-xl border border-[#d8c7ad] bg-[#fffaf2] px-4 text-sm font-semibold text-[#7a4f2f] transition hover:bg-[#f3eadc]"
                href={`/s/${storeSlug}/orders/${encodeURIComponent(orderNumber)}/payment-confirmation`}
              >
                Konfirmasi pembayaran
              </Link>
            ) : null}
            <Link
              className="inline-flex h-11 flex-1 items-center justify-center rounded-xl border border-[#d8c7ad] bg-white px-4 text-sm font-semibold text-[#3b2f24] transition hover:bg-[#f7f1e8]"
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
