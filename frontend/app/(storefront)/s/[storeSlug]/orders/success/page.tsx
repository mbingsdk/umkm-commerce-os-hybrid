import Link from "next/link";
import { PaymentNotice } from "@/components/public/public-ui";
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
      <Card className="overflow-hidden border-[#E3D2BC] bg-white/90 shadow-[0_18px_55px_rgba(89,63,38,0.1)]">
        <div className="bg-[#251F1A] p-6 text-[#FFFDF8] sm:p-8">
          <Badge className="bg-white/15 text-[#FFFDF8]" tone="neutral">
            Pesanan dibuat
          </Badge>
          <h1 className="mt-4 text-2xl font-bold tracking-tight sm:text-3xl">Pesanan berhasil dibuat</h1>
          <p className="mt-3 max-w-2xl text-sm leading-7 text-[#E3D2BC]">
            Simpan nomor order ini untuk komunikasi dengan toko. Pembayaran dilakukan manual dan akan ditinjau oleh penjual.
          </p>
        </div>

        <CardHeader>
          <CardTitle>Detail pesanan</CardTitle>
          <CardDescription>Toko akan menghubungi kamu untuk instruksi pembayaran dan konfirmasi pesanan.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-5">
          <div className="rounded-2xl border border-[#E3D2BC] bg-[#FFFDF8] p-4">
            <p className="text-xs font-semibold uppercase tracking-wide text-[#6F6256]">Nomor order</p>
            <p className="mt-1 text-xl font-bold text-[#251F1A]">{orderNumber ?? "Belum tersedia"}</p>
          </div>

          {grandTotal > 0 ? (
            <div className="flex items-center justify-between rounded-2xl border border-[#E3D2BC] p-4">
              <span className="text-sm text-[#6F6256]">Total final dari backend</span>
              <span className="text-lg font-bold text-[#B96E45]">{formatRupiah(grandTotal)}</span>
            </div>
          ) : null}

          <PaymentNotice>
            {paymentMessage ??
              "Silakan ikuti instruksi pembayaran dari toko. Setelah transfer, kirim konfirmasi pembayaran agar penjual bisa meninjau pesananmu."}
          </PaymentNotice>

          <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap">
            <Link
              className="inline-flex h-11 flex-1 items-center justify-center rounded-xl bg-[#251F1A] px-4 text-sm font-semibold text-[#FFFDF8] transition hover:bg-[#16110E]"
              href={`/s/${storeSlug}`}
            >
              Lanjut Belanja
            </Link>
            {orderNumber ? (
              <Link
                className="inline-flex h-11 flex-1 items-center justify-center rounded-xl border border-[#E3D2BC] bg-[#FFFDF8] px-4 text-sm font-semibold text-[#B96E45] transition hover:bg-[#F1E7D8]"
                href={`/s/${storeSlug}/orders/${encodeURIComponent(orderNumber)}/payment-confirmation`}
              >
                Konfirmasi pembayaran
              </Link>
            ) : null}
            <Link
              className="inline-flex h-11 flex-1 items-center justify-center rounded-xl border border-[#E3D2BC] bg-white px-4 text-sm font-semibold text-[#7C3F25] transition hover:bg-[#F8F1E7]"
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
