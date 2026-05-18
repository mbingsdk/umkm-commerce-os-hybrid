import { EmptyState } from "@/components/feedback/empty-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PermissionGate } from "@/lib/permissions/permission-gate";
import { permissions } from "@/lib/permissions/permissions";

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Dashboard</CardTitle>
          <CardDescription>
            Shell dashboard sudah aktif dengan auth client-side, tenant aktif, dan navigasi berbasis permission.
          </CardDescription>
        </CardHeader>
        <CardContent className="text-sm leading-6 text-neutral-600">
          Halaman ini tetap ringan: modul bisnis belum dibangun, dan fetch dashboard sengaja tetap di client karena token
          MVP hidup di store browser.
        </CardContent>
      </Card>

      <EmptyState
        title="Belum ada data dashboard"
        description="State kosong ini menjadi pola awal sebelum modul bisnis mulai mengisi ringkasan operasional."
        action={
          <PermissionGate
            permission={permissions.storePublish}
            fallback={
              <Button variant="outline" disabled>
                Aksi tersedia nanti
              </Button>
            }
          >
            <Button variant="outline" disabled>
              Publish toko tersedia nanti
            </Button>
          </PermissionGate>
        }
      />
    </div>
  );
}
