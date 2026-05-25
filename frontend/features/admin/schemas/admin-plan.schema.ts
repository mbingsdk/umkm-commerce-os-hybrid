import { z } from "zod";

const nullableLimitText = (field: string) =>
  z
    .string()
    .trim()
    .refine((value) => {
      if (!value) {
        return true;
      }
      const parsed = Number(value);
      return Number.isInteger(parsed) && parsed >= 0;
    }, `${field} harus angka bulat 0 atau lebih, atau kosong untuk unlimited.`);

export const adminPlanSchema = z.object({
  code: z
    .string()
    .trim()
    .min(1, "Kode paket wajib diisi.")
    .max(60, "Kode paket maksimal 60 karakter.")
    .regex(/^[a-z0-9_-]+$/, "Kode hanya boleh huruf kecil, angka, dash, atau underscore."),
  name: z.string().trim().min(1, "Nama paket wajib diisi.").max(120, "Nama paket maksimal 120 karakter."),
  description: z.string().trim().max(500, "Deskripsi maksimal 500 karakter.").optional(),
  priceMonthly: z
    .number({ error: "Harga bulanan wajib diisi." })
    .int("Harga bulanan harus angka bulat.")
    .min(0, "Harga bulanan tidak boleh negatif."),
  productLimit: nullableLimitText("Limit produk"),
  staffLimit: nullableLimitText("Limit staff"),
  canUsePos: z.boolean(),
  canUseDiscovery: z.boolean(),
  canUseCourier: z.boolean(),
  isActive: z.boolean()
});

export type AdminPlanFormValues = z.infer<typeof adminPlanSchema>;
