import { z } from "zod";
import type { ProductStatus } from "@/features/catalog/types";

const slugPattern = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

export const productStatuses = ["draft", "active", "inactive", "archived"] as const satisfies readonly ProductStatus[];

export const productSchema = z
  .object({
    categoryId: z.string().optional(),
    name: z.string().trim().min(1, "Nama produk wajib diisi."),
    slug: z.string().trim().regex(slugPattern, "Slug hanya boleh huruf kecil, angka, dan tanda hubung."),
    description: z.string().trim().optional(),
    sku: z.string().trim().optional(),
    barcode: z.string().trim().optional(),
    price: z.number().min(0, "Harga harus nol atau lebih."),
    compareAtPrice: z.number().min(0, "Harga pembanding harus nol atau lebih.").nullable(),
    costPrice: z.number().min(0, "Harga modal harus nol atau lebih.").nullable(),
    weightGram: z.number().int().min(0, "Berat harus nol atau lebih."),
    initialStock: z.number().int().min(0, "Stok awal harus nol atau lebih."),
    status: z.enum(productStatuses),
    isDiscoverable: z.boolean(),
    trackInventory: z.boolean(),
    allowBackorder: z.boolean()
  })
  .refine((values) => values.compareAtPrice === null || values.compareAtPrice >= values.price, {
    path: ["compareAtPrice"],
    message: "Harga pembanding harus sama atau lebih besar dari harga jual."
  });

export type ProductFormValues = z.infer<typeof productSchema>;
