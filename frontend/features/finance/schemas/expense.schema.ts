import { z } from "zod";

export const expenseCategories = [
  { value: "operasional", label: "Operasional" },
  { value: "bahan_baku", label: "Bahan baku" },
  { value: "gaji", label: "Gaji" },
  { value: "pengiriman", label: "Pengiriman" },
  { value: "marketing", label: "Marketing" },
  { value: "lainnya", label: "Lainnya" }
] as const;

export const expenseSchema = z.object({
  category: z.string().trim().optional(),
  title: z.string().trim().min(1, "Judul pengeluaran wajib diisi."),
  amount: z.number().int().positive("Nominal harus lebih dari 0."),
  expenseDate: z.string().trim().min(1, "Tanggal pengeluaran wajib diisi."),
  paymentMethod: z.string().trim().optional(),
  note: z.string().trim().optional()
});

export type ExpenseFormValues = z.infer<typeof expenseSchema>;

export function expenseCategoryLabel(value?: string) {
  return expenseCategories.find((category) => category.value === value)?.label ?? value ?? "Tanpa kategori";
}
