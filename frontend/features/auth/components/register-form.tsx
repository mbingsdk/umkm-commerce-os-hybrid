"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import type { ReactNode } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { register } from "@/features/auth/api/auth.api";
import { useAuthStore } from "@/lib/stores/auth.store";

const registerSchema = z
  .object({
    name: z.string().trim().min(1, "Nama wajib diisi."),
    email: z.string().trim().email("Email belum valid."),
    phone: z.string().trim().optional(),
    password: z.string().min(8, "Password minimal 8 karakter."),
    confirmPassword: z.string().min(1, "Konfirmasi password wajib diisi.")
  })
  .refine((values) => values.password === values.confirmPassword, {
    path: ["confirmPassword"],
    message: "Konfirmasi password belum sama."
  });

type RegisterFormValues = z.infer<typeof registerSchema>;

export function RegisterForm() {
  const router = useRouter();
  const setSession = useAuthStore((state) => state.setSession);

  const form = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      name: "",
      email: "",
      phone: "",
      password: "",
      confirmPassword: ""
    }
  });

  const registerMutation = useMutation({
    mutationFn: async (values: RegisterFormValues) =>
      register({
        name: values.name,
        email: values.email,
        phone: values.phone,
        password: values.password
      }),
    onSuccess: (session) => {
      setSession(session);
      router.push("/onboarding/create-store");
    }
  });

  return (
    <form className="space-y-4" onSubmit={form.handleSubmit((values) => registerMutation.mutate(values))}>
      <Field
        label="Nama lengkap"
        error={form.formState.errors.name?.message}
        input={
          <Input
            autoComplete="name"
            placeholder="Nama pemilik toko"
            hasError={!!form.formState.errors.name}
            {...form.register("name")}
          />
        }
      />
      <Field
        label="Email"
        error={form.formState.errors.email?.message}
        input={
          <Input
            type="email"
            autoComplete="email"
            placeholder="owner@toko.id"
            hasError={!!form.formState.errors.email}
            {...form.register("email")}
          />
        }
      />
      <Field
        label="Nomor HP"
        error={form.formState.errors.phone?.message}
        input={
          <Input
            autoComplete="tel"
            placeholder="08123456789"
            hasError={!!form.formState.errors.phone}
            {...form.register("phone")}
          />
        }
      />
      <Field
        label="Password"
        error={form.formState.errors.password?.message}
        input={
          <Input
            type="password"
            autoComplete="new-password"
            placeholder="Minimal 8 karakter"
            hasError={!!form.formState.errors.password}
            {...form.register("password")}
          />
        }
      />
      <Field
        label="Konfirmasi password"
        error={form.formState.errors.confirmPassword?.message}
        input={
          <Input
            type="password"
            autoComplete="new-password"
            placeholder="Ulangi password"
            hasError={!!form.formState.errors.confirmPassword}
            {...form.register("confirmPassword")}
          />
        }
      />

      {registerMutation.isError ? (
        <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">
          {registerMutation.error instanceof Error ? registerMutation.error.message : "Registrasi gagal. Coba lagi."}
        </p>
      ) : null}

      <Button className="w-full" type="submit" isLoading={registerMutation.isPending}>
        Buat akun
      </Button>

      <p className="text-sm text-neutral-500">
        Sudah punya akun?{" "}
        <Link className="font-semibold text-primary-700 hover:text-primary-800" href="/login">
          Masuk
        </Link>
      </p>
    </form>
  );
}

function Field({
  label,
  input,
  error
}: {
  label: string;
  input: ReactNode;
  error?: string;
}) {
  return (
    <label className="block space-y-2 text-sm font-medium text-neutral-700">
      {label}
      {input}
      {error ? <span className="block text-sm font-normal text-red-600">{error}</span> : null}
    </label>
  );
}
