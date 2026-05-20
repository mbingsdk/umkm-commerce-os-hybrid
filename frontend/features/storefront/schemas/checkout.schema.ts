import { z } from "zod";

export const checkoutSchema = z.object({
  customerName: z.string().trim().min(2, "Nama wajib diisi."),
  customerPhone: z.string().trim().min(8, "Nomor HP wajib diisi."),
  customerEmail: z
    .string()
    .trim()
    .optional()
    .refine((value) => !value || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value), "Email tidak valid."),
  recipientName: z.string().trim().optional(),
  recipientPhone: z.string().trim().optional(),
  address: z.string().trim().min(8, "Alamat wajib diisi."),
  city: z.string().trim().optional(),
  province: z.string().trim().optional(),
  postalCode: z.string().trim().optional(),
  customerNote: z.string().trim().optional(),
  paymentMethod: z.literal("manual_transfer")
});

export type CheckoutFormValues = z.infer<typeof checkoutSchema>;
