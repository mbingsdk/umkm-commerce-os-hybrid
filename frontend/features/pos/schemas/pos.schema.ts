import { z } from "zod";

export const openSessionSchema = z.object({
  openingCashAmount: z.number().int("Kas awal harus berupa angka bulat.").min(0, "Kas awal tidak boleh negatif."),
  note: z.string().trim().max(500, "Catatan maksimal 500 karakter.").optional()
});

export type OpenSessionFormValues = z.infer<typeof openSessionSchema>;

export const closeSessionSchema = z.object({
  closingCashAmount: z.number().int("Kas aktual harus berupa angka bulat.").min(0, "Kas aktual tidak boleh negatif."),
  note: z.string().trim().max(500, "Catatan maksimal 500 karakter.").optional()
});

export type CloseSessionFormValues = z.infer<typeof closeSessionSchema>;

export const posPaymentSubmitSchema = z
  .object({
    itemCount: z.number().int().positive("Cart POS masih kosong."),
    paymentMethod: z.enum(["cash", "qris_manual"]),
    amountPaid: z.number().int("Jumlah dibayar harus berupa angka bulat.").min(0, "Jumlah dibayar tidak boleh negatif."),
    subtotalEstimate: z.number().int().min(0)
  })
  .superRefine((value, context) => {
    if (value.paymentMethod === "cash" && value.amountPaid < value.subtotalEstimate) {
      context.addIssue({
        code: "custom",
        path: ["amountPaid"],
        message: "Jumlah tunai harus sama dengan atau lebih besar dari total estimasi."
      });
    }

    if (value.paymentMethod === "qris_manual" && value.amountPaid !== value.subtotalEstimate) {
      context.addIssue({
        code: "custom",
        path: ["amountPaid"],
        message: "Untuk QRIS manual, jumlah dibayar harus sama dengan total estimasi."
      });
    }
  });

export type POSPaymentSubmitValues = z.infer<typeof posPaymentSubmitSchema>;
