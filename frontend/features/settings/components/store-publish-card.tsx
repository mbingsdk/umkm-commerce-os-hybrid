"use client";

import Link from "next/link";
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog } from "@/components/ui/dialog";
import type { Store } from "@/features/settings/api/store.api";
import { formatDateTime } from "@/lib/format/date";

type StorePublishCardProps = {
  store: Store;
  publicUrl: string;
  canPublish: boolean;
  isSubmitting?: boolean;
  onPublish: () => void;
  onUnpublish: () => void;
};

export function StorePublishCard({
  store,
  publicUrl,
  canPublish,
  isSubmitting = false,
  onPublish,
  onUnpublish
}: StorePublishCardProps) {
  const [confirmUnpublishOpen, setConfirmUnpublishOpen] = useState(false);
  const published = store.status === "published";

  function handleConfirmUnpublish() {
    if (!canPublish || isSubmitting) {
      return;
    }

    setConfirmUnpublishOpen(false);
    onUnpublish();
  }

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <CardTitle>Status publish</CardTitle>
              <CardDescription>
                Toko yang publish bisa dibuka publik jika tenant masih aktif/trialing.
              </CardDescription>
            </div>
            <Badge tone={published ? "success" : "warning"}>{published ? "Published" : "Belum publish"}</Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-4">
            <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">Preview link toko public</p>
            <Link
              className="mt-2 block break-all text-sm font-semibold text-primary-700 hover:text-primary-800"
              href={publicUrl}
              target="_blank"
            >
              {publicUrl}
            </Link>
            <p className="mt-2 text-xs leading-5 text-neutral-500">
              Link ini aman dibagikan setelah toko publish. Jika toko belum publish, customer akan melihat halaman tidak tersedia.
            </p>
          </div>

          {store.publishedAt ? (
            <p className="text-sm text-neutral-500">Terakhir publish: {formatDateTime(store.publishedAt)}</p>
          ) : null}

          <div className="flex flex-col gap-3 sm:flex-row">
            {published ? (
              <Button
                type="button"
                variant="outline"
                isLoading={isSubmitting}
                disabled={!canPublish || isSubmitting}
                onClick={() => setConfirmUnpublishOpen(true)}
              >
                Unpublish toko
              </Button>
            ) : (
              <Button
                type="button"
                isLoading={isSubmitting}
                disabled={!canPublish || isSubmitting}
                onClick={onPublish}
              >
                Publish toko
              </Button>
            )}
            <Link href={publicUrl} target="_blank">
              <Button type="button" variant="ghost">
                Buka storefront
              </Button>
            </Link>
          </div>

          {!canPublish ? (
            <p className="rounded-xl bg-amber-50 p-3 text-sm text-amber-700">
              Role aktifmu belum memiliki izin publish/unpublish toko.
            </p>
          ) : null}
        </CardContent>
      </Card>

      <Dialog
        open={confirmUnpublishOpen}
        title="Unpublish toko?"
        description="Storefront publik tidak bisa dibuka customer sampai toko dipublish kembali."
        onClose={() => {
          if (!isSubmitting) {
            setConfirmUnpublishOpen(false);
          }
        }}
        footer={
          <>
            <Button
              type="button"
              variant="ghost"
              disabled={isSubmitting}
              onClick={() => setConfirmUnpublishOpen(false)}
            >
              Batal
            </Button>
            <Button
              type="button"
              variant="danger"
              isLoading={isSubmitting}
              disabled={!canPublish || isSubmitting}
              onClick={handleConfirmUnpublish}
            >
              Ya, unpublish
            </Button>
          </>
        }
      >
        <p className="text-sm leading-6 text-neutral-600">
          Produk dan data toko tidak dihapus, tetapi link publik <span className="font-semibold">{publicUrl}</span>{" "}
          akan menampilkan halaman tidak tersedia.
        </p>
      </Dialog>
    </>
  );
}
