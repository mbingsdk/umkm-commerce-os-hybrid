import { z } from "zod";

const slugPattern = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

export const categorySchema = z.object({
  name: z.string().trim().min(1, "Nama kategori wajib diisi."),
  slug: z.string().trim().regex(slugPattern, "Slug hanya boleh huruf kecil, angka, dan tanda hubung."),
  description: z.string().trim().optional(),
  sortOrder: z.number().int().min(0, "Urutan harus nol atau lebih."),
  isActive: z.boolean()
});

export type CategoryFormValues = z.infer<typeof categorySchema>;
