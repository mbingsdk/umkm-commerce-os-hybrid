import type { MetadataRoute } from "next";
import { siteUrl } from "@/lib/seo/metadata";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
const MAX_SITEMAP_BATCHES = 5;

export const revalidate = 3600;

type ApiSuccessResponse<T> = {
  success: true;
  data?: T;
  meta?: {
    pagination?: {
      next_cursor?: string | null;
      has_more: boolean;
    };
  };
};

type SitemapStore = {
  slug: string;
  store_url?: string;
};

type SitemapProduct = {
  slug: string;
  product_url?: string;
  store: {
    slug: string;
  };
};

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const now = new Date();
  const staticPages: MetadataRoute.Sitemap = [
    entry("/", now, "daily", 1),
    entry("/explore", now, "daily", 0.9),
    entry("/stores", now, "daily", 0.8),
    entry("/products", now, "daily", 0.8)
  ];

  const [stores, products] = await Promise.all([fetchPublicStores(), fetchPublicProducts()]);

  const storePages = stores.map((store) => entry(normalizePublicPath(store.store_url ?? `/s/${store.slug}`), now, "daily", 0.7));
  const productPages = products.map((product) =>
    entry(
      normalizePublicPath(product.product_url ?? `/s/${product.store.slug}/products/${product.slug}`),
      now,
      "weekly",
      0.6
    )
  );

  return dedupeSitemap([...staticPages, ...storePages, ...productPages]);
}

async function fetchPublicStores() {
  return fetchPaginated<SitemapStore>("/api/v1/public/discovery/stores");
}

async function fetchPublicProducts() {
  return fetchPaginated<SitemapProduct>("/api/v1/public/discovery/products");
}

async function fetchPaginated<T>(path: string) {
  const items: T[] = [];
  let cursor: string | undefined;

  for (let page = 0; page < MAX_SITEMAP_BATCHES; page += 1) {
    const params = new URLSearchParams({ limit: "100" });
    if (cursor) {
      params.set("cursor", cursor);
    }

    try {
      const response = await fetch(`${API_BASE_URL}${path}?${params.toString()}`, {
        next: { revalidate }
      });
      if (!response.ok) {
        break;
      }

      const body = (await response.json()) as ApiSuccessResponse<T[]>;
      if (body.success !== true || !Array.isArray(body.data)) {
        break;
      }

      items.push(...body.data);
      if (!body.meta?.pagination?.has_more || !body.meta.pagination.next_cursor) {
        break;
      }
      cursor = body.meta.pagination.next_cursor;
    } catch {
      break;
    }
  }

  return items;
}

function entry(
  path: string,
  lastModified: Date,
  changeFrequency: MetadataRoute.Sitemap[number]["changeFrequency"],
  priority: number
) {
  return {
    url: siteUrl(path),
    lastModified,
    changeFrequency,
    priority
  };
}

function normalizePublicPath(value: string) {
  try {
    const parsed = new URL(value, siteUrl("/"));
    return `${parsed.pathname}${parsed.search}`;
  } catch {
    return "/";
  }
}

function dedupeSitemap(items: MetadataRoute.Sitemap) {
  const seen = new Set<string>();
  return items.filter((item) => {
    if (seen.has(item.url)) {
      return false;
    }
    seen.add(item.url);
    return true;
  });
}
