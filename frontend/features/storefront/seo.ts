import type { PublicProductDetail } from "@/features/storefront/types";

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
