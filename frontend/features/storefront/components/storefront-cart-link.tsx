"use client";

import Link from "next/link";
import { ShoppingCart } from "lucide-react";
import { useCartStore } from "@/features/storefront/cart.store";

export function StorefrontCartLink({ storeSlug }: { storeSlug: string }) {
  const cartStoreSlug = useCartStore((state) => state.storeSlug);
  const itemCount = useCartStore((state) =>
    state.storeSlug === storeSlug ? state.items.reduce((total, item) => total + item.quantity, 0) : 0
  );

  return (
    <Link
      aria-label={itemCount > 0 ? `Keranjang, ${itemCount} item` : "Keranjang"}
      className="relative inline-flex h-10 items-center justify-center rounded-xl border border-neutral-200 bg-white px-3 text-sm font-semibold text-neutral-700 transition hover:border-primary-300 hover:text-primary-700"
      href={`/s/${storeSlug}/cart`}
    >
      <ShoppingCart className="h-4 w-4" aria-hidden="true" />
      <span className="ml-2 hidden sm:inline">Keranjang</span>
      {cartStoreSlug === storeSlug && itemCount > 0 ? (
        <span className="absolute -right-2 -top-2 flex h-5 min-w-5 items-center justify-center rounded-full bg-primary-600 px-1.5 text-[11px] font-bold text-white">
          {itemCount}
        </span>
      ) : null}
    </Link>
  );
}
