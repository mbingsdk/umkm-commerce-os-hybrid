"use client";

import Link from "next/link";
import { Compass, Home, Package, ShoppingCart, Truck } from "lucide-react";
import { usePathname } from "next/navigation";
import { useCartStore } from "@/features/storefront/cart.store";
import { cn } from "@/lib/utils/cn";

type StorefrontBottomNavProps = {
  storeSlug: string;
};

export function StorefrontBottomNav({ storeSlug }: StorefrontBottomNavProps) {
  const pathname = usePathname();
  const cartStoreSlug = useCartStore((state) => state.storeSlug);
  const itemCount = useCartStore((state) =>
    state.storeSlug === storeSlug ? state.items.reduce((total, item) => total + item.quantity, 0) : 0
  );

  const cartCount = cartStoreSlug === storeSlug ? itemCount : 0;
  const links = [
    { label: "Toko", href: `/s/${storeSlug}`, icon: Home },
    { label: "Produk", href: `/s/${storeSlug}/products`, icon: Package },
    { label: "Lacak", href: `/s/${storeSlug}/track-order`, icon: Truck },
    { label: "Keranjang", href: `/s/${storeSlug}/cart`, icon: ShoppingCart, count: cartCount },
    { label: "Jelajah", href: "/", icon: Compass }
  ];

  return (
    <nav
      aria-label="Navigasi toko mobile"
      className="fixed inset-x-0 bottom-0 z-40 border-t border-[#E3D2BC] bg-[#FFFDF8]/96 px-2 py-2 shadow-[0_-12px_30px_rgba(80,57,34,0.10)] backdrop-blur md:hidden"
    >
      <div className="mx-auto grid max-w-md grid-cols-5 gap-1">
        {links.map(({ label, href, icon: Icon, count }) => (
          <Link
            className={cn(
              "relative flex min-h-12 flex-col items-center justify-center rounded-2xl text-xs font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#B96E45]",
              isActivePath(pathname, href)
                ? "bg-[#251F1A] text-[#FFFDF8]"
                : "text-[#6F6256] hover:bg-[#F1E7D8] hover:text-[#251F1A]"
            )}
            href={href}
            key={href}
          >
            <Icon aria-hidden="true" className="mb-0.5 h-4 w-4" />
            <span>{label}</span>
            {count ? (
              <span className="absolute right-2 top-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-[#B96E45] px-1 text-[10px] font-bold text-white">
                {count}
              </span>
            ) : null}
          </Link>
        ))}
      </div>
    </nav>
  );
}

function isActivePath(pathname: string, href: string) {
  if (href.endsWith("/products")) {
    return pathname === href || pathname.startsWith(`${href}/`) || pathname.includes("/categories/");
  }

  return pathname === href;
}
