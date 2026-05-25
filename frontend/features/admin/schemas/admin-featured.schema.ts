import { z } from "zod";

const optionalUuid = (message: string) =>
  z
    .string()
    .trim()
    .optional()
    .refine((value) => !value || z.uuid().safeParse(value).success, message);

const optionalDateTime = (message: string) =>
  z
    .string()
    .trim()
    .optional()
    .refine((value) => !value || !Number.isNaN(new Date(value).getTime()), message);

export const adminFeaturedSchema = z
  .object({
    itemType: z.enum(["store", "product"], { error: "Tipe featured wajib dipilih." }),
    tenantId: z.string().trim().min(1, "Tenant ID wajib diisi.").uuid("Tenant ID harus UUID valid."),
    storeId: optionalUuid("Store ID harus UUID valid."),
    productId: optionalUuid("Product ID harus UUID valid."),
    placement: z.enum(["home", "stores", "products", "category", "city"], {
      error: "Placement wajib dipilih."
    }),
    sortOrder: z
      .number({ error: "Urutan wajib diisi." })
      .int("Urutan harus angka bulat.")
      .min(0, "Urutan tidak boleh negatif."),
    startsAt: optionalDateTime("Tanggal mulai tidak valid."),
    endsAt: optionalDateTime("Tanggal selesai tidak valid."),
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

    if (value.startsAt && value.endsAt && new Date(value.endsAt).getTime() <= new Date(value.startsAt).getTime()) {
      context.addIssue({
        code: "custom",
        path: ["endsAt"],
        message: "Tanggal selesai harus setelah tanggal mulai."
      });
    }
  });

export type AdminFeaturedFormValues = z.infer<typeof adminFeaturedSchema>;
