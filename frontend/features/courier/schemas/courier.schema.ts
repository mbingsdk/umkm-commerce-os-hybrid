import { z } from "zod";

export const courierZoneSchema = z.object({
  name: z.string().trim().min(1, "Nama zona wajib diisi.").max(120, "Nama zona maksimal 120 karakter."),
  description: z.string().trim().max(240, "Deskripsi maksimal 240 karakter.").optional(),
  rate: z.number().int("Ongkir harus berupa angka bulat.").min(0, "Ongkir tidak boleh negatif."),
  sortOrder: z.number().int("Urutan harus berupa angka bulat.").min(0, "Urutan tidak boleh negatif."),
  isActive: z.boolean()
});

export type CourierZoneFormValues = z.infer<typeof courierZoneSchema>;
