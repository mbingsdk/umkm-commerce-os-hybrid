import { z } from "zod";

export const cancelOrderSchema = z.object({
  reason: z.string().trim().min(1, "Alasan pembatalan wajib diisi.").max(160, "Alasan maksimal 160 karakter."),
  note: z.string().trim().max(500, "Catatan maksimal 500 karakter.").optional()
});

export type CancelOrderFormValues = z.infer<typeof cancelOrderSchema>;

export const paymentReviewSchema = z.object({
  paymentConfirmationId: z.string().trim().min(1, "Pilih konfirmasi pembayaran yang akan direview."),
  note: z.string().trim().max(500, "Catatan maksimal 500 karakter.").optional()
});

export type PaymentReviewFormValues = z.infer<typeof paymentReviewSchema>;
