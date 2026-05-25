import { z } from "zod";

export const stockAdjustmentSchema = z.object({
  adjustmentType: z.enum(["in", "out"], {
    error: "Tipe penyesuaian wajib dipilih."
  }),
  quantity: z
    .number({ error: "Jumlah wajib diisi." })
    .int("Jumlah harus berupa angka bulat.")
    .positive("Jumlah wajib lebih dari 0."),
  reason: z.string().trim().min(1, "Alasan wajib diisi.").max(120, "Alasan maksimal 120 karakter."),
  note: z.string().trim().max(500, "Catatan maksimal 500 karakter.").optional(),
  confirmedStockOut: z.boolean()
});

export type StockAdjustmentFormValues = z.infer<typeof stockAdjustmentSchema>;

export const thresholdSchema = z.object({
  lowStockThreshold: z
    .number({ error: "Threshold wajib diisi." })
    .int("Threshold harus berupa angka bulat.")
    .min(0, "Threshold tidak boleh negatif.")
});

export type ThresholdFormValues = z.infer<typeof thresholdSchema>;
