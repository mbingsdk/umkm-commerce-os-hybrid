"use client";

import { useState, type ImgHTMLAttributes } from "react";
import { cn } from "@/lib/utils/cn";

type SafeImageProps = Omit<ImgHTMLAttributes<HTMLImageElement>, "src"> & {
  src?: string;
  fallbackLabel?: string;
  fallbackClassName?: string;
};

export function SafeImage({
  src,
  alt,
  className,
  fallbackLabel,
  fallbackClassName,
  onError,
  ...props
}: SafeImageProps) {
  const [failedSrc, setFailedSrc] = useState<string>();
  const hasFailed = Boolean(src && failedSrc === src);

  if (!src || hasFailed) {
    return (
      <div
        aria-label={alt || fallbackLabel}
        className={cn("flex items-center justify-center bg-neutral-100 text-sm text-neutral-400", fallbackClassName)}
        role={alt || fallbackLabel ? "img" : undefined}
      >
        {fallbackLabel}
      </div>
    );
  }

  return (
    // Centralize plain <img> usage here so remote public-upload URLs can still degrade gracefully without requiring
    // a broad Next.js image host allowlist during the local/public storefront phase.
    // eslint-disable-next-line @next/next/no-img-element
    <img
      alt={alt}
      className={className}
      onError={(event) => {
        if (src) {
          setFailedSrc(src);
        }
        onError?.(event);
      }}
      src={src}
      {...props}
    />
  );
}
