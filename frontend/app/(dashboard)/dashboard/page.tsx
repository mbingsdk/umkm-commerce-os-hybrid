import { EmptyState } from "@/components/feedback/empty-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Dashboard</CardTitle>
          <CardDescription>
            Shell dashboard sudah siap. Ringkasan tenant, auth guard, dan permission-aware navigation belum
            diimplementasikan pada foundation ini.
          </CardDescription>
        </CardHeader>
        <CardContent className="text-sm leading-6 text-neutral-600">
          Halaman dashboard sengaja belum melakukan fetch server-side karena token MVP nantinya hidup di client store.
        </CardContent>
      </Card>

      <EmptyState
        title="Belum ada data dashboard"
        description="State kosong ini menjadi pola awal sebelum modul bisnis mulai mengisi ringkasan operasional."
        action={
          <Button variant="outline" disabled>
            Aksi tersedia nanti
          </Button>
        }
      />
    </div>
  );
}
