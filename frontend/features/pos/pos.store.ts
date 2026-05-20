"use client";

import { create } from "zustand";
import type { POSCartLine, POSPaymentMethod, POSProduct } from "@/features/pos/types";

type POSState = {
  items: POSCartLine[];
  selectedProduct: POSProduct | null;
  paymentMethod: POSPaymentMethod;
  amountPaid: number;
  setSelectedProduct: (product: POSProduct | null) => void;
  setPaymentMethod: (method: POSPaymentMethod) => void;
  setAmountPaid: (amount: number) => void;
  addProduct: (product: POSProduct) => void;
  updateQuantity: (productId: string, quantity: number) => void;
  removeItem: (productId: string) => void;
  clearCart: () => void;
  resetPayment: () => void;
};

export const usePOSStore = create<POSState>()((set) => ({
  items: [],
  selectedProduct: null,
  paymentMethod: "cash",
  amountPaid: 0,
  setSelectedProduct: (product) => set({ selectedProduct: product }),
  setPaymentMethod: (method) => set({ paymentMethod: method }),
  setAmountPaid: (amount) => set({ amountPaid: Math.max(0, amount) }),
  addProduct: (product) =>
    set((state) => {
      const existing = state.items.find((item) => item.productId === product.productId);
      if (existing) {
        return {
          items: state.items.map((item) =>
            item.productId === product.productId
              ? { ...item, quantity: Math.min(item.quantity + 1, item.availableStock) }
              : item
          )
        };
      }

      return {
        items: [
          ...state.items,
          {
            productId: product.productId,
            name: product.name,
            sku: product.sku,
            price: product.price,
            availableStock: product.availableStock,
            quantity: 1
          }
        ]
      };
    }),
  updateQuantity: (productId, quantity) =>
    set((state) => ({
      items: state.items
        .map((item) =>
          item.productId === productId
            ? { ...item, quantity: Math.min(Math.max(1, quantity), item.availableStock) }
            : item
        )
        .filter((item) => item.quantity > 0)
    })),
  removeItem: (productId) => set((state) => ({ items: state.items.filter((item) => item.productId !== productId) })),
  clearCart: () => set({ items: [], selectedProduct: null, amountPaid: 0 }),
  resetPayment: () => set({ paymentMethod: "cash", amountPaid: 0 })
}));
