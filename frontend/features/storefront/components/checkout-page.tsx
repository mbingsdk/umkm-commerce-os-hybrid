"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMemo, useState, type ReactNode } from "react";
import { useForm, useWatch } from "react-hook-form";
import { EmptyState } from "@/components/feedback/empty-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { usePublicCourierZones } from "@/features/courier/hooks/use-courier-zones";
import { checkoutPublicStore } from "@/features/storefront/api/checkout.api";
import { useCartStore, getCartEstimatedSubtotal } from "@/features/storefront/cart.store";
import { checkoutSchema, type CheckoutFormValues } from "@/features/storefront/schemas/checkout.schema";
import { ApiError } from "@/lib/api/errors";
import { createIdempotencyKey } from "@/lib/api/idempotency";
import { formatRupiah } from "@/lib/format/money";

type CheckoutPageProps = {
  storeSlug: string;
};

export function CheckoutPage({ storeSlug }: CheckoutPageProps) {
  const router = useRouter();
  const cartStoreSlug = useCartStore((state) => state.storeSlug);
  const items = useCartStore((state) => (state.storeSlug === storeSlug ? state.items : []));
  const clearCart = useCartStore((state) => state.clearCart);
  const [submitError, setSubmitError] = useState<string>();
  const [submitPending, setSubmitPending] = useState(false);
  const subtotal = getCartEstimatedSubtotal(items);
  const courierZonesQuery = usePublicCourierZones(storeSlug);

  const form = useForm<CheckoutFormValues>({
    resolver: zodResolver(checkoutSchema),
    defaultValues: {
      customerName: "",
      customerPhone: "",
      customerEmail: "",
      recipientName: "",
      recipientPhone: "",
      address: "",
      city: "",
      province: "",
      postalCode: "",
      courierZoneId: "",
      customerNote: "",
      paymentMethod: "manual_transfer"
    }
  });
  const selectedCourierZoneId = useWatch({ control: form.control, name: "courierZoneId" });
  const selectedCourierZone = useMemo(
    () => courierZonesQuery.data?.find((zone) => zone.id === selectedCourierZoneId),
    [courierZonesQuery.data, selectedCourierZoneId]
  );
  const shippingEstimate = selectedCourierZone?.rate ?? 0;

  if (cartStoreSlug && cartStoreSlug !== storeSlug) {
    return (
      <main className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
        <EmptyState
          title="Keranjang bukan dari toko ini"
          description="Untuk saat ini, checkout publik hanya mendukung satu toko dalam satu transaksi."
          action={
            <Link
              className="inline-flex h-10 items-center justify-center rounded-xl bg-primary-600 px-4 text-sm font-semibold text-white transition hover:bg-primary-700"
              href={`/s/${cartStoreSlug}/cart`}
            >
              Lihat Keranjang Aktif
            </Link>
          }
        />
      </main>
    );
  }

  if (items.length === 0) {
    return (
      <main className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
        <EmptyState
          title="Belum ada produk untuk checkout"
          description="Tambahkan produk ke keranjang terlebih dahulu."
          action={
            <Link
              className="inline-flex h-10 items-center justify-center rounded-xl bg-primary-600 px-4 text-sm font-semibold text-white transition hover:bg-primary-700"
              href={`/s/${storeSlug}`}
            >
              Pilih Produk
            </Link>
          }
        />
      </main>
    );
  }

  async function onSubmit(values: CheckoutFormValues) {
    if (submitPending) {
      return;
    }

    setSubmitPending(true);
    setSubmitError(undefined);

    try {
      const result = await checkoutPublicStore(
        storeSlug,
        {
          items: items.map((item) => ({
            product_id: item.productId,
            quantity: item.quantity
          })),
          customer: {
            name: values.customerName,
            phone: values.customerPhone,
            email: values.customerEmail || undefined
          },
          shipping_address: {
            recipient_name: values.recipientName || values.customerName,
            recipient_phone: values.recipientPhone || values.customerPhone,
            address: values.address,
            city: values.city || undefined,
            province: values.province || undefined,
            postal_code: values.postalCode || undefined
          },
          shipping: values.courierZoneId ? { courier_zone_id: values.courierZoneId } : undefined,
          payment_method: "manual_transfer",
          customer_note: values.customerNote || undefined
        },
        createIdempotencyKey("checkout")
      );

      clearCart();

      const searchParams = new URLSearchParams({
        order_number: result.order_number,
        grand_total: String(result.totals.grand_total),
        payment_message: result.payment_instruction.message
      });
      router.push(`/s/${storeSlug}/orders/success?${searchParams.toString()}`);
    } catch (error) {
      if (error instanceof ApiError) {
        if (error.code === "INSUFFICIENT_STOCK") {
          setSubmitError("Stok produk tidak cukup. Silakan kembali ke keranjang dan kurangi jumlah item.");
          return;
        }
        if (error.code === "IDEMPOTENCY_CONFLICT") {
          setSubmitError("Permintaan checkout bentrok. Coba kirim ulang pesanan sekali lagi.");
          return;
        }
        if (error.code === "VALIDATION_ERROR") {
          setSubmitError(error.message || "Data checkout belum valid. Periksa kembali isianmu.");
          return;
        }
        setSubmitError(error.message);
        return;
      }

      setSubmitError("Checkout gagal diproses. Coba beberapa saat lagi.");
    } finally {
      setSubmitPending(false);
    }
  }

  const isCheckoutSubmitting = form.formState.isSubmitting || submitPending;

  return (
    <main className="mx-auto grid max-w-6xl gap-6 px-4 py-8 sm:px-6 lg:grid-cols-[minmax(0,1fr)_360px] lg:px-8">
      <form className="space-y-5" onSubmit={form.handleSubmit(onSubmit)}>
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-neutral-950">Checkout</h1>
          <p className="mt-1 text-sm text-neutral-500">Isi data pembeli dan alamat pengiriman.</p>
        </div>

        {submitError ? (
          <div className="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm leading-6 text-red-700">
            {submitError}
          </div>
        ) : null}

        <Card>
          <CardHeader>
            <CardTitle>Data pembeli</CardTitle>
            <CardDescription>Nomor HP dipakai toko untuk menghubungi pembeli.</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2">
            <FormField label="Nama pembeli" error={form.formState.errors.customerName?.message}>
              <Input hasError={!!form.formState.errors.customerName} {...form.register("customerName")} />
            </FormField>
            <FormField label="Nomor HP / WhatsApp" error={form.formState.errors.customerPhone?.message}>
              <Input
                hasError={!!form.formState.errors.customerPhone}
                placeholder="08123456789"
                {...form.register("customerPhone")}
              />
            </FormField>
            <FormField className="sm:col-span-2" label="Email opsional" error={form.formState.errors.customerEmail?.message}>
              <Input
                hasError={!!form.formState.errors.customerEmail}
                placeholder="nama@email.com"
                type="email"
                {...form.register("customerEmail")}
              />
            </FormField>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Alamat pengiriman</CardTitle>
            <CardDescription>Alamat dipakai toko untuk mengirim dan menghubungi penerima.</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2">
            <FormField label="Nama penerima opsional" error={form.formState.errors.recipientName?.message}>
              <Input placeholder="Sama dengan pembeli jika kosong" {...form.register("recipientName")} />
            </FormField>
            <FormField label="HP penerima opsional" error={form.formState.errors.recipientPhone?.message}>
              <Input placeholder="Sama dengan pembeli jika kosong" {...form.register("recipientPhone")} />
            </FormField>
            <FormField className="sm:col-span-2" label="Alamat lengkap" error={form.formState.errors.address?.message}>
              <Input hasError={!!form.formState.errors.address} {...form.register("address")} />
            </FormField>
            <FormField label="Kota" error={form.formState.errors.city?.message}>
              <Input {...form.register("city")} />
            </FormField>
            <FormField label="Provinsi" error={form.formState.errors.province?.message}>
              <Input {...form.register("province")} />
            </FormField>
            <FormField label="Kode pos" error={form.formState.errors.postalCode?.message}>
              <Input {...form.register("postalCode")} />
            </FormField>
            <FormField className="sm:col-span-2" label="Catatan untuk toko" error={form.formState.errors.customerNote?.message}>
              <Input placeholder="Contoh: kirim sore, hubungi sebelum datang" {...form.register("customerNote")} />
            </FormField>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Zona pengiriman</CardTitle>
            <CardDescription>
              Pilih zona kurir toko. Estimasi di frontend hanya bantuan; total final tetap dihitung backend.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {courierZonesQuery.isPending ? (
              <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-4 text-sm text-neutral-600">
                Memuat pilihan zona pengiriman...
              </div>
            ) : courierZonesQuery.isError ? (
              <div className="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm leading-6 text-amber-800">
                Zona pengiriman belum bisa dimuat. Kamu tetap bisa checkout dan toko akan mengonfirmasi ongkir manual.
              </div>
            ) : null}

            <FormField label="Pilih zona pengiriman" error={form.formState.errors.courierZoneId?.message}>
              <select
                className="h-10 w-full rounded-xl border border-neutral-200 bg-white px-3 text-sm outline-none transition focus:border-primary-500 focus:ring-4 focus:ring-primary-100"
                {...form.register("courierZoneId")}
              >
                <option value="">Konfirmasi ongkir manual oleh toko</option>
                {(courierZonesQuery.data ?? []).map((zone) => (
                  <option key={zone.id} value={zone.id}>
                    {zone.name} - {formatRupiah(zone.rate)}
                  </option>
                ))}
              </select>
            </FormField>

            {selectedCourierZone ? (
              <div className="rounded-2xl border border-primary-100 bg-primary-50 p-4 text-sm leading-6 text-primary-900">
                Ongkir estimasi untuk zona <strong>{selectedCourierZone.name}</strong>:{" "}
                <strong>{formatRupiah(selectedCourierZone.rate)}</strong>.
              </div>
            ) : (
              <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-4 text-sm leading-6 text-neutral-600">
                Jika zona belum dipilih, pesanan dibuat dengan ongkir Rp0 dan toko perlu mengonfirmasi biaya kirim manual.
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Metode pembayaran</CardTitle>
            <CardDescription>Pembayaran dilakukan manual oleh toko untuk rilis pilot ini.</CardDescription>
          </CardHeader>
          <CardContent>
            <label className="flex items-start gap-3 rounded-2xl border border-primary-200 bg-primary-50 p-4">
              <input
                className="mt-1"
                type="radio"
                value="manual_transfer"
                {...form.register("paymentMethod")}
                defaultChecked
              />
              <span>
                <span className="block text-sm font-semibold text-primary-950">Transfer manual</span>
                <span className="mt-1 block text-sm leading-6 text-primary-900">
                  Payment gateway belum terhubung. Toko akan mengirim instruksi pembayaran dan memeriksa pembayaran dari dashboard.
                </span>
              </span>
            </label>
          </CardContent>
        </Card>

        <div className="lg:hidden">
          <CheckoutSummary
            isLoading={isCheckoutSubmitting}
            items={items}
            onSubmit={form.handleSubmit(onSubmit)}
            selectedCourierZoneName={selectedCourierZone?.name}
            shippingEstimate={shippingEstimate}
            storeSlug={storeSlug}
            subtotal={subtotal}
          />
        </div>
      </form>

      <aside className="hidden lg:block lg:sticky lg:top-6 lg:self-start">
        <CheckoutSummary
          isLoading={isCheckoutSubmitting}
          items={items}
          onSubmit={form.handleSubmit(onSubmit)}
          selectedCourierZoneName={selectedCourierZone?.name}
          shippingEstimate={shippingEstimate}
          storeSlug={storeSlug}
          subtotal={subtotal}
        />
      </aside>
    </main>
  );
}

function CheckoutSummary({
  isLoading,
  items,
  onSubmit,
  selectedCourierZoneName,
  shippingEstimate,
  storeSlug,
  subtotal
}: {
  isLoading: boolean;
  items: Array<{
    productId: string;
    name: string;
    displayPrice: number;
    quantity: number;
  }>;
  onSubmit: () => void;
  selectedCourierZoneName?: string;
  shippingEstimate: number;
  storeSlug: string;
  subtotal: number;
}) {
  const itemCount = items.reduce((total, item) => total + item.quantity, 0);
  const estimatedTotal = subtotal + shippingEstimate;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Ringkasan checkout</CardTitle>
        <CardDescription>Harga final dihitung oleh backend saat pesanan dibuat.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-3">
          {items.map((item) => (
            <div key={item.productId} className="flex items-start justify-between gap-3 text-sm">
              <div>
                <p className="font-medium text-neutral-950">{item.name}</p>
                <p className="text-neutral-500">
                  {item.quantity} × {formatRupiah(item.displayPrice)}
                </p>
              </div>
              <p className="font-semibold text-neutral-950">{formatRupiah(item.quantity * item.displayPrice)}</p>
            </div>
          ))}
        </div>

        <div className="border-t border-neutral-100 pt-4">
          <div className="flex items-center justify-between text-sm">
            <span className="text-neutral-500">{itemCount} item</span>
            <span className="font-semibold text-neutral-950">{formatRupiah(subtotal)}</span>
          </div>
          <div className="mt-2 flex items-center justify-between text-sm">
            <span className="text-neutral-500">
              Ongkir{selectedCourierZoneName ? ` (${selectedCourierZoneName})` : " manual"}
            </span>
            <span className="font-semibold text-neutral-950">{formatRupiah(shippingEstimate)}</span>
          </div>
          <div className="mt-4 flex items-center justify-between text-base">
            <span className="font-semibold text-neutral-950">Estimasi total</span>
            <span className="text-xl font-bold text-primary-700">{formatRupiah(estimatedTotal)}</span>
          </div>
        </div>

        <Button className="w-full" isLoading={isLoading} disabled={isLoading || items.length === 0} onClick={onSubmit} size="lg" type="button">
          Buat Pesanan
        </Button>
        <Link className="block text-center text-sm font-semibold text-primary-700" href={`/s/${storeSlug}/cart`}>
          Kembali ke keranjang
        </Link>
      </CardContent>
    </Card>
  );
}

function FormField({
  label,
  error,
  className,
  children
}: {
  label: string;
  error?: string;
  className?: string;
  children: ReactNode;
}) {
  return (
    <label className={className ? `space-y-1 ${className}` : "space-y-1"}>
      <span className="text-sm font-medium text-neutral-700">{label}</span>
      {children}
      {error ? <span className="block text-xs font-medium text-red-600">{error}</span> : null}
    </label>
  );
}
