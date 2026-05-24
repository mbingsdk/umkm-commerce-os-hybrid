import { z } from "zod";

const optionalText = (max: number, label: string) =>
  z
    .string()
    .trim()
    .max(max, `${label} maksimal ${max} karakter`)
    .optional()
    .or(z.literal(""));

const timePattern = /^([01]\d|2[0-3]):[0-5]\d$/;

export const storeProfileSchema = z.object({
  name: z.string().trim().min(2, "Nama toko minimal 2 karakter").max(120, "Nama toko maksimal 120 karakter"),
  description: optionalText(500, "Deskripsi"),
  phone: optionalText(32, "Nomor telepon"),
  whatsapp: optionalText(32, "Nomor WhatsApp"),
  email: z.string().trim().email("Email tidak valid").optional().or(z.literal("")),
  address: optionalText(300, "Alamat"),
  city: optionalText(80, "Kota"),
  province: optionalText(80, "Provinsi"),
  postalCode: optionalText(16, "Kode pos"),
  isDiscoverable: z.boolean()
});

export const businessHourItemSchema = z
  .object({
    dayOfWeek: z.number().int().min(1).max(7),
    openTime: z.string().trim().optional().or(z.literal("")),
    closeTime: z.string().trim().optional().or(z.literal("")),
    isClosed: z.boolean()
  })
  .superRefine((value, ctx) => {
    if (value.isClosed) {
      return;
    }

    if (!value.openTime || !timePattern.test(value.openTime)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["openTime"],
        message: "Jam buka wajib format HH:MM"
      });
    }

    if (!value.closeTime || !timePattern.test(value.closeTime)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["closeTime"],
        message: "Jam tutup wajib format HH:MM"
      });
    }

    if (value.openTime && value.closeTime && value.openTime >= value.closeTime) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["closeTime"],
        message: "Jam tutup harus setelah jam buka"
      });
    }
  });

export const businessHoursSchema = z.object({
  items: z.array(businessHourItemSchema).length(7, "Lengkapi 7 hari operasional")
});

export type StoreProfileFormValues = z.infer<typeof storeProfileSchema>;
export type BusinessHoursFormValues = z.infer<typeof businessHoursSchema>;
