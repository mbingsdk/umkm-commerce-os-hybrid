import type { Metadata } from "next";
import type { ReactNode } from "react";
import { AppProviders } from "@/components/layout/app-providers";
import "./globals.css";

export const metadata: Metadata = {
  metadataBase: new URL(process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000"),
  title: "UMKM Commerce OS",
  description: "Commerce OS untuk UMKM: toko online, inventory, kasir, keuangan, dan kurir lokal.",
  openGraph: {
    title: "UMKM Commerce OS",
    description: "Commerce OS untuk UMKM: toko online, inventory, kasir, keuangan, dan kurir lokal.",
    locale: "id_ID",
    siteName: "UMKM Commerce OS",
    type: "website"
  }
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
