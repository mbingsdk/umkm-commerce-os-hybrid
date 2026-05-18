import type { ReactNode } from "react";
import { Badge } from "@/components/ui/badge";

const placeholderNav = ["Dashboard", "Orders", "Products", "Inventory", "POS", "Finance"];

export function DashboardShell({ children }: { children: ReactNode }) {
  return (
    <main className="min-h-screen bg-neutral-50">
      <div className="mx-auto flex min-h-screen max-w-7xl">
        <aside className="hidden w-64 border-r border-neutral-200 bg-white p-5 lg:block">
          <p className="text-sm font-semibold text-primary-700">UMKM Commerce OS</p>
          <Badge className="mt-3" tone="neutral">
            Shell only
          </Badge>
          <nav className="mt-8 space-y-2 text-sm text-neutral-600">
            {placeholderNav.map((item) => (
              <div key={item} className="rounded-xl px-3 py-2">
                {item}
              </div>
            ))}
          </nav>
        </aside>

        <div className="flex min-w-0 flex-1 flex-col">
          <header className="border-b border-neutral-200 bg-white px-4 py-4 sm:px-6 lg:px-8">
            <div className="flex items-center justify-between gap-4">
              <div>
                <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">Dashboard shell</p>
                <p className="text-sm text-neutral-600">Auth dan tenant switcher masuk pada sprint berikutnya.</p>
              </div>
              <Badge tone="primary">Sprint 1</Badge>
            </div>
          </header>

          <div className="flex-1 px-4 py-6 sm:px-6 lg:px-8">{children}</div>
        </div>
      </div>
    </main>
  );
}
