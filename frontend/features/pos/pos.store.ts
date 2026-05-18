"use client";

import { create } from "zustand";

export type POSCartItem = {
  productId: string;
  name: string;
  price: number;
  quantity: number;
};

type POSState = {
  items: POSCartItem[];
  addItem: (item: POSCartItem) => void;
  clear: () => void;
};

export const usePOSStore = create<POSState>()((set) => ({
  items: [],
  addItem: (item) => set((state) => ({ items: [...state.items, item] })),
  clear: () => set({ items: [] })
}));
