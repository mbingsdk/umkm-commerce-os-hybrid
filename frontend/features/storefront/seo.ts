import type { PublicProductDetail, PublicStore } from "@/features/storefront/types";

export function toAbsoluteURL(value?: string) {
  if (!value) {
    return undefined;
  }

  try {
    return new URL(value, getSiteURL()).toString();
  } catch {
    return undefined;
  }
}

export function getSiteURL() {
  return process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";
}

export function buildProductJsonLd(product: PublicProductDetail) {
  const productURL = `${getSiteURL()}/s/${product.store.slug}/products/${product.slug}`;
  const description =
    product.seo?.description ??
    product.description ??
    `Beli ${product.name} dari ${product.store.name}${product.store.city ? ` di ${product.store.city}` : ""}.`;
  const imageURLs = product.images
    .map((image) => toAbsoluteURL(image.url))
    .filter((image): image is string => Boolean(image));

  return {
    "@context": "https://schema.org",
    "@type": "Product",
    name: product.name,
    description,
    ...(imageURLs.length > 0 ? { image: imageURLs } : {}),
    offers: {
      "@type": "Offer",
      availability: schemaAvailability(product.stock.stockStatus),
      price: product.price,
      priceCurrency: "IDR",
      url: productURL
    }
  };
}

export function buildStoreJsonLd(store: PublicStore) {
  const storeURL = `${getSiteURL()}/s/${store.slug}`;
  const description =
    store.seo?.description ??
    store.description ??
    `${store.name}${store.city ? ` di ${store.city}` : ""} di UMKM Commerce OS.`;
  const image = toAbsoluteURL(store.bannerUrl ?? store.logoUrl);
  const logo = toAbsoluteURL(store.logoUrl);

  return {
    "@context": "https://schema.org",
    "@type": "Store",
    name: store.name,
    description,
    url: storeURL,
    ...(image ? { image } : {}),
    ...(logo ? { logo } : {}),
    ...(store.phone || store.whatsapp ? { telephone: store.phone ?? store.whatsapp } : {}),
    ...(store.city || store.province
      ? {
          address: {
            "@type": "PostalAddress",
            addressLocality: store.city,
            addressRegion: store.province,
            addressCountry: "ID"
          }
        }
      : {})
  };
}

export function serializeJsonLd(value: object) {
  return JSON.stringify(value).replace(/</g, "\\u003c");
}

function schemaAvailability(status: PublicProductDetail["stock"]["stockStatus"]) {
  switch (status) {
    case "in_stock":
      return "https://schema.org/InStock";
    case "low_stock":
      return "https://schema.org/LimitedAvailability";
    case "out_of_stock":
      return "https://schema.org/OutOfStock";
  }
}
