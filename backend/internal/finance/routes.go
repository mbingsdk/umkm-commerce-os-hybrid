package finance

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	tenantMiddleware func(http.Handler) http.Handler,
	requirePermission func(permission.Permission) func(http.Handler) http.Handler,
) {
	r.Route("/finance", func(r chi.Router) {
		r.Use(tenantMiddleware)

		r.With(requirePermission(permission.FinanceReadSummary)).Get("/summary", handler.Summary)
		r.With(requirePermission(permission.FinanceReadReport)).Get("/reports/daily", handler.DailyReport)
		r.With(requirePermission(permission.FinanceReadReport)).Get("/reports/monthly", handler.MonthlyReport)
		r.With(requirePermission(permission.FinanceReadExpense)).Get("/expenses", handler.ListExpenses)
		r.With(requirePermission(permission.FinanceCreateExpense)).Post("/expenses", handler.CreateExpense)
		r.With(requirePermission(permission.FinanceUpdateExpense)).Patch("/expenses/{expenseId}", handler.UpdateExpense)
		r.With(requirePermission(permission.FinanceDeleteExpense)).Delete("/expenses/{expenseId}", handler.DeleteExpense)
	})
}
