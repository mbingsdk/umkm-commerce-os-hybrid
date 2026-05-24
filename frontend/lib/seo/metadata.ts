import type { Metadata } from "next";

const appName = process.env.NEXT_PUBLIC_APP_NAME ?? "UMKM Commerce OS";
const siteURL = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";

export function siteUrl(path = "/") {
  return new URL(path, siteURL).toString();
}

export function publicPageMetadata({
  title,
  description,
  path,
  type = "website"
}: {
  title: string;
  description: string;
  path: string;
  type?: "website" | "article";
}): Metadata {
  const url = siteUrl(path);

  return {
    title,
    description,
    alternates: {
      canonical: url
    },
    openGraph: {
      title,
      description,
      locale: "id_ID",
      siteName: appName,
      type,
      url
    },
    twitter: {
      card: "summary_large_image",
      title,
      description
    }
  };
}
