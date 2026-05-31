"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import Link from "next/link";
import { type ReactNode } from "react";
import { useForm, useWatch } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { submitPublicPaymentConfirmation } from "@/features/storefront/api/payment-confirmation.api";
import { ApiError } from "@/lib/api/errors";
import { createIdempotencyKey } from "@/lib/api/idempotency";
import { formatRupiah } from "@/lib/format/money";
import { useToastStore } from "@/lib/stores/toast.store";

const paymentConfirmationSchema = z.object({
  customerPhone: z.string().trim().min(8, "Nomor HP wajib diisi.").max(32, "Nomor HP terlalu panjang."),
  payerName: z.string().trim().min(2, "Nama pengirim wajib diisi.").max(120, "Nama pengirim terlalu panjang."),
  bankName: z.string().trim().min(2, "Nama bank wajib diisi.").max(120, "Nama bank terlalu panjang."),
  transferAmount: z
    .number({ error: "Nominal transfer wajib diisi." })
    .int("Nominal harus angka bulat.")
    .positive("Nominal transfer wajib lebih dari Rp0."),
  transferDate: z
    .string()
    .trim()
    .min(1, "Tanggal transfer wajib diisi.")
    .regex(/^\d{4}-\d{2}-\d{2}$/, "Tanggal transfer harus format YYYY-MM-DD."),
  proofUrl: z.string().trim().max(500, "Referensi bukti transfer maksimal 500 karakter.").optional(),
  note: z.string().trim().max(500, "Catatan maksimal 500 karakter.").optional()
});

type PaymentConfirmationFormValues = z.infer<typeof paymentConfirmationSchema>;

type PaymentConfirmationPageProps = {
  storeSlug: string;
  orderNumber: string;
};

export function PaymentConfirmationPage({ storeSlug, orderNumber }: PaymentConfirmationPageProps) {
  const pushToast = useToastStore((state) => state.pushToast);
  const form = useForm<PaymentConfirmationFormValues>({
    resolver: zodResolver(paymentConfirmationSchema),
    defaultValues: {
      customerPhone: "",
      payerName: "",
      bankName: "",
      transferAmount: 0,
      transferDate: todayInputValue(),
      proofUrl: "",
      note: ""
    }
  });

  const confirmationMutation = useMutation({
    mutationFn: (values: PaymentConfirmationFormValues) =>
      submitPublicPaymentConfirmation(
        storeSlug,
        orderNumber,
        {
          customer_phone: values.customerPhone.trim(),
          payer_name: values.payerName.trim(),
          bank_name: values.bankName.trim(),
          transfer_amount: values.transferAmount,
          transfer_date: values.transferDate,
          proof_url: values.proofUrl?.trim() || "",
          note: values.note?.trim() || ""
        },
        createIdempotencyKey("payment")
      ),
    onSuccess: () => {
      pushToast({
        tone: "success",
        title: "Konfirmasi pembayaran terkirim",
        description: "Penjual akan memeriksa pembayaran dari dashboard toko."
      });
    },
    onError: (error) => {
      pushToast({
        tone: "error",
        title: "Konfirmasi pembayaran gagal",
        description: paymentConfirmationErrorMessage(error)
      });
    }
  });
  const submittedAmount = useWatch({ control: form.control, name: "transferAmount" });

  function handleSubmit(values: PaymentConfirmationFormValues) {
    if (confirmationMutation.isPending) {
      return;
    }

    confirmationMutation.mutate(values);
  }

  if (confirmationMutation.isSuccess) {
    const result = confirmationMutation.data;

    return (
      <main className="mx-auto max-w-3xl px-4 py-10 sm:px-6 lg:px-8">
        <Card className="overflow-hidden border-[#eadfce] bg-white/90 shadow-[0_18px_55px_rgba(89,63,38,0.1)]">
          <div className="bg-[#2f2923] p-6 text-[#fffaf2] sm:p-8">
            <p className="inline-flex rounded-full bg-white/15 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-white">
              Menunggu review toko
            </p>
            <h1 className="mt-4 text-2xl font-bold tracking-tight sm:text-3xl">Konfirmasi pembayaran terkirim</h1>
            <p className="mt-3 max-w-2xl text-sm leading-7 text-[#eadfce]">
              {result.message || "Bukti pembayaran sudah diterima dan akan dicek oleh penjual."}
            </p>
          </div>
          <CardContent className="space-y-5">
            <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf2] p-4">
              <p className="text-xs font-semibold uppercase tracking-wide text-[#7b6a58]">Nomor order</p>
              <p className="mt-1 text-xl font-bold text-[#241c16]">{result.order_number}</p>
            </div>
            <div className="rounded-2xl border border-[#d8c7ad] bg-[#fff6df] p-4 text-sm leading-6 text-[#4b3a29]">
              Pembayaran belum otomatis berstatus lunas. Penjual tetap perlu memeriksa dan menyetujui konfirmasi ini dari
              dashboard toko.
            </div>
            <div className="flex flex-col gap-3 sm:flex-row">
              <Link
                className="inline-flex h-11 flex-1 items-center justify-center rounded-xl bg-[#2f2923] px-4 text-sm font-semibold text-[#fffaf2] transition hover:bg-[#1f1a16]"
                href={`/s/${storeSlug}/track-order`}
              >
                Lacak pesanan
              </Link>
              <Link
                className="inline-flex h-11 flex-1 items-center justify-center rounded-xl border border-[#d8c7ad] bg-white px-4 text-sm font-semibold text-[#3b2f24] transition hover:bg-[#f7f1e8]"
                href={`/s/${storeSlug}`}
              >
                Kembali ke toko
              </Link>
            </div>
          </CardContent>
        </Card>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-6">
        <Link className="text-sm font-semibold text-[#7a4f2f] hover:text-[#3b2f24]" href={`/s/${storeSlug}`}>
          ← Kembali ke toko
        </Link>
        <h1 className="mt-3 text-2xl font-bold tracking-tight text-[#241c16]">Konfirmasi pembayaran</h1>
        <p className="mt-1 text-sm leading-6 text-[#7b6a58]">
          Isi data transfer untuk order <span className="font-semibold text-neutral-800">{orderNumber}</span>. Halaman ini
          publik dan tidak membutuhkan login.
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
        <form className="space-y-5" onSubmit={form.handleSubmit(handleSubmit)}>
          {confirmationMutation.isError ? (
            <div className="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm leading-6 text-red-700">
              {paymentConfirmationErrorMessage(confirmationMutation.error)}
            </div>
          ) : null}

          <Card className="border-[#eadfce] bg-white/90 shadow-[0_12px_35px_rgba(89,63,38,0.06)]">
            <CardHeader>
              <CardTitle>Verifikasi pesanan</CardTitle>
              <CardDescription>
                Nomor HP harus sama dengan nomor yang dipakai saat checkout agar toko bisa memverifikasi pesanan.
              </CardDescription>
            </CardHeader>
            <CardContent className="grid gap-4 sm:grid-cols-2">
              <FormField label="Nomor HP / WhatsApp checkout" error={form.formState.errors.customerPhone?.message}>
                <Input
                  autoComplete="tel"
                  hasError={!!form.formState.errors.customerPhone}
                  inputMode="tel"
                  placeholder="08123456789"
                  {...form.register("customerPhone")}
                />
              </FormField>
              <FormField label="Nama pengirim rekening" error={form.formState.errors.payerName?.message}>
                <Input
                  autoComplete="name"
                  hasError={!!form.formState.errors.payerName}
                  placeholder="Nama sesuai bukti transfer"
                  {...form.register("payerName")}
                />
              </FormField>
            </CardContent>
          </Card>

          <Card className="border-[#eadfce] bg-white/90 shadow-[0_12px_35px_rgba(89,63,38,0.06)]">
            <CardHeader>
              <CardTitle>Detail transfer</CardTitle>
              <CardDescription>
                Data ini hanya mengirim konfirmasi ke toko. Status lunas tetap diputuskan oleh penjual.
              </CardDescription>
            </CardHeader>
            <CardContent className="grid gap-4 sm:grid-cols-2">
              <FormField label="Nama bank / e-wallet" error={form.formState.errors.bankName?.message}>
                <Input hasError={!!form.formState.errors.bankName} placeholder="BCA, BRI, Mandiri, DANA..." {...form.register("bankName")} />
              </FormField>
              <FormField label="Nominal transfer" error={form.formState.errors.transferAmount?.message}>
                <Input
                  hasError={!!form.formState.errors.transferAmount}
                  inputMode="numeric"
                  min={1}
                  placeholder="50000"
                  type="number"
                  {...form.register("transferAmount", {
                    setValueAs: (value) => (value === "" ? 0 : Number(value))
                  })}
                />
              </FormField>
              <FormField label="Tanggal transfer" error={form.formState.errors.transferDate?.message}>
                <Input hasError={!!form.formState.errors.transferDate} type="date" {...form.register("transferDate")} />
              </FormField>
              <FormField label="Link/referensi bukti transfer opsional" error={form.formState.errors.proofUrl?.message}>
                <Input
                  hasError={!!form.formState.errors.proofUrl}
                  placeholder="Link gambar atau nomor referensi transfer"
                  {...form.register("proofUrl")}
                />
              </FormField>
              <FormField className="sm:col-span-2" label="Catatan opsional" error={form.formState.errors.note?.message}>
                <textarea
                  className="min-h-24 w-full rounded-xl border border-neutral-300 bg-white px-3 py-2 text-sm text-neutral-950 shadow-sm outline-none transition placeholder:text-neutral-400 focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
                  placeholder="Contoh: transfer dari rekening atas nama keluarga."
                  {...form.register("note")}
                />
              </FormField>
            </CardContent>
          </Card>

          <Button
            className="w-full bg-[#2f2923] hover:bg-[#1f1a16] sm:w-auto"
            disabled={confirmationMutation.isPending}
            isLoading={confirmationMutation.isPending}
            size="lg"
            type="submit"
          >
            Kirim konfirmasi pembayaran
          </Button>
        </form>

        <aside className="space-y-4 lg:sticky lg:top-6 lg:self-start">
          <Card className="border-[#eadfce] bg-white/90 shadow-[0_16px_45px_rgba(89,63,38,0.08)]">
            <CardHeader>
              <CardTitle>Ringkasan</CardTitle>
              <CardDescription>Pastikan nominal sesuai transfer yang benar-benar dikirim.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 text-sm">
              <InfoRow label="Order" value={orderNumber} />
              <InfoRow
                label="Nominal diisi"
                value={submittedAmount && submittedAmount > 0 ? formatRupiah(submittedAmount) : "Belum diisi"}
              />
              <div className="rounded-2xl border border-[#d8c7ad] bg-[#fff6df] p-4 leading-6 text-[#4b3a29]">
                Upload bukti transfer belum tersedia di halaman publik. Jika punya bukti berupa link atau nomor referensi, isi pada kolom opsional.
              </div>
            </CardContent>
          </Card>
        </aside>
      </div>
    </main>
  );
}

function paymentConfirmationErrorMessage(error: unknown) {
  if (error instanceof ApiError) {
    if (error.code === "NOT_FOUND") {
      return "Order tidak ditemukan. Pastikan nomor order dan nomor HP checkout sudah benar.";
    }
    if (error.code === "IDEMPOTENCY_CONFLICT") {
      return "Konfirmasi ini bentrok dengan pengiriman sebelumnya. Muat ulang halaman lalu coba sekali lagi.";
    }
    if (error.code === "VALIDATION_ERROR") {
      return error.message || "Data konfirmasi belum valid. Periksa kembali isianmu.";
    }
    return error.message;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return "Konfirmasi pembayaran gagal dikirim. Coba beberapa saat lagi.";
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

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">{label}</p>
      <p className="mt-1 break-words font-semibold text-[#241c16]">{value}</p>
    </div>
  );
}

function todayInputValue() {
  const today = new Date();
  today.setMinutes(today.getMinutes() - today.getTimezoneOffset());
  return today.toISOString().slice(0, 10);
}
