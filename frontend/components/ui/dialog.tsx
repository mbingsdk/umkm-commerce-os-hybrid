"use client";

import { X } from "lucide-react";
import { useEffect } from "react";
import type { ReactNode } from "react";
import { Button } from "@/components/ui/button";

type DialogProps = {
  open: boolean;
  title: string;
  description?: string;
  onClose: () => void;
  children: ReactNode;
  footer?: ReactNode;
};

export function Dialog({ open, title, description, onClose, children, footer }: DialogProps) {
  useEffect(() => {
    if (!open) {
      return;
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onClose();
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [onClose, open]);

  if (!open) {
    return null;
  }

  return (
    <div
      className="fixed inset-0 z-[60] flex items-center justify-center bg-neutral-950/50 p-4"
      role="presentation"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) {
          onClose();
        }
      }}
    >
      <section
        aria-describedby={description ? "dialog-description" : undefined}
        aria-labelledby="dialog-title"
        className="w-full max-w-lg rounded-2xl border border-neutral-200 bg-white shadow-floating"
        role="dialog"
        aria-modal="true"
      >
        <div className="flex items-start justify-between gap-4 border-b border-neutral-100 p-5">
          <div>
            <h2 id="dialog-title" className="text-lg font-semibold text-neutral-950">
              {title}
            </h2>
            {description ? (
              <p id="dialog-description" className="mt-1 text-sm leading-6 text-neutral-500">
                {description}
              </p>
            ) : null}
          </div>
          <Button variant="ghost" size="sm" aria-label="Tutup dialog" onClick={onClose} autoFocus>
            <X className="h-4 w-4" />
          </Button>
        </div>
        <div className="p-5">{children}</div>
        {footer ? <div className="flex justify-end gap-3 border-t border-neutral-100 p-5">{footer}</div> : null}
      </section>
    </div>
  );
}
