"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";

type ShareLinkButtonProps = {
  label?: string;
};

export function ShareLinkButton({ label = "Bagikan" }: ShareLinkButtonProps) {
  const [status, setStatus] = useState<string>();

  async function handleShare() {
    const url = window.location.href;

    try {
      if (navigator.share) {
        await navigator.share({
          title: document.title,
          url
        });
        setStatus("Berhasil dibagikan");
        return;
      }

      if (navigator.clipboard) {
        await navigator.clipboard.writeText(url);
        setStatus("Tautan disalin");
        return;
      }

      setStatus("Salin tautan dari bilah alamat");
    } catch {
      setStatus("Belum berhasil dibagikan");
    }
  }

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Button onClick={handleShare} size="sm" type="button" variant="outline">
        {label}
      </Button>
      {status ? (
        <span aria-live="polite" className="text-xs text-neutral-500">
          {status}
        </span>
      ) : null}
    </div>
  );
}
