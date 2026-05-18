import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export default function RegisterPage() {
  return (
    <section className="space-y-6">
      <div>
        <Badge tone="primary">Auth shell</Badge>
        <h1 className="mt-3 text-2xl font-bold text-neutral-950">Register</h1>
        <p className="mt-3 text-sm leading-6 text-neutral-500">
          Registrasi dan onboarding tenant belum diaktifkan. Shell ini hanya menyiapkan struktur form awal.
        </p>
      </div>

      <form className="space-y-4">
        <label className="block space-y-2 text-sm font-medium text-neutral-700">
          Nama lengkap
          <Input placeholder="Nama pemilik toko" disabled />
        </label>
        <label className="block space-y-2 text-sm font-medium text-neutral-700">
          Email
          <Input type="email" placeholder="owner@toko.id" disabled />
        </label>
        <Button className="w-full" disabled>
          Buat akun
        </Button>
      </form>
    </section>
  );
}
