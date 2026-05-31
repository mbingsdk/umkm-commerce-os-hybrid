"use client";

import Link from "next/link";
import { Home, Package, Search, Store } from "lucide-react";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils/cn";

const dockItems = [
  { label: "Beranda", href: "/", icon: Home },
  { label: "Produk", href: "/products", icon: Package },
  { label: "Toko", href: "/stores", icon: Store },
  { label: "Cari", href: "/search", icon: Search }
];

export function PublicBottomDock() {
  const pathname = usePathname();

  return (
    <nav
      aria-label="Navigasi discovery mobile"
      className="fixed inset-x-0 bottom-0 z-40 border-t border-[#E3D2BC] bg-[#FFFDF8]/96 px-2 py-2 shadow-[0_-12px_30px_rgba(80,57,34,0.10)] backdrop-blur md:hidden"
    >
      <div className="mx-auto grid max-w-md grid-cols-4 gap-1">
        {dockItems.map(({ label, href, icon: Icon }) => (
          <Link
            className={cn(
              "flex min-h-12 flex-col items-center justify-center rounded-2xl text-xs font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#B96E45]",
              isActivePath(pathname, href)
                ? "bg-[#251F1A] text-[#FFFDF8]"
                : "text-[#6F6256] hover:bg-[#F1E7D8] hover:text-[#251F1A]"
            )}
            href={href}
            key={href}
          >
            <Icon aria-hidden="true" className="mb-0.5 h-4 w-4" />
            <span>{label}</span>
          </Link>
        ))}
      </div>
    </nav>
  );
}

function isActivePath(pathname: string, href: string) {
  if (href === "/") {
    return pathname === "/";
  }

  if (href === "/products") {
    return pathname === "/products" || pathname.startsWith("/products/") || pathname.startsWith("/category/");
  }

  if (href === "/stores") {
    return pathname === "/stores" || pathname.startsWith("/stores/") || pathname.startsWith("/city/");
  }

  return pathname === href || pathname.startsWith(`${href}/`);
}
