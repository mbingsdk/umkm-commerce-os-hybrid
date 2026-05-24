"use client";

import Link from "next/link";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { closePOSSession, createPOSTransaction } from "@/features/pos/api/pos.api";
import { CloseSessionDialog } from "@/features/pos/components/close-session-dialog";
import { ReceiptDialog } from "@/features/pos/components/receipt-dialog";
import { usePOSProducts } from "@/features/pos/hooks/use-pos";
import { usePOSStore } from "@/features/pos/pos.store";
import type { POSCartLine, POSProduct, POSSession, POSTransaction } from "@/features/pos/types";
import { createIdempotencyKey } from "@/lib/api/idempotency";
import { queryKeys } from "@/lib/api/query-keys";
import { formatDateTime } from "@/lib/format/date";
import { formatRupiah } from "@/lib/format/money";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";
import { useToastStore } from "@/lib/stores/toast.store";

type POSRegisterScreenProps = {
  session: POSSession;
};

export function POSRegisterScreen({ session }: POSRegisterScreenProps) {
  const queryClient = useQueryClient();
  const pushToast = useToastStore((state) => state.pushToast);
  const tenantId = useTenantStore((state) => state.selectedTenantId);
  const userPermissions = useTenantStore((state) => state.permissions);
  const canReadProducts = userPermissions.includes(permissions.posReadProduct);
  const canCreateTransaction = userPermissions.includes(permissions.posCreateTransaction);
  const canCloseSession = userPermissions.includes(permissions.posCloseSession);
  const [query, setQuery] = useState("");
  const [category, setCategory] = useState("");
  const [closeOpen, setCloseOpen] = useState(false);
  const [receipt, setReceipt] = useState<{ transaction: POSTransaction; items: POSCartLine[] } | null>(null);
  const items = usePOSStore((state) => state.items);
  const paymentMethod = usePOSStore((state) => state.paymentMethod);
  const amountPaid = usePOSStore((state) => state.amountPaid);
  const selectedProduct = usePOSStore((state) => state.selectedProduct);
  const setSelectedProduct = usePOSStore((state) => state.setSelectedProduct);
  const setPaymentMethod = usePOSStore((state) => state.setPaymentMethod);
  const setAmountPaid = usePOSStore((state) => state.setAmountPaid);
  const addProduct = usePOSStore((state) => state.addProduct);
  const updateQuantity = usePOSStore((state) => state.updateQuantity);
  const removeItem = usePOSStore((state) => state.removeItem);
  const clearCart = usePOSStore((state) => state.clearCart);
  const productsQuery = usePOSProducts({ query: query.trim(), limit: 100 }, canReadProducts);

  const categories = useMemo(() => {
    const values = new Map<string, string>();
    for (const product of productsQuery.data ?? []) {
      if (product.categoryId && product.categoryName) {
        values.set(product.categoryId, product.categoryName);
      }
    }
    return Array.from(values.entries()).map(([id, name]) => ({ id, name }));
  }, [productsQuery.data]);

  const products = useMemo(
    () => (productsQuery.data ?? []).filter((product) => !category || product.categoryId === category),
    [category, productsQuery.data]
  );
  const subtotalEstimate = items.reduce((total, item) => total + item.price * item.quantity, 0);
  const changeEstimate = paymentMethod === "cash" ? Math.max(0, amountPaid - subtotalEstimate) : 0;
  const qrisMismatch = paymentMethod === "qris_manual" && amountPaid !== subtotalEstimate;
  const cashUnderpaid = paymentMethod === "cash" && amountPaid < subtotalEstimate;
  const cartEmpty = items.length === 0;

  const createMutation = useMutation({
    mutationFn: () =>
      createPOSTransaction({
        sessionId: session.id,
        items: items.map((item) => ({ productId: item.productId, quantity: item.quantity })),
        paymentMethod,
        amountPaid,
        note: "Transaksi dari POS dashboard",
        idempotencyKey: createIdempotencyKey("pos")
      }),
    onSuccess: async (transaction) => {
      setReceipt({ transaction, items: [...items] });
      clearCart();
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.posProducts(tenantId) }),
        queryClient.invalidateQueries({ queryKey: queryKeys.posTransactions(tenantId) })
      ]);
      pushToast({ tone: "success", title: "Transaksi POS berhasil" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Transaksi POS gagal", description: error.message });
    }
  });

  const closeMutation = useMutation({
    mutationFn: (values: { closingCashAmount: number; note?: string }) => closePOSSession(session.id, values),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.posTransactions(tenantId) });
      pushToast({ tone: "success", title: "Sesi kasir berhasil ditutup" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Sesi gagal ditutup", description: error.message });
    }
  });

  function handleProductClick(product: POSProduct) {
    setSelectedProduct(product);
    if (product.availableStock > 0) {
      addProduct(product);
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 rounded-2xl border border-neutral-200 bg-white p-4 shadow-soft lg:flex-row lg:items-center lg:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">Sesi aktif</p>
          <div className="mt-1 flex flex-wrap items-center gap-2">
            <h1 className="text-xl font-semibold text-neutral-950">{session.sessionNumber}</h1>
            <Badge tone="success">Open</Badge>
          </div>
          <p className="mt-1 text-sm text-neutral-500">
            Dibuka {formatDateTime(session.openedAt)} • Kas awal {formatRupiah(session.openingCash)}
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button type="button" variant="outline">
            <Link href="/dashboard/pos/history">Riwayat POS</Link>
          </Button>
          <Button type="button" variant="danger" disabled={!canCloseSession} onClick={() => setCloseOpen(true)}>
            Tutup sesi
          </Button>
        </div>
      </div>

      <div className="grid min-h-[680px] gap-4 xl:grid-cols-[1fr_420px]">
        <section className="space-y-4 rounded-2xl border border-neutral-200 bg-white p-4 shadow-soft">
          <div className="grid gap-3 lg:grid-cols-[1fr_220px]">
            <Input
              aria-label="Cari produk POS"
              autoFocus
              className="h-12 text-base"
              enterKeyHint="search"
              placeholder="Cari produk atau SKU..."
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
            <select
              className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
              value={category}
              onChange={(event) => setCategory(event.target.value)}
            >
              <option value="">Semua kategori</option>
              {categories.map((item) => (
                <option key={item.id} value={item.id}>
                  {item.name}
                </option>
              ))}
            </select>
          </div>

          {!canReadProducts ? (
            <EmptyState
              title="Akses produk POS belum tersedia"
              description="Role aktifmu belum memiliki izin untuk melihat produk POS."
            />
          ) : productsQuery.isPending ? (
            <LoadingState lines={5} />
          ) : productsQuery.isError ? (
            <ErrorState
              title="Produk POS gagal dimuat"
              description="Coba muat ulang daftar produk."
              onRetry={() => void productsQuery.refetch()}
            />
          ) : products.length === 0 ? (
            <EmptyState title="Produk tidak ditemukan" description="Coba kata kunci lain atau ubah filter kategori." />
          ) : (
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
              {products.map((product) => (
                <ProductTile
                  key={product.productId}
                  product={product}
                  selected={selectedProduct?.productId === product.productId}
                  onClick={() => handleProductClick(product)}
                />
              ))}
            </div>
          )}
        </section>

        <aside className="flex flex-col rounded-2xl border border-neutral-200 bg-white shadow-soft">
          <div className="border-b border-neutral-100 p-5">
            <h2 className="text-lg font-semibold text-neutral-950">Cart POS</h2>
            <p className="mt-1 text-sm text-neutral-500">Subtotal frontend hanya estimasi; backend tetap menghitung final.</p>
          </div>

          <div className="min-h-0 flex-1 space-y-3 overflow-y-auto p-5">
            {items.length === 0 ? (
              <EmptyState
                title="Cart masih kosong"
                description="Pilih produk dari daftar untuk mulai transaksi POS."
              />
            ) : (
              items.map((item) => (
                <CartLine
                  key={item.productId}
                  item={item}
                  onQuantityChange={(quantity) => updateQuantity(item.productId, quantity)}
                  onRemove={() => removeItem(item.productId)}
                />
              ))
            )}
          </div>

          <div className="space-y-4 border-t border-neutral-100 p-5">
            <div className="space-y-2 text-sm">
              <MoneyRow label="Subtotal estimasi" value={subtotalEstimate} strong />
              <MoneyRow label="Kembalian estimasi" value={changeEstimate} />
            </div>

            <label className="block text-sm font-medium text-neutral-800">
              Metode pembayaran
              <select
                className="mt-2 h-10 w-full rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
                value={paymentMethod}
                onChange={(event) => {
                  const method = event.target.value as "cash" | "qris_manual";
                  setPaymentMethod(method);
                  if (method === "qris_manual") {
                    setAmountPaid(subtotalEstimate);
                  }
                }}
              >
                <option value="cash">Tunai</option>
                <option value="qris_manual">QRIS manual</option>
              </select>
            </label>

            <label className="block text-sm font-medium text-neutral-800">
              Jumlah dibayar
              <Input
                className="mt-2"
                type="number"
                min={0}
                step={1}
                value={amountPaid}
                onChange={(event) => setAmountPaid(Number(event.target.value))}
              />
            </label>

            <div className="grid grid-cols-2 gap-2">
              <Button type="button" variant="outline" onClick={() => setAmountPaid(subtotalEstimate)}>
                Bayar pas
              </Button>
              <Button type="button" variant="outline" onClick={clearCart} disabled={cartEmpty}>
                Kosongkan
              </Button>
            </div>

            {qrisMismatch ? (
              <p className="rounded-xl bg-amber-50 p-3 text-sm text-amber-800">
                Untuk QRIS manual, jumlah dibayar harus sama dengan total estimasi.
              </p>
            ) : null}

            <Button
              className="w-full"
              type="button"
              isLoading={createMutation.isPending}
              disabled={createMutation.isPending || !canCreateTransaction || cartEmpty || cashUnderpaid || qrisMismatch}
              onClick={() => {
                if (createMutation.isPending || !canCreateTransaction || cartEmpty || cashUnderpaid || qrisMismatch) {
                  return;
                }

                createMutation.mutate();
              }}
            >
              Proses pembayaran
            </Button>
          </div>
        </aside>
      </div>

      <ReceiptDialog
        open={!!receipt}
        transaction={receipt?.transaction ?? null}
        items={receipt?.items ?? []}
        onClose={() => setReceipt(null)}
      />

      <CloseSessionDialog
        open={closeOpen}
        session={session}
        closedSession={closeMutation.data ?? null}
        isSubmitting={closeMutation.isPending}
        error={closeMutation.isError ? closeMutation.error.message : undefined}
        onClose={async () => {
          setCloseOpen(false);
          if (closeMutation.data) {
            await queryClient.invalidateQueries({ queryKey: queryKeys.posCurrentSession(tenantId) });
            clearCart();
          }
          closeMutation.reset();
        }}
        onSubmit={(values) => closeMutation.mutate(values)}
      />
    </div>
  );
}

function ProductTile({ product, selected, onClick }: { product: POSProduct; selected: boolean; onClick: () => void }) {
  const isOut = product.availableStock <= 0;

  return (
    <button
      type="button"
      disabled={isOut}
      className={[
        "min-h-[230px] touch-manipulation overflow-hidden rounded-2xl border bg-white text-left shadow-sm transition hover:-translate-y-0.5 hover:shadow-soft disabled:cursor-not-allowed disabled:opacity-60",
        selected ? "border-primary-500 ring-2 ring-primary-100" : "border-neutral-200"
      ].join(" ")}
      onClick={onClick}
    >
      <div
        className="h-32 bg-neutral-100"
        style={
          product.image
            ? {
                backgroundImage: `url(${product.image})`,
                backgroundPosition: "center",
                backgroundSize: "cover"
              }
            : undefined
        }
      />
      <div className="space-y-2 p-3">
        <div>
          <p className="line-clamp-2 font-semibold text-neutral-950">{product.name}</p>
          <p className="mt-1 text-xs text-neutral-500">{product.sku || product.categoryName || "Produk POS"}</p>
        </div>
        <div className="flex items-center justify-between gap-2">
          <p className="font-semibold text-primary-700">{formatRupiah(product.price)}</p>
          <Badge tone={isOut ? "danger" : product.availableStock <= 3 ? "warning" : "success"}>
            {isOut ? "Habis" : `${product.availableStock} stok`}
          </Badge>
        </div>
      </div>
    </button>
  );
}

function CartLine({
  item,
  onQuantityChange,
  onRemove
}: {
  item: POSCartLine;
  onQuantityChange: (quantity: number) => void;
  onRemove: () => void;
}) {
  return (
    <div className="rounded-2xl border border-neutral-200 p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="font-semibold text-neutral-950">{item.name}</p>
          <p className="mt-1 text-xs text-neutral-500">
            {item.sku || "Tanpa SKU"} • {formatRupiah(item.price)}
          </p>
        </div>
        <Button type="button" variant="ghost" size="sm" onClick={onRemove}>
          Hapus
        </Button>
      </div>
      <div className="mt-4 flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Button type="button" variant="outline" size="sm" onClick={() => onQuantityChange(item.quantity - 1)}>
            -
          </Button>
          <Input
            className="w-20 text-center"
            type="number"
            min={1}
            max={item.availableStock}
            value={item.quantity}
            onChange={(event) => onQuantityChange(Number(event.target.value))}
          />
          <Button type="button" variant="outline" size="sm" onClick={() => onQuantityChange(item.quantity + 1)}>
            +
          </Button>
        </div>
        <p className="font-semibold text-neutral-950">{formatRupiah(item.price * item.quantity)}</p>
      </div>
    </div>
  );
}

function MoneyRow({ label, value, strong = false }: { label: string; value: number; strong?: boolean }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <span className={strong ? "font-semibold text-neutral-950" : "text-neutral-500"}>{label}</span>
      <span className={strong ? "font-semibold text-neutral-950" : "text-neutral-700"}>{formatRupiah(value)}</span>
    </div>
  );
}
