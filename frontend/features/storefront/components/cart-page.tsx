"use client";

import Link from "next/link";
import { Minus, Plus, Trash2 } from "lucide-react";
import { EmptyState } from "@/components/feedback/empty-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useCartStore, getCartEstimatedSubtotal } from "@/features/storefront/cart.store";
import { SafeImage } from "@/features/storefront/components/safe-image";
import { formatRupiah } from "@/lib/format/money";

type CartPageProps = {
  storeSlug: string;
};

export function CartPage({ storeSlug }: CartPageProps) {
  const cartStoreSlug = useCartStore((state) => state.storeSlug);
  const items = useCartStore((state) => (state.storeSlug === storeSlug ? state.items : []));
  const updateQuantity = useCartStore((state) => state.updateQuantity);
  const removeItem = useCartStore((state) => state.removeItem);
  const clearCart = useCartStore((state) => state.clearCart);
  const subtotal = getCartEstimatedSubtotal(items);

  if (cartStoreSlug && cartStoreSlug !== storeSlug) {
    return (
      <main className="mx-auto max-w-4xl px-4 py-6 sm:px-6 lg:px-8">
        <EmptyState
          title="Keranjang berisi produk toko lain"
          description="Untuk saat ini, checkout hanya mendukung satu toko per transaksi. Kosongkan keranjang jika ingin belanja dari toko ini."
          action={
            <Button onClick={clearCart} type="button">
              Kosongkan Keranjang
            </Button>
          }
        />
      </main>
    );
  }

  if (items.length === 0) {
    return (
      <main className="mx-auto max-w-4xl px-4 py-6 sm:px-6 lg:px-8">
        <EmptyState
          title="Keranjangmu masih kosong"
          description="Yuk pilih produk dari toko ini sebelum lanjut checkout."
          action={
            <Link
              className="inline-flex h-10 items-center justify-center rounded-xl bg-[#251F1A] px-4 text-sm font-semibold text-[#FFFDF8] transition hover:bg-[#16110E]"
              href={`/s/${storeSlug}`}
            >
              Lihat Produk
            </Link>
          }
        />
      </main>
    );
  }

  return (
    <main className="mx-auto grid max-w-[1500px] gap-4 px-4 py-4 sm:px-6 sm:py-6 lg:grid-cols-[minmax(0,1fr)_320px] lg:px-8">
      <section className="space-y-3">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-[#251F1A]">Keranjang</h1>
          <p className="mt-1 text-sm text-[#6F6256]">Atur jumlah produk sebelum checkout.</p>
        </div>

        <div className="space-y-2.5">
          {items.map((item) => (
            <Card key={item.productId} className="rounded-[22px] border-[#E3D2BC] bg-[#FFFDF8] shadow-[0_8px_24px_rgba(89,63,38,0.055)]">
              <CardContent className="grid gap-3 p-3 sm:grid-cols-[80px_minmax(0,1fr)_auto] sm:items-center">
                <Link href={`/s/${storeSlug}/products/${item.slug}`} className="block">
                  <SafeImage
                    alt={item.name}
                    className="aspect-[5/4] w-full rounded-2xl object-cover sm:h-20 sm:aspect-square"
                    fallbackClassName="aspect-[5/4] w-full rounded-2xl sm:h-20 sm:aspect-square"
                    fallbackLabel="Foto"
                    src={item.imageUrl}
                  />
                </Link>

                <div className="min-w-0">
                  <Link
                    href={`/s/${storeSlug}/products/${item.slug}`}
                    className="font-semibold text-[#251F1A] hover:text-[#B96E45]"
                  >
                    {item.name}
                  </Link>
                  <p className="mt-1 text-sm text-[#6F6256]">{formatRupiah(item.displayPrice)} / item</p>
                  <p className="mt-1.5 text-sm font-semibold text-[#B96E45]">
                    Estimasi: {formatRupiah(item.displayPrice * item.quantity)}
                  </p>
                </div>

                <div className="flex items-center justify-between gap-3 sm:flex-col sm:items-end">
                  <div className="inline-flex items-center rounded-xl border border-[#E3D2BC] bg-white">
                    <button
                      aria-label={`Kurangi ${item.name}`}
                      className="flex h-9 w-9 touch-manipulation items-center justify-center text-[#6F6256] hover:text-[#B96E45]"
                      onClick={() => updateQuantity(item.productId, item.quantity - 1)}
                      type="button"
                    >
                      <Minus className="h-4 w-4" />
                    </button>
                    <span className="min-w-8 text-center text-sm font-semibold">{item.quantity}</span>
                    <button
                      aria-label={`Tambah ${item.name}`}
                      className="flex h-9 w-9 touch-manipulation items-center justify-center text-[#6F6256] hover:text-[#B96E45]"
                      onClick={() => updateQuantity(item.productId, item.quantity + 1)}
                      type="button"
                    >
                      <Plus className="h-4 w-4" />
                    </button>
                  </div>
                  <button
                    className="inline-flex items-center gap-1 text-sm font-semibold text-red-600 hover:text-red-700"
                    onClick={() => removeItem(item.productId)}
                    type="button"
                  >
                    <Trash2 className="h-4 w-4" />
                    Hapus
                  </button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      <aside className="lg:sticky lg:top-20 lg:self-start">
        <Card className="rounded-[24px] border-[#E3D2BC] bg-[#FFFDF8] shadow-[0_10px_30px_rgba(89,63,38,0.07)]">
          <CardHeader className="p-4">
            <CardTitle>Ringkasan pesanan</CardTitle>
            <CardDescription>Total di halaman ini adalah estimasi. Backend akan menghitung ulang harga final.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 p-4">
            <div className="flex items-center justify-between text-sm">
              <span className="text-[#6F6256]">Subtotal estimasi</span>
              <span className="font-bold text-[#251F1A]">{formatRupiah(subtotal)}</span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-[#6F6256]">Ongkir</span>
              <span className="font-semibold text-[#251F1A]">Dihitung saat checkout</span>
            </div>
            <Link
              className="inline-flex h-11 w-full items-center justify-center rounded-xl bg-[#251F1A] px-4 text-sm font-semibold text-[#FFFDF8] transition hover:bg-[#16110E]"
              href={`/s/${storeSlug}/checkout`}
            >
              Lanjut Checkout
            </Link>
            <Link className="block text-center text-sm font-semibold text-[#B96E45]" href={`/s/${storeSlug}`}>
              Tambah produk lain
            </Link>
          </CardContent>
        </Card>
      </aside>
    </main>
  );
}
