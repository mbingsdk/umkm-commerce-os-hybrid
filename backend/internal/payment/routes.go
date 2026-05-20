package payment

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
)

func RegisterPublicRoutes(r chi.Router, handler *Handler) {
	r.Post("/public/stores/{storeSlug}/orders/{orderNumber}/payment-confirmation", handler.PublicConfirm)
}

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	tenantMiddleware func(http.Handler) http.Handler,
	requirePermission func(permission.Permission) func(http.Handler) http.Handler,
) {
	r.Group(func(r chi.Router) {
		r.Use(tenantMiddleware)

		r.With(requirePermission(permission.OrderReadDetail)).Get("/orders/{orderId}/payment-confirmations", handler.ListConfirmations)
		r.With(requirePermission(permission.OrderUpdatePaymentStatus)).Post("/orders/{orderId}/confirm-payment", handler.ConfirmPayment)
		r.With(requirePermission(permission.OrderUpdatePaymentStatus)).Post("/orders/{orderId}/reject-payment", handler.RejectPayment)
	})
}
