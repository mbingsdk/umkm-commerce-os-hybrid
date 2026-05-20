import { apiFetch, apiFetchWithMeta } from "@/lib/api/client";
import type {
  CreateExpenseInput,
  DailyReport,
  DailyReportDay,
  Expense,
  ExpenseFilters,
  FinanceDateFilters,
  FinanceMetrics,
  FinanceSummary,
  ListExpensesResult,
  MonthlyReport,
  MonthlyReportFilters,
  MonthlyReportMonth,
  Pagination,
  UpdateExpenseInput
} from "@/features/finance/types";

type ApiPaginationMeta = {
  pagination?: {
    limit: number;
    next_cursor?: string | null;
    has_more: boolean;
  };
};

type ApiFinanceMetrics = {
  gross_sales: number;
  online_sales: number;
  pos_sales: number;
  total_expenses: number;
  net_estimate: number;
  order_count: number;
  pos_transaction_count: number;
  total_orders: number;
  total_pos_transactions: number;
  average_order_value: number;
};

type ApiFinanceSummary = ApiFinanceMetrics & {
  date_from: string;
  date_to: string;
  note: string;
};

type ApiDailyReportDay = ApiFinanceMetrics & {
  date: string;
};

type ApiDailyReport = {
  date_from: string;
  date_to: string;
  days: ApiDailyReportDay[];
  summary: ApiFinanceMetrics;
  note: string;
};

type ApiMonthlyReportMonth = ApiFinanceMetrics & {
  year: number;
  month: number;
};

type ApiMonthlyReport = {
  year: number;
  month?: number | null;
  months: ApiMonthlyReportMonth[];
  summary: ApiFinanceMetrics;
  note: string;
};

type ApiExpense = {
  id: string;
  category_id?: string | null;
  category?: string;
  category_name?: string;
  title: string;
  amount: number;
  expense_date: string;
  payment_method?: string;
  note?: string;
  created_by?: {
    id: string;
    name?: string;
  } | null;
  created_at: string;
  updated_at: string;
};

export async function getFinanceSummary(filters: FinanceDateFilters = {}): Promise<FinanceSummary> {
  const suffix = toQueryString(filters);
  const result = await apiFetch<ApiFinanceSummary>(`/api/v1/finance/summary${suffix}`);
  return {
    ...normalizeMetrics(result),
    dateFrom: result.date_from,
    dateTo: result.date_to,
    note: result.note
  };
}

export async function getDailyReport(filters: FinanceDateFilters = {}): Promise<DailyReport> {
  const suffix = toQueryString(filters);
  const result = await apiFetch<ApiDailyReport>(`/api/v1/finance/reports/daily${suffix}`);
  return {
    dateFrom: result.date_from,
    dateTo: result.date_to,
    days: result.days.map(normalizeDailyDay),
    summary: normalizeMetrics(result.summary),
    note: result.note
  };
}

export async function getMonthlyReport(filters: MonthlyReportFilters = {}): Promise<MonthlyReport> {
  const params = new URLSearchParams();
  if (filters.year) {
    params.set("year", String(filters.year));
  }
  if (filters.month) {
    params.set("month", String(filters.month));
  }

  const suffix = params.size > 0 ? `?${params.toString()}` : "";
  const result = await apiFetch<ApiMonthlyReport>(`/api/v1/finance/reports/monthly${suffix}`);
  return {
    year: result.year,
    month: result.month,
    months: result.months.map(normalizeMonthlyMonth),
    summary: normalizeMetrics(result.summary),
    note: result.note
  };
}

export async function listExpenses(filters: ExpenseFilters = {}): Promise<ListExpensesResult> {
  const suffix = toQueryString(filters);
  const result = await apiFetchWithMeta<ApiExpense[], ApiPaginationMeta>(`/api/v1/finance/expenses${suffix}`);

  return {
    expenses: result.data.map(normalizeExpense),
    pagination: normalizePagination(result.meta)
  };
}

export async function createExpense(input: CreateExpenseInput): Promise<Expense> {
  const result = await apiFetch<ApiExpense>("/api/v1/finance/expenses", {
    method: "POST",
    body: JSON.stringify(toExpensePayload(input))
  });

  return normalizeExpense(result);
}

export async function updateExpense(expenseId: string, input: UpdateExpenseInput): Promise<Expense> {
  const result = await apiFetch<ApiExpense>(`/api/v1/finance/expenses/${expenseId}`, {
    method: "PATCH",
    body: JSON.stringify(toExpensePayload(input))
  });

  return normalizeExpense(result);
}

export async function deleteExpense(expenseId: string): Promise<Expense> {
  const result = await apiFetch<ApiExpense>(`/api/v1/finance/expenses/${expenseId}`, {
    method: "DELETE"
  });

  return normalizeExpense(result);
}

function toQueryString(filters: Record<string, unknown>) {
  const params = new URLSearchParams();

  Object.entries(filters).forEach(([key, value]) => {
    if (value == null || value === "") {
      return;
    }

    const apiKey =
      key === "dateFrom"
        ? "date_from"
        : key === "dateTo"
          ? "date_to"
          : key === "query"
            ? "q"
            : key;

    params.set(apiKey, String(value));
  });

  return params.size > 0 ? `?${params.toString()}` : "";
}

function normalizeMetrics(metrics: ApiFinanceMetrics): FinanceMetrics {
  return {
    grossSales: metrics.gross_sales,
    onlineSales: metrics.online_sales,
    posSales: metrics.pos_sales,
    totalExpenses: metrics.total_expenses,
    netEstimate: metrics.net_estimate,
    orderCount: metrics.order_count,
    posTransactionCount: metrics.pos_transaction_count,
    totalOrders: metrics.total_orders,
    totalPosTransactions: metrics.total_pos_transactions,
    averageOrderValue: metrics.average_order_value
  };
}

function normalizeDailyDay(day: ApiDailyReportDay): DailyReportDay {
  return {
    ...normalizeMetrics(day),
    date: day.date
  };
}

function normalizeMonthlyMonth(month: ApiMonthlyReportMonth): MonthlyReportMonth {
  return {
    ...normalizeMetrics(month),
    year: month.year,
    month: month.month
  };
}

function normalizeExpense(expense: ApiExpense): Expense {
  return {
    id: expense.id,
    categoryId: expense.category_id,
    category: expense.category,
    categoryName: expense.category_name,
    title: expense.title,
    amount: expense.amount,
    expenseDate: expense.expense_date,
    paymentMethod: expense.payment_method,
    note: expense.note,
    createdBy: expense.created_by,
    createdAt: expense.created_at,
    updatedAt: expense.updated_at
  };
}

function normalizePagination(meta?: ApiPaginationMeta): Pagination {
  return {
    limit: meta?.pagination?.limit ?? 20,
    nextCursor: meta?.pagination?.next_cursor ?? null,
    hasMore: meta?.pagination?.has_more ?? false
  };
}

function toExpensePayload(input: UpdateExpenseInput) {
  return {
    category: input.category ?? "",
    title: input.title,
    amount: input.amount,
    expense_date: input.expenseDate,
    payment_method: input.paymentMethod ?? "",
    note: input.note ?? ""
  };
}
