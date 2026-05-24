"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { login } from "@/features/auth/api/auth.api";
import { listTenants } from "@/features/tenant/api/tenant.api";
import { useAuthStore } from "@/lib/stores/auth.store";
import { useTenantStore } from "@/lib/stores/tenant.store";

const loginSchema = z.object({
  email: z.string().trim().email("Email belum valid."),
  password: z.string().min(1, "Password wajib diisi.")
});

type LoginFormValues = z.infer<typeof loginSchema>;

export function LoginForm() {
  const router = useRouter();
  const setSession = useAuthStore((state) => state.setSession);
  const setTenants = useTenantStore((state) => state.setTenants);
  const selectTenant = useTenantStore((state) => state.selectTenant);

  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: ""
    }
  });

  const loginMutation = useMutation({
    mutationFn: async (values: LoginFormValues) => {
      const session = await login(values);
      setSession(session);
      const tenants = await listTenants();
      return { session, tenants };
    },
    onSuccess: ({ tenants }) => {
      setTenants(tenants);

      if (tenants.length === 0) {
        router.push("/onboarding/create-store");
        return;
      }

      selectTenant(tenants[0]);
      router.push("/dashboard");
    }
  });

  function handleSubmit(values: LoginFormValues) {
    if (loginMutation.isPending) {
      return;
    }

    loginMutation.mutate(values);
  }

  return (
    <form className="space-y-4" onSubmit={form.handleSubmit(handleSubmit)}>
      <div className="space-y-2">
        <label htmlFor="email" className="block text-sm font-medium text-neutral-700">
          Email
        </label>
        <Input
          id="email"
          type="email"
          autoComplete="email"
          placeholder="owner@toko.id"
          hasError={!!form.formState.errors.email}
          {...form.register("email")}
        />
        {form.formState.errors.email ? (
          <p className="text-sm text-red-600">{form.formState.errors.email.message}</p>
        ) : null}
      </div>

      <div className="space-y-2">
        <label htmlFor="password" className="block text-sm font-medium text-neutral-700">
          Password
        </label>
        <Input
          id="password"
          type="password"
          autoComplete="current-password"
          placeholder="••••••••"
          hasError={!!form.formState.errors.password}
          {...form.register("password")}
        />
        {form.formState.errors.password ? (
          <p className="text-sm text-red-600">{form.formState.errors.password.message}</p>
        ) : null}
      </div>

      {loginMutation.isError ? (
        <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">
          {loginMutation.error instanceof Error ? loginMutation.error.message : "Login gagal. Coba lagi."}
        </p>
      ) : null}

      <Button className="w-full" type="submit" isLoading={loginMutation.isPending} disabled={loginMutation.isPending}>
        Masuk
      </Button>

      <p className="text-sm text-neutral-500">
        Belum punya akun?{" "}
        <Link className="font-semibold text-primary-700 hover:text-primary-800" href="/register">
          Daftar sekarang
        </Link>
      </p>
    </form>
  );
}
