"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

export type CartItem = {
  productId: string;
  storeSlug: string;
  name: string;
  displayPrice: number;
  quantity: number;
};

type CartState = {
  storeSlug: string | null;
  items: CartItem[];
  addItem: (item: CartItem) => { ok: boolean; reason?: "different_store" };
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

        set({
          storeSlug: item.storeSlug,
          items: [...state.items, item]
        });

        return { ok: true };
      },
      clearCart: () => set({ storeSlug: null, items: [] })
    }),
    {
      name: "umkm-cart"
    }
  )
);
