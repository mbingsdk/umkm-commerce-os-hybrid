import { z } from "zod";

const optionalUuid = (message: string) =>
  z
    .string()
    .trim()
    .optional()
    .refine((value) => !value || z.uuid().safeParse(value).success, message);

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

export const adminFeaturedSchema = z
  .object({
    itemType: z.enum(["store", "product"], { error: "Tipe featured wajib dipilih." }),
    tenantId: z.string().trim().min(1, "Tenant ID wajib diisi.").uuid("Tenant ID harus UUID valid."),
    storeId: optionalUuid("Store ID harus UUID valid."),
    productId: optionalUuid("Product ID harus UUID valid."),
    placement: z.enum(["home", "stores", "products", "category", "city"], {
      error: "Placement wajib dipilih."
    }),
    sortOrder: z.number().int("Urutan harus angka bulat.").min(0, "Urutan tidak boleh negatif."),
    startsAt: z.string().trim().optional(),
    endsAt: z.string().trim().optional(),
    isActive: z.boolean()
  })
  .superRefine((value, context) => {
    if (value.itemType === "store" && !value.storeId) {
      context.addIssue({
        code: "custom",
        path: ["storeId"],
        message: "Store ID wajib diisi untuk featured store."
      });
    }

    if (value.itemType === "product" && !value.productId) {
      context.addIssue({
        code: "custom",
        path: ["productId"],
        message: "Product ID wajib diisi untuk featured product."
      });
    }

    if (value.startsAt && value.endsAt && new Date(value.endsAt).getTime() < new Date(value.startsAt).getTime()) {
      context.addIssue({
        code: "custom",
        path: ["endsAt"],
        message: "Tanggal selesai tidak boleh lebih awal dari tanggal mulai."
      });
    }
  });

export type AdminFeaturedFormValues = z.infer<typeof adminFeaturedSchema>;

export const adminPlanSchema = z.object({
  code: z
    .string()
    .trim()
    .min(1, "Kode paket wajib diisi.")
    .max(60, "Kode paket maksimal 60 karakter.")
    .regex(/^[a-z0-9_-]+$/, "Kode hanya boleh huruf kecil, angka, dash, atau underscore."),
  name: z.string().trim().min(1, "Nama paket wajib diisi.").max(120, "Nama paket maksimal 120 karakter."),
  description: z.string().trim().max(500, "Deskripsi maksimal 500 karakter.").optional(),
  priceMonthly: z.number().int("Harga bulanan harus angka bulat.").min(0, "Harga bulanan tidak boleh negatif."),
  productLimit: nullableLimitText("Limit produk"),
  staffLimit: nullableLimitText("Limit staff"),
  canUsePos: z.boolean(),
  canUseDiscovery: z.boolean(),
  canUseCourier: z.boolean(),
  isActive: z.boolean()
});

export type AdminPlanFormValues = z.infer<typeof adminPlanSchema>;

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
