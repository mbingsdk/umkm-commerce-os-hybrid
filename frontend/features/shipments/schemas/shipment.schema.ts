import { z } from "zod";

export const courierTypeSchema = z.enum(["internal", "manual"], {
  error: "Tipe kurir wajib dipilih."
});

export const shipmentStatusSchema = z.enum(
  ["pending", "ready_for_pickup", "picked_up", "on_delivery", "delivered", "failed", "cancelled"],
  {
    error: "Status pengiriman wajib dipilih."
  }
);

export const createShipmentSchema = z.object({
  courierType: courierTypeSchema,
  courierName: z.string().trim().max(120, "Nama kurir maksimal 120 karakter.").optional(),
  trackingNumber: z.string().trim().max(120, "Nomor resi maksimal 120 karakter.").optional(),
  shippingCost: z
    .number({ error: "Ongkir wajib diisi." })
    .int("Ongkir harus berupa angka bulat.")
    .min(0, "Ongkir tidak boleh negatif."),
  assignedToName: z.string().trim().max(120, "Nama petugas maksimal 120 karakter.").optional(),
  assignedToPhone: z
    .string()
    .trim()
    .max(32, "HP petugas maksimal 32 karakter.")
    .refine((value) => !value || value.length >= 8, "HP petugas minimal 8 karakter.")
    .optional(),
  note: z.string().trim().max(500, "Catatan maksimal 500 karakter.").optional()
});

export type CreateShipmentFormValues = z.infer<typeof createShipmentSchema>;

export const updateShipmentStatusSchema = z.object({
  status: shipmentStatusSchema,
  note: z.string().trim().max(500, "Catatan maksimal 500 karakter.").optional()
});

export type UpdateShipmentStatusFormValues = z.infer<typeof updateShipmentStatusSchema>;
