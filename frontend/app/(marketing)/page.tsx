import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

export default function MarketingHomePage() {
  return (
    <main className="mx-auto grid min-h-screen max-w-6xl items-center gap-8 px-4 py-12 sm:px-6 lg:grid-cols-[1fr_360px] lg:px-8">
      <section className="max-w-3xl">
        <Badge tone="primary">Sprint 1 frontend foundation</Badge>
        <h1 className="mt-4 text-4xl font-bold tracking-tight text-neutral-950 sm:text-5xl">
          Fondasi commerce untuk toko online, dashboard, dan POS UMKM.
        </h1>
        <p className="mt-5 max-w-2xl text-base leading-7 text-neutral-600">
          Struktur dasar sudah siap untuk pengalaman marketing, storefront, auth, dan dashboard tanpa mengaktifkan
          fitur bisnis sebelum waktunya.
        </p>
        <div className="mt-8 flex flex-wrap gap-3">
          <Button>Mulai nanti</Button>
          <Button variant="outline">Lihat shell</Button>
        </div>
      </section>

      <Card>
        <CardContent className="space-y-4">
          <div>
            <p className="text-sm font-semibold text-neutral-950">Yang sudah siap</p>
            <p className="mt-1 text-sm leading-6 text-neutral-500">
              Provider global, token visual, primitive UI, dan route shell awal.
            </p>
          </div>
          <div className="space-y-2 text-sm text-neutral-600">
            <p>• Storefront tetap ringan</p>
            <p>• Dashboard tetap client-ready</p>
            <p>• Fitur bisnis belum diaktifkan</p>
          </div>
        </CardContent>
      </Card>
    </main>
  );
}
