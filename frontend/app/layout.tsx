import type { Metadata } from "next";
import type { ReactNode } from "react";
import { AppProviders } from "@/components/layout/app-providers";
import "./globals.css";

export const metadata: Metadata = {
  title: "UMKM Commerce OS",
  description: "Commerce OS untuk UMKM: toko online, inventory, kasir, keuangan, dan kurir lokal."
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="id">
      <body>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}

