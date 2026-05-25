"use client";

import Link from "next/link";
import { useState, type FormEvent } from "react";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { OrderStatusBadge, PaymentStatusBadge } from "@/features/orders/components/status-badges";
import { shipmentStatusLabel } from "@/features/shipments/constants";
import { ShipmentStatusBadge } from "@/features/shipments/components/shipment-status-badge";
import { usePublicOrderTracking } from "@/features/shipments/hooks/use-shipments";
import type { PublicTrackingResult, PublicTrackingTimelineItem } from "@/features/shipments/types";
import { formatDateTime } from "@/lib/format/date";
import { formatRupiah } from "@/lib/format/money";

type TrackOrderPageProps = {
  storeSlug: string;
};

export function TrackOrderPage({ storeSlug }: TrackOrderPageProps) {
  const [orderNumber, setOrderNumber] = useState("");
  const [phone, setPhone] = useState("");
  const [submitted, setSubmitted] = useState<{ orderNumber: string; phone: string } | null>(null);
  const [localError, setLocalError] = useState<string>();
  const trackingQuery = usePublicOrderTracking(
    storeSlug,
    submitted?.orderNumber ?? "",
    submitted?.phone ?? "",
    !!submitted
  );

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (trackingQuery.isFetching) {
      return;
    }

    const nextOrderNumber = orderNumber.trim();
    const nextPhone = phone.trim();

    if (!nextOrderNumber || !nextPhone) {
      setLocalError("Nomor pesanan dan nomor HP wajib diisi.");
      return;
    }

    setLocalError(undefined);
    setSubmitted({ orderNumber: nextOrderNumber, phone: nextPhone });
  }

  return (
    <main className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-6">
        <Link className="text-sm font-semibold text-primary-700 hover:text-primary-800" href={`/s/${storeSlug}`}>
          ← Kembali ke toko
        </Link>
        <h1 className="mt-3 text-2xl font-bold tracking-tight text-neutral-950">Lacak pesanan</h1>
        <p className="mt-1 text-sm leading-6 text-neutral-500">
          Masukkan nomor pesanan dan nomor HP yang dipakai saat checkout. Tracking publik tidak membutuhkan login.
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-[360px_minmax(0,1fr)]">
        <Card>
          <CardHeader>
            <CardTitle>Cek status</CardTitle>
            <CardDescription>Nomor HP dipakai sebagai verifikasi agar data pesanan tetap aman.</CardDescription>
          </CardHeader>
          <CardContent>
            <form className="space-y-4" onSubmit={handleSubmit}>
              {localError ? (
                <div className="rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700">{localError}</div>
              ) : null}

              <label className="space-y-1">
                <span className="text-sm font-medium text-neutral-700">Nomor pesanan</span>
                <Input
                  value={orderNumber}
                  onChange={(event) => setOrderNumber(event.target.value)}
                  placeholder="ORD-20260520-..."
                />
              </label>

              <label className="space-y-1">
                <span className="text-sm font-medium text-neutral-700">Nomor HP / WhatsApp</span>
                <Input value={phone} onChange={(event) => setPhone(event.target.value)} placeholder="08123456789" />
              </label>

              <Button className="w-full" type="submit" isLoading={trackingQuery.isFetching} disabled={trackingQuery.isFetching}>
                Lacak pesanan
              </Button>
            </form>
          </CardContent>
        </Card>

        <section className="space-y-6">
          {!submitted ? (
            <EmptyState
              title="Masukkan data pesanan"
              description="Hasil tracking akan tampil di sini setelah nomor pesanan dan nomor HP cocok."
            />
          ) : trackingQuery.isPending ? (
            <LoadingState lines={4} />
          ) : trackingQuery.isError ? (
            <ErrorState
              title="Pesanan tidak ditemukan"
              description="Pastikan nomor pesanan dan nomor HP sama dengan data checkout."
              onRetry={() => void trackingQuery.refetch()}
            />
          ) : (
            <TrackingResult data={trackingQuery.data} />
          )}
        </section>
      </div>
    </main>
  );
}

function TrackingResult({ data }: { data: PublicTrackingResult }) {
  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>{data.orderNumber}</CardTitle>
          <CardDescription>
            Halo {data.customerName}. Berikut ringkasan aman status pesananmu.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-wrap gap-2">
            <OrderStatusBadge status={data.status} />
            <PaymentStatusBadge status={data.paymentStatus} />
            {data.shipmentStatus ? <ShipmentStatusBadge status={data.shipmentStatus} /> : null}
          </div>

          <div className="grid gap-3 sm:grid-cols-3">
            <MiniMetric label="Subtotal" value={formatRupiah(data.totals.subtotal)} />
            <MiniMetric label="Ongkir" value={formatRupiah(data.totals.shippingCost)} />
            <MiniMetric label="Total" value={formatRupiah(data.totals.grandTotal)} />
          </div>
        </CardContent>
      </Card>

      {data.shipment ? (
        <Card>
          <CardHeader>
            <CardTitle>Info pengiriman</CardTitle>
            <CardDescription>Data publik yang aman untuk customer.</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-3 text-sm sm:grid-cols-2">
            <InfoRow label="Kurir" value={data.shipment.courierName || data.shipment.courierType} />
            <InfoRow label="Nomor resi" value={data.shipment.trackingNumber || "Belum tersedia"} />
            <InfoRow label="Status" value={shipmentStatusLabel(data.shipment.status)} />
            <InfoRow
              label="Dikirim"
              value={data.shipment.shippedAt ? formatDateTime(data.shipment.shippedAt) : "Belum dikirim"}
            />
          </CardContent>
        </Card>
      ) : (
        <EmptyState
          title="Pengiriman belum dibuat"
          description="Toko belum membuat shipment untuk pesanan ini. Silakan cek lagi nanti atau hubungi toko."
        />
      )}

      <Card>
        <CardHeader>
          <CardTitle>Item pesanan</CardTitle>
          <CardDescription>Harga berasal dari snapshot saat checkout.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {data.items.map((item) => (
            <div key={`${item.productName}-${item.quantity}-${item.subtotal}`} className="flex items-start justify-between gap-4 text-sm">
              <div>
                <p className="font-semibold text-neutral-950">{item.productName}</p>
                <p className="mt-1 text-neutral-500">
                  {item.quantity} × {formatRupiah(item.unitPrice)}
                </p>
              </div>
              <p className="font-semibold text-neutral-950">{formatRupiah(item.subtotal)}</p>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Timeline</CardTitle>
          <CardDescription>Status pengiriman yang sudah tercatat.</CardDescription>
        </CardHeader>
        <CardContent>
          <PublicTimeline timeline={data.timeline} />
        </CardContent>
      </Card>
    </>
  );
}

function MiniMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl bg-neutral-50 p-4">
      <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">{label}</p>
      <p className="mt-2 font-semibold text-neutral-950">{value}</p>
    </div>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">{label}</p>
      <p className="mt-1 text-neutral-700">{value}</p>
    </div>
  );
}

function PublicTimeline({ timeline }: { timeline: PublicTrackingTimelineItem[] }) {
  if (timeline.length === 0) {
    return <p className="text-sm text-neutral-500">Belum ada update pengiriman.</p>;
  }

  return (
    <div className="space-y-4">
      {timeline.map((item) => (
        <div key={`${item.status}-${item.createdAt}`} className="border-l-2 border-primary-200 pl-4">
          <p className="text-sm font-semibold text-neutral-950">{shipmentStatusLabel(item.status)}</p>
          <p className="mt-1 text-xs text-neutral-500">{formatDateTime(item.createdAt)}</p>
          {item.note ? <p className="mt-2 text-sm leading-6 text-neutral-600">{item.note}</p> : null}
        </div>
      ))}
    </div>
  );
}
