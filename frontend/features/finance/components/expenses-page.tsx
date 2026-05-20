"use client";

import { useMutation, useQueryClient, type QueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { DataTable, type DataTableColumn } from "@/components/data-display/data-table";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  createExpense,
  deleteExpense,
  updateExpense
} from "@/features/finance/api/finance.api";
import { ExpenseFormDialog } from "@/features/finance/components/expense-form-dialog";
import { expenseCategories, expenseCategoryLabel, type ExpenseFormValues } from "@/features/finance/schemas/expense.schema";
import { useExpenses } from "@/features/finance/hooks/use-finance";
import type { Expense, ExpenseFilters } from "@/features/finance/types";
import { formatDate } from "@/lib/format/date";
import { formatRupiah } from "@/lib/format/money";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";
import { useToastStore } from "@/lib/stores/toast.store";

const pageLimit = 20;

export function ExpensesPage() {
  const queryClient = useQueryClient();
  const pushToast = useToastStore((state) => state.pushToast);
  const tenantId = useTenantStore((state) => state.selectedTenantId);
  const userPermissions = useTenantStore((state) => state.permissions);
  const canRead = userPermissions.includes(permissions.financeReadExpense);
  const canCreate = userPermissions.includes(permissions.financeCreateExpense);
  const canUpdate = userPermissions.includes(permissions.financeUpdateExpense);
  const canDelete = userPermissions.includes(permissions.financeDeleteExpense);
  const [query, setQuery] = useState("");
  const [category, setCategory] = useState("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [cursor, setCursor] = useState<string | null>(null);
  const [cursorStack, setCursorStack] = useState<string[]>([]);
  const [formOpen, setFormOpen] = useState(false);
  const [editingExpense, setEditingExpense] = useState<Expense | null>(null);
  const [deletingExpense, setDeletingExpense] = useState<Expense | null>(null);

  const filters = useMemo<ExpenseFilters>(
    () => ({
      query: query.trim(),
      category,
      dateFrom,
      dateTo,
      cursor,
      limit: pageLimit
    }),
    [category, cursor, dateFrom, dateTo, query]
  );
  const expensesQuery = useExpenses(filters, canRead);

  const createMutation = useMutation({
    mutationFn: createExpense,
    onSuccess: async () => {
      await invalidateFinanceQueries(queryClient, tenantId);
      setFormOpen(false);
      pushToast({ tone: "success", title: "Pengeluaran ditambahkan" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Gagal menambah pengeluaran", description: error.message });
    }
  });

  const updateMutation = useMutation({
    mutationFn: ({ expenseId, values }: { expenseId: string; values: ExpenseFormValues }) =>
      updateExpense(expenseId, values),
    onSuccess: async () => {
      await invalidateFinanceQueries(queryClient, tenantId);
      setEditingExpense(null);
      setFormOpen(false);
      pushToast({ tone: "success", title: "Pengeluaran diperbarui" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Gagal memperbarui pengeluaran", description: error.message });
    }
  });

  const deleteMutation = useMutation({
    mutationFn: deleteExpense,
    onSuccess: async () => {
      await invalidateFinanceQueries(queryClient, tenantId);
      setDeletingExpense(null);
      pushToast({ tone: "success", title: "Pengeluaran dihapus" });
    },
    onError: (error) => {
      pushToast({ tone: "error", title: "Gagal menghapus pengeluaran", description: error.message });
    }
  });

  function resetPagination() {
    setCursor(null);
    setCursorStack([]);
  }

  if (!canRead) {
    return (
      <EmptyState
        title="Akses pengeluaran belum tersedia"
        description="Role aktifmu belum memiliki izin untuk melihat data pengeluaran."
      />
    );
  }

  if (expensesQuery.isPending) {
    return <LoadingState lines={5} />;
  }

  if (expensesQuery.isError) {
    return (
      <ErrorState
        title="Pengeluaran belum bisa dimuat"
        description="Coba muat ulang atau ubah filter pencarian."
        onRetry={() => void expensesQuery.refetch()}
      />
    );
  }

  const expenses = expensesQuery.data.expenses;
  const pagination = expensesQuery.data.pagination;
  const columns: Array<DataTableColumn<Expense>> = [
    {
      key: "title",
      header: "Pengeluaran",
      render: (expense) => (
        <div>
          <p className="font-semibold text-neutral-950">{expense.title}</p>
          <p className="mt-1 text-xs text-neutral-500">
            {expense.categoryName || expenseCategoryLabel(expense.category)}
          </p>
        </div>
      )
    },
    {
      key: "date",
      header: "Tanggal",
      render: (expense) => formatDate(expense.expenseDate)
    },
    {
      key: "method",
      header: "Metode",
      render: (expense) => paymentMethodLabel(expense.paymentMethod)
    },
    {
      key: "amount",
      header: "Nominal",
      render: (expense) => <span className="font-semibold text-neutral-950">{formatRupiah(expense.amount)}</span>
    },
    {
      key: "actions",
      header: "Aksi",
      render: (expense) => (
        <div className="flex flex-wrap gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={!canUpdate}
            onClick={() => {
              setEditingExpense(expense);
              setFormOpen(true);
            }}
          >
            Edit
          </Button>
          <Button
            type="button"
            variant="danger"
            size="sm"
            disabled={!canDelete}
            onClick={() => setDeletingExpense(expense)}
          >
            Hapus
          </Button>
        </div>
      )
    }
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-neutral-950">Pengeluaran</h1>
          <p className="mt-1 text-sm text-neutral-500">
            Catat expense sederhana untuk estimasi net. Belum termasuk ledger akuntansi penuh.
          </p>
        </div>
        <Button
          type="button"
          disabled={!canCreate}
          onClick={() => {
            setEditingExpense(null);
            setFormOpen(true);
          }}
        >
          Tambah pengeluaran
        </Button>
      </div>

      <div className="grid gap-3 rounded-2xl border border-neutral-200 bg-white p-4 lg:grid-cols-[1.2fr_180px_160px_160px]">
        <Input
          placeholder="Cari judul atau catatan..."
          value={query}
          onChange={(event) => {
            setQuery(event.target.value);
            resetPagination();
          }}
        />
        <select
          className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
          value={category}
          onChange={(event) => {
            setCategory(event.target.value);
            resetPagination();
          }}
        >
          <option value="">Semua kategori</option>
          {expenseCategories.map((item) => (
            <option key={item.value} value={item.value}>
              {item.label}
            </option>
          ))}
        </select>
        <Input
          type="date"
          value={dateFrom}
          onChange={(event) => {
            setDateFrom(event.target.value);
            resetPagination();
          }}
        />
        <Input
          type="date"
          value={dateTo}
          onChange={(event) => {
            setDateTo(event.target.value);
            resetPagination();
          }}
        />
      </div>

      {expenses.length === 0 ? (
        <EmptyState
          title="Belum ada pengeluaran"
          description="Tambahkan biaya operasional, bahan baku, gaji, pengiriman, marketing, atau lainnya."
          action={
            canCreate ? (
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setEditingExpense(null);
                  setFormOpen(true);
                }}
              >
                Tambah pengeluaran pertama
              </Button>
            ) : null
          }
        />
      ) : (
        <>
          <DataTable columns={columns} rows={expenses} getRowKey={(expense) => expense.id} />
          <div className="flex items-center justify-between gap-3">
            <Button
              type="button"
              variant="outline"
              disabled={cursorStack.length === 0}
              onClick={() => {
                const previous = cursorStack[cursorStack.length - 1];
                setCursor(previous || null);
                setCursorStack((items) => items.slice(0, -1));
              }}
            >
              Sebelumnya
            </Button>
            <p className="text-sm text-neutral-500">Menampilkan maksimal {pagination.limit} pengeluaran per halaman.</p>
            <Button
              type="button"
              variant="outline"
              disabled={!pagination.hasMore || !pagination.nextCursor}
              onClick={() => {
                setCursorStack((items) => [...items, cursor ?? ""]);
                setCursor(pagination.nextCursor ?? null);
              }}
            >
              Berikutnya
            </Button>
          </div>
        </>
      )}

      <ExpenseFormDialog
        open={formOpen}
        expense={editingExpense}
        isSubmitting={createMutation.isPending || updateMutation.isPending}
        error={createMutation.error?.message ?? updateMutation.error?.message}
        onClose={() => {
          setFormOpen(false);
          setEditingExpense(null);
        }}
        onSubmit={(values) => {
          if (editingExpense) {
            updateMutation.mutate({ expenseId: editingExpense.id, values });
            return;
          }
          createMutation.mutate(values);
        }}
      />

      <Dialog
        open={!!deletingExpense}
        title="Hapus pengeluaran?"
        description="Data akan dihapus secara soft delete jika backend mendukung. Ringkasan finance akan dihitung ulang."
        onClose={() => setDeletingExpense(null)}
        footer={
          <>
            <Button type="button" variant="outline" onClick={() => setDeletingExpense(null)}>
              Batal
            </Button>
            <Button
              type="button"
              variant="danger"
              isLoading={deleteMutation.isPending}
              onClick={() => deletingExpense && deleteMutation.mutate(deletingExpense.id)}
            >
              Hapus
            </Button>
          </>
        }
      >
        <p className="text-sm leading-6 text-neutral-600">
          Kamu akan menghapus <span className="font-semibold text-neutral-950">{deletingExpense?.title}</span> senilai{" "}
          <span className="font-semibold text-neutral-950">{formatRupiah(deletingExpense?.amount ?? 0)}</span>.
        </p>
      </Dialog>
    </div>
  );
}

async function invalidateFinanceQueries(queryClient: QueryClient, tenantId: string | null) {
  await Promise.all([
    queryClient.invalidateQueries({ queryKey: ["finance-expenses", tenantId] }),
    queryClient.invalidateQueries({ queryKey: ["finance-summary", tenantId] }),
    queryClient.invalidateQueries({ queryKey: ["finance-daily-report", tenantId] }),
    queryClient.invalidateQueries({ queryKey: ["finance-monthly-report", tenantId] }),
    queryClient.invalidateQueries({ queryKey: ["dashboard-summary", tenantId] })
  ]);
}

function paymentMethodLabel(value?: string) {
  const labels: Record<string, string> = {
    cash: "Cash",
    bank_transfer: "Transfer bank",
    qris_manual: "QRIS manual",
    other: "Lainnya"
  };

  return value ? labels[value] ?? value : "Tidak dicatat";
}
