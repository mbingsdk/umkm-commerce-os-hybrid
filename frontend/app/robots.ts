import type { MetadataRoute } from "next";
import { siteUrl } from "@/lib/seo/metadata";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: "*",
      allow: ["/", "/explore", "/stores", "/products", "/search", "/s/"],
      disallow: [
        "/admin",
        "/admin/",
        "/dashboard",
        "/dashboard/",
        "/login",
        "/register",
        "/onboarding",
        "/api/",
        "/s/*/cart",
        "/s/*/checkout",
        "/s/*/orders",
        "/s/*/track-order"
      ]
    },
    sitemap: siteUrl("/sitemap.xml")
  };
}
