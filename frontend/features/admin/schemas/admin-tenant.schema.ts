import { z } from "zod";

export const adminTenantStatusSchema = z.object({
  status: z.enum(["trialing", "active", "suspended", "cancelled"], {
    error: "Status tenant wajib dipilih."
  }),
  reason: z.string().trim().max(500, "Reason maksimal 500 karakter.").optional()
});

export type AdminTenantStatusFormValues = z.infer<typeof adminTenantStatusSchema>;

export const adminTenantPlanSchema = z.object({
  planId: z.string().trim().min(1, "Paket wajib dipilih.").uuid("Plan ID harus UUID valid."),
  reason: z.string().trim().max(500, "Reason maksimal 500 karakter.").optional()
});

export type AdminTenantPlanFormValues = z.infer<typeof adminTenantPlanSchema>;
