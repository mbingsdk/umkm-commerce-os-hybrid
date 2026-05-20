export type Pagination = {
  limit: number;
  nextCursor?: string | null;
  hasMore: boolean;
};

export type FinanceMetrics = {
  grossSales: number;
  onlineSales: number;
  posSales: number;
  totalExpenses: number;
  netEstimate: number;
  orderCount: number;
  posTransactionCount: number;
  totalOrders: number;
  totalPosTransactions: number;
  averageOrderValue: number;
};

export type FinanceSummary = FinanceMetrics & {
  dateFrom: string;
  dateTo: string;
  note: string;
};

export type DailyReportDay = FinanceMetrics & {
  date: string;
};

export type DailyReport = {
  dateFrom: string;
  dateTo: string;
  days: DailyReportDay[];
  summary: FinanceMetrics;
  note: string;
};

export type MonthlyReportMonth = FinanceMetrics & {
  year: number;
  month: number;
};

export type MonthlyReport = {
  year: number;
  month?: number | null;
  months: MonthlyReportMonth[];
  summary: FinanceMetrics;
  note: string;
};

export type Expense = {
  id: string;
  categoryId?: string | null;
  category?: string;
  categoryName?: string;
  title: string;
  amount: number;
  expenseDate: string;
  paymentMethod?: string;
  note?: string;
  createdBy?: {
    id: string;
    name?: string;
  } | null;
  createdAt: string;
  updatedAt: string;
};

export type FinanceDateFilters = {
  dateFrom?: string;
  dateTo?: string;
};

export type MonthlyReportFilters = {
  year?: number;
  month?: number | "";
};

export type ExpenseFilters = FinanceDateFilters & {
  query?: string;
  category?: string;
  cursor?: string | null;
  limit?: number;
};

export type ListExpensesResult = {
  expenses: Expense[];
  pagination: Pagination;
};

export type CreateExpenseInput = {
  category?: string;
  title: string;
  amount: number;
  expenseDate: string;
  paymentMethod?: string;
  note?: string;
};

export type UpdateExpenseInput = Partial<CreateExpenseInput>;
