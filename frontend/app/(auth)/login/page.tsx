import { Badge } from "@/components/ui/badge";
import { AuthRedirectGuard } from "@/features/auth/components/auth-redirect-guard";
import { LoginForm } from "@/features/auth/components/login-form";

export default function LoginPage() {
  return (
    <AuthRedirectGuard>
      <section className="space-y-6">
        <div>
          <Badge tone="primary">Auth</Badge>
          <h1 className="mt-3 text-2xl font-bold text-neutral-950">Login</h1>
          <p className="mt-3 text-sm leading-6 text-neutral-500">
            Masuk untuk memilih tenant aktif dan melanjutkan pekerjaan toko hari ini.
          </p>
        </div>

        <LoginForm />
      </section>
    </AuthRedirectGuard>
  );
}
