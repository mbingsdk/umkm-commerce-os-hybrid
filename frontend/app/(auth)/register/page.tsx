import { Badge } from "@/components/ui/badge";
import { RegisterForm } from "@/features/auth/components/register-form";

export default function RegisterPage() {
  return (
    <section className="space-y-6">
      <div>
        <Badge tone="primary">Auth</Badge>
        <h1 className="mt-3 text-2xl font-bold text-neutral-950">Daftar</h1>
        <p className="mt-3 text-sm leading-6 text-neutral-500">
          Buat akun pemilik toko, lalu lanjutkan ke onboarding untuk menyiapkan toko pertama.
        </p>
      </div>

      <RegisterForm />
    </section>
  );
}
