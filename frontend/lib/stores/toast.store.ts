"use client";

import { create } from "zustand";

export type ToastTone = "success" | "error" | "warning" | "info";

export type ToastItem = {
  id: string;
  title: string;
  description?: string;
  tone: ToastTone;
};

type ToastState = {
  toasts: ToastItem[];
  pushToast: (toast: Omit<ToastItem, "id" | "tone"> & { tone?: ToastTone }) => string;
  dismissToast: (id: string) => void;
};

export const useToastStore = create<ToastState>()((set) => ({
  toasts: [],
  pushToast: ({ title, description, tone = "info" }) => {
    const id = crypto.randomUUID();
    set((state) => ({
      toasts: [...state.toasts, { id, title, description, tone }].slice(-4)
    }));
    return id;
  },
  dismissToast: (id) =>
    set((state) => ({
      toasts: state.toasts.filter((toast) => toast.id !== id)
    }))
}));
