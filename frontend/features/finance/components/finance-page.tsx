"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { EmptyState } from "@/components/feedback/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useDailyReport, useFinanceSummary, useMonthlyReport } from "@/features/finance/hooks/use-finance";
import type { DailyReportDay, FinanceMetrics, FinanceSummary, MonthlyReportMonth } from "@/features/finance/types";
import { formatDate } from "@/lib/format/date";
import { formatRupiah } from "@/lib/format/money";
import { permissions } from "@/lib/permissions/permissions";
import { useTenantStore } from "@/lib/stores/tenant.store";

export function FinancePage() {
  const today = useMemo(() => new Date(), []);
  const [dateFrom, setDateFrom] = useState(() => firstDayOfMonth(today));
  const [dateTo, setDateTo] = useState(() => toISODate(today));
  const [year, setYear] = useState(today.getFullYear());
  const [month, setMonth] = useState<number | "">(today.getMonth() + 1);
  const userPermissions = useTenantStore((state) => state.permissions);
  const canReadSummary = userPermissions.includes(permissions.financeReadSummary);
  const canReadReport = userPermissions.includes(permissions.financeReadReport);
  const canReadExpense = userPermissions.includes(permissions.financeReadExpense);
  const filters = useMemo(() => ({ dateFrom, dateTo }), [dateFrom, dateTo]);
  const monthlyFilters = useMemo(() => ({ year, month }), [month, year]);
  const summaryQuery = useFinanceSummary(filters, canReadSummary);
  const dailyReportQuery = useDailyReport(filters, canReadReport);
  const monthlyReportQuery = useMonthlyReport(monthlyFilters, canReadReport);

  if (!canReadSummary && !canReadReport) {
    return (
      <EmptyState
        title="Akses finance belum tersedia"
        description="Role aktifmu belum memiliki izin untuk melihat ringkasan atau laporan finance."
      />
    );
  }

  const isInitialLoading =
    (canReadSummary && summaryQuery.isPending) || (canReadReport && dailyReportQuery.isPending && monthlyReportQuery.isPending);

  if (isInitialLoading) {
    return <LoadingState lines={5} />;
  }

  const hasBlockingError = (canReadSummary && summaryQuery.isError) || (canReadReport && dailyReportQuery.isError);
  if (hasBlockingError) {
    return (
      <ErrorState
        title="Finance belum bisa dimuat"
        description="Coba muat ulang laporan finance atau periksa filter tanggal."
        onRetry={() => {
          void summaryQuery.refetch();
          void dailyReportQuery.refetch();
          void monthlyReportQuery.refetch();
        }}
      />
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-neutral-950">Keuangan</h1>
          <p className="mt-1 text-sm text-neutral-500">
            Ringkasan penjualan, pengeluaran, dan estimasi net sederhana untuk tenant aktif.
          </p>
        </div>
        {canReadExpense ? (
          <LinkButton href="/dashboard/finance/expenses">Kelola pengeluaran</LinkButton>
        ) : null}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Filter periode</CardTitle>
          <CardDescription>Pilih rentang tanggal untuk summary dan laporan harian.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-3 md:grid-cols-[1fr_1fr_auto]">
          <Input type="date" value={dateFrom} onChange={(event) => setDateFrom(event.target.value)} />
          <Input type="date" value={dateTo} onChange={(event) => setDateTo(event.target.value)} />
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              setDateFrom(firstDayOfMonth(today));
              setDateTo(toISODate(today));
            }}
          >
            Bulan ini
          </Button>
        </CardContent>
      </Card>

      {canReadSummary && summaryQuery.data ? <FinanceSummaryCards summary={summaryQuery.data} /> : null}

      {canReadSummary && summaryQuery.data ? <SalesBreakdown metrics={summaryQuery.data} /> : null}

      {canReadReport ? (
        <div className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
          <DailyReportCard days={dailyReportQuery.data?.days ?? []} />
          <MonthlyReportCard
            year={year}
            month={month}
            onYearChange={setYear}
            onMonthChange={setMonth}
            months={monthlyReportQuery.data?.months ?? []}
            isPending={monthlyReportQuery.isPending}
            isError={monthlyReportQuery.isError}
            onRetry={() => void monthlyReportQuery.refetch()}
          />
        </div>
      ) : null}
    </div>
  );
}

function FinanceSummaryCards({ summary }: { summary: FinanceSummary }) {
  const cards = [
    { label: "Gross sales", value: formatRupiah(summary.grossSales), description: "Online paid + POS completed." },
    { label: "Online sales", value: formatRupiah(summary.onlineSales), description: `${summary.orderCount} order paid.` },
    {
      label: "POS sales",
      value: formatRupiah(summary.posSales),
      description: `${summary.posTransactionCount} transaksi POS.`
    },
    { label: "Pengeluaran", value: formatRupiah(summary.totalExpenses), description: "Expense aktif, belum ledger." },
    { label: "Net estimate", value: formatRupiah(summary.netEstimate), description: "Estimasi, bukan laba akuntansi." },
    { label: "AOV", value: formatRupiah(summary.averageOrderValue), description: "Rata-rata nilai transaksi." }
  ];

  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      {cards.map((card) => (
        <Card key={card.label}>
          <CardHeader className="pb-2">
            <CardDescription>{card.label}</CardDescription>
            <CardTitle className="text-2xl">{card.value}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs leading-5 text-neutral-500">{card.description}</p>
          </CardContent>
        </Card>
      ))}
      <Card className="border-amber-200 bg-amber-50 shadow-none sm:col-span-2 xl:col-span-3">
        <CardContent>
          <p className="text-sm font-semibold text-amber-900">Catatan penting</p>
          <p className="mt-1 text-sm leading-6 text-amber-800">
            Net estimate belum menghitung HPP/modal detail, pajak, refund, payout, atau settlement payment gateway.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}

function SalesBreakdown({ metrics }: { metrics: FinanceMetrics }) {
  const rows = [
    { label: "Online", value: metrics.onlineSales },
    { label: "POS", value: metrics.posSales },
    { label: "Expense", value: metrics.totalExpenses }
  ];
  const max = Math.max(...rows.map((row) => row.value), 1);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Komposisi sales & expense</CardTitle>
        <CardDescription>Chart ringan berbasis bar sederhana.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {rows.map((row) => (
          <div key={row.label} className="space-y-2">
            <div className="flex items-center justify-between gap-3 text-sm">
              <span className="font-medium text-neutral-700">{row.label}</span>
              <span className="font-semibold text-neutral-950">{formatRupiah(row.value)}</span>
            </div>
            <div className="h-3 overflow-hidden rounded-full bg-neutral-100">
              <div
                className={row.label === "Expense" ? "h-full rounded-full bg-amber-500" : "h-full rounded-full bg-primary-600"}
                style={{ width: row.value === 0 ? "0%" : `${Math.max(6, (row.value / max) * 100)}%` }}
              />
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}

function DailyReportCard({ days }: { days: DailyReportDay[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Laporan harian</CardTitle>
        <CardDescription>Ringkasan per hari dari rentang tanggal terpilih.</CardDescription>
      </CardHeader>
      <CardContent>
        {days.length === 0 ? (
          <p className="rounded-xl border border-dashed border-neutral-200 p-4 text-sm text-neutral-500">
            Belum ada data laporan harian pada periode ini.
          </p>
        ) : (
          <div className="space-y-3">
            {days.slice(0, 10).map((day) => (
              <div key={day.date} className="rounded-xl border border-neutral-200 p-3">
                <div className="flex items-center justify-between gap-3">
                  <p className="font-semibold text-neutral-950">{formatDate(day.date)}</p>
                  <p className="font-semibold text-neutral-950">{formatRupiah(day.netEstimate)}</p>
                </div>
                <p className="mt-1 text-xs text-neutral-500">
                  Sales {formatRupiah(day.grossSales)} • Expense {formatRupiah(day.totalExpenses)}
                </p>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function MonthlyReportCard({
  year,
  month,
  onYearChange,
  onMonthChange,
  months,
  isPending,
  isError,
  onRetry
}: {
  year: number;
  month: number | "";
  onYearChange: (value: number) => void;
  onMonthChange: (value: number | "") => void;
  months: MonthlyReportMonth[];
  isPending: boolean;
  isError: boolean;
  onRetry: () => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Laporan bulanan</CardTitle>
        <CardDescription>Gunakan tahun/bulan untuk melihat agregat sederhana.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-3 sm:grid-cols-2">
          <Input
            type="number"
            value={year}
            onChange={(event) => onYearChange(Number(event.target.value) || new Date().getFullYear())}
          />
          <select
            className="h-10 rounded-xl border border-neutral-300 bg-white px-3 text-sm text-neutral-950 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-100"
            value={month}
            onChange={(event) => onMonthChange(event.target.value === "" ? "" : Number(event.target.value))}
          >
            <option value="">Semua bulan</option>
            {Array.from({ length: 12 }).map((_, index) => (
              <option key={index + 1} value={index + 1}>
                {new Intl.DateTimeFormat("id-ID", { month: "long" }).format(new Date(year, index, 1))}
              </option>
            ))}
          </select>
        </div>
        {isPending ? (
          <LoadingState lines={2} />
        ) : isError ? (
          <ErrorState title="Laporan bulanan gagal dimuat" description="Coba muat ulang kartu ini." onRetry={onRetry} />
        ) : months.length === 0 ? (
          <p className="rounded-xl border border-dashed border-neutral-200 p-4 text-sm text-neutral-500">
            Belum ada data laporan bulanan.
          </p>
        ) : (
          <div className="space-y-3">
            {months.map((item) => (
              <div key={`${item.year}-${item.month}`} className="rounded-xl border border-neutral-200 p-3">
                <div className="flex items-center justify-between gap-3">
                  <p className="font-semibold text-neutral-950">
                    {new Intl.DateTimeFormat("id-ID", { month: "long", year: "numeric" }).format(
                      new Date(item.year, item.month - 1, 1)
                    )}
                  </p>
                  <p className="font-semibold text-neutral-950">{formatRupiah(item.netEstimate)}</p>
                </div>
                <p className="mt-1 text-xs text-neutral-500">
                  Sales {formatRupiah(item.grossSales)} • Expense {formatRupiah(item.totalExpenses)}
                </p>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function LinkButton({ href, children }: { href: string; children: string }) {
  return (
    <Link
      href={href}
      className="inline-flex h-10 items-center justify-center rounded-xl bg-primary-600 px-4 text-sm font-semibold text-white transition hover:bg-primary-700"
    >
      {children}
    </Link>
  );
}

function toISODate(value: Date) {
  return value.toISOString().slice(0, 10);
}

function firstDayOfMonth(value: Date) {
  return `${value.getFullYear()}-${String(value.getMonth() + 1).padStart(2, "0")}-01`;
}
