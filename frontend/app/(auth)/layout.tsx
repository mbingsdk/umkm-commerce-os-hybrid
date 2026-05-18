import type { ReactNode } from "react";

export default function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <main className="flex min-h-screen items-center justify-center px-4 py-10">
      <section className="w-full max-w-md rounded-3xl border border-neutral-200 bg-white p-8 shadow-soft">
        {children}
      </section>
    </main>
  );
}

