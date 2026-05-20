"use client";

import type { ReactNode } from "react";
import { useEffect } from "react";
import { cn } from "@/lib/utils/cn";
import { useToastStore } from "@/lib/stores/toast.store";

type ToastProps = {
  title: string;
  description?: string;
  tone?: "success" | "error" | "warning" | "info";
  children?: ReactNode;
};

const toneClassNames = {
  success: "border-green-200 bg-green-50 text-green-900",
  error: "border-red-200 bg-red-50 text-red-900",
  warning: "border-amber-200 bg-amber-50 text-amber-900",
  info: "border-blue-200 bg-blue-50 text-blue-900"
};

export function Toast({ title, description, tone = "info", children }: ToastProps) {
  return (
    <section className={cn("rounded-2xl border p-4 shadow-soft", toneClassNames[tone])}>
      <p className="text-sm font-semibold">{title}</p>
      {description ? <p className="mt-1 text-sm opacity-80">{description}</p> : null}
      {children}
    </section>
  );
}

export function ToastViewportPlaceholder() {
  const toasts = useToastStore((state) => state.toasts);
  const dismissToast = useToastStore((state) => state.dismissToast);

  useEffect(() => {
    if (toasts.length === 0) {
      return;
    }

    const timers = toasts.map((toast) => window.setTimeout(() => dismissToast(toast.id), 4500));
    return () => timers.forEach(window.clearTimeout);
  }, [dismissToast, toasts]);

  return (
    <div
      aria-live="polite"
      aria-relevant="additions text"
      className="pointer-events-none fixed bottom-4 right-4 z-[70] flex w-full max-w-sm flex-col gap-3"
      data-toast-viewport
    >
      {toasts.map((toast) => (
        <div key={toast.id} className="pointer-events-auto">
          <Toast title={toast.title} description={toast.description} tone={toast.tone}>
            <button
              type="button"
              className="mt-2 text-xs font-semibold underline"
              onClick={() => dismissToast(toast.id)}
            >
              Tutup
            </button>
          </Toast>
        </div>
      ))}
    </div>
  );
}
