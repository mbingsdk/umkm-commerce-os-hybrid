"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

export type CartItem = {
  productId: string;
  storeSlug: string;
  name: string;
  slug: string;
  imageUrl?: string;
  displayPrice: number;
  quantity: number;
};

type CartState = {
  storeSlug: string | null;
  items: CartItem[];
  addItem: (item: CartItem) => { ok: boolean; reason?: "different_store" };
  updateQuantity: (productId: string, quantity: number) => void;
  removeItem: (productId: string) => void;
  clearCart: () => void;
};

export const useCartStore = create<CartState>()(
  persist(
    (set, get) => ({
      storeSlug: null,
      items: [],
      addItem: (item) => {
        const state = get();

        if (state.storeSlug && state.storeSlug !== item.storeSlug) {
          return { ok: false, reason: "different_store" };
        }

        const existing = state.items.find((cartItem) => cartItem.productId === item.productId);
        if (existing) {
          set({
            storeSlug: item.storeSlug,
            items: state.items.map((cartItem) =>
              cartItem.productId === item.productId
                ? { ...cartItem, quantity: Math.max(1, cartItem.quantity + item.quantity) }
                : cartItem
            )
          });
          return { ok: true };
        }

        set({
          storeSlug: item.storeSlug,
          items: [...state.items, { ...item, quantity: Math.max(1, item.quantity) }]
        });

        return { ok: true };
      },
      updateQuantity: (productId, quantity) => {
        set((state) => ({
          items: state.items.map((item) =>
            item.productId === productId ? { ...item, quantity: Math.max(1, quantity) } : item
          )
        }));
      },
      removeItem: (productId) => {
        set((state) => {
          const items = state.items.filter((item) => item.productId !== productId);
          return {
            storeSlug: items.length === 0 ? null : state.storeSlug,
            items
          };
        });
      },
      clearCart: () => set({ storeSlug: null, items: [] })
    }),
    {
      name: "umkm-cart"
    }
  )
);

export function getCartEstimatedSubtotal(items: CartItem[]) {
  return items.reduce((total, item) => total + item.displayPrice * item.quantity, 0);
}
