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
      className="relative inline-flex h-9 items-center justify-center rounded-xl border border-[#E3D2BC] bg-[#FFFDF8] px-2.5 text-sm font-semibold text-[#6F6256] transition hover:border-[#B96E45] hover:text-[#B96E45] md:h-10 md:px-3"
      href={`/s/${storeSlug}/cart`}
    >
      <ShoppingCart className="h-4 w-4" aria-hidden="true" />
      <span className="ml-2 hidden lg:inline">Keranjang</span>
      {cartStoreSlug === storeSlug && itemCount > 0 ? (
        <span className="absolute -right-2 -top-2 flex h-5 min-w-5 items-center justify-center rounded-full bg-[#B96E45] px-1.5 text-[11px] font-bold text-white">
          {itemCount}
        </span>
      ) : null}
    </Link>
  );
}
