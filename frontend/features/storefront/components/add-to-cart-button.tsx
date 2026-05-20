"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useCartStore, type CartItem } from "@/features/storefront/cart.store";

type AddToCartButtonProps = {
  item: CartItem;
  disabled?: boolean;
  className?: string;
  size?: "sm" | "md" | "lg";
  label?: string;
};

export function AddToCartButton({
  item,
  disabled = false,
  className,
  size = "md",
  label = "Tambah ke Keranjang"
}: AddToCartButtonProps) {
  const addItem = useCartStore((state) => state.addItem);
  const clearCart = useCartStore((state) => state.clearCart);
  const [message, setMessage] = useState<string>();

  function handleAdd() {
    const result = addItem(item);

    if (result.ok) {
      setMessage("Produk masuk keranjang.");
      return;
    }

    if (result.reason === "different_store") {
      const shouldReplace = window.confirm(
        "Keranjang berisi produk dari toko lain. Kosongkan keranjang dan tambahkan produk ini?"
      );
      if (!shouldReplace) {
        return;
      }

      clearCart();
      addItem(item);
      setMessage("Keranjang diganti dengan produk toko ini.");
    }
  }

  return (
    <div className="space-y-1">
      <Button className={className} disabled={disabled} onClick={handleAdd} size={size} type="button">
        {disabled ? "Stok Habis" : label}
      </Button>
      {message ? <p className="text-xs font-medium text-primary-700">{message}</p> : null}
    </div>
  );
}
