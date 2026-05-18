"use client";

import { Button } from "@/components/ui/button";

export type ExistingImage = {
  id: string;
  url: string;
  altText?: string;
  isPrimary: boolean;
};

type ImageUploaderProps = {
  files: File[];
  onFilesChange: (files: File[]) => void;
  existingImages?: ExistingImage[];
  onDeleteExisting?: (imageId: string) => void;
  disabled?: boolean;
};

export function ImageUploader({
  files,
  onFilesChange,
  existingImages = [],
  onDeleteExisting,
  disabled = false
}: ImageUploaderProps) {
  return (
    <div className="space-y-4">
      <label className="flex cursor-pointer flex-col gap-2 rounded-2xl border border-dashed border-neutral-300 bg-neutral-50 p-4 text-sm text-neutral-600 transition hover:border-primary-300 hover:bg-primary-50">
        <span className="font-medium text-neutral-900">Pilih gambar produk</span>
        <span>JPEG, PNG, atau WebP. Validasi final tetap dilakukan backend.</span>
        <input
          className="sr-only"
          type="file"
          accept="image/jpeg,image/png,image/webp"
          multiple
          disabled={disabled}
          onChange={(event) => {
            const nextFiles = Array.from(event.target.files ?? []);
            onFilesChange(nextFiles);
            event.target.value = "";
          }}
        />
      </label>

      {files.length > 0 ? (
        <div className="rounded-2xl border border-neutral-200 bg-white p-4">
          <p className="text-sm font-medium text-neutral-900">Siap diunggah</p>
          <ul className="mt-3 space-y-2 text-sm text-neutral-600">
            {files.map((file) => (
              <li key={`${file.name}-${file.size}`} className="flex items-center justify-between gap-3">
                <span className="truncate">{file.name}</span>
                <span className="shrink-0 text-xs text-neutral-400">{Math.ceil(file.size / 1024)} KB</span>
              </li>
            ))}
          </ul>
        </div>
      ) : null}

      {existingImages.length > 0 ? (
        <div className="rounded-2xl border border-neutral-200 bg-white p-4">
          <p className="text-sm font-medium text-neutral-900">Gambar tersimpan</p>
          <ul className="mt-3 space-y-3 text-sm text-neutral-600">
            {existingImages.map((image) => (
              <li key={image.id} className="flex flex-col gap-2 rounded-xl bg-neutral-50 p-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="min-w-0">
                  <p className="truncate font-medium text-neutral-900">{image.altText || "Gambar produk"}</p>
                  <p className="truncate text-xs text-neutral-500">{image.url}</p>
                  {image.isPrimary ? <p className="mt-1 text-xs font-semibold text-primary-700">Gambar utama</p> : null}
                </div>
                {onDeleteExisting ? (
                  <Button type="button" variant="outline" size="sm" disabled={disabled} onClick={() => onDeleteExisting(image.id)}>
                    Hapus
                  </Button>
                ) : null}
              </li>
            ))}
          </ul>
        </div>
      ) : null}
    </div>
  );
}
