package outbox

import (
	"context"
	"errors"
	"fmt"
)

var ErrUnknownEvent = errors.New("unknown outbox event type")

type Handler interface {
	Handle(ctx context.Context, event Event) error
}

type HandlerFunc func(ctx context.Context, event Event) error

func (fn HandlerFunc) Handle(ctx context.Context, event Event) error {
	return fn(ctx, event)
}

type Dispatcher struct {
	handlers map[string]Handler
}

func NewDispatcher(handlers map[string]Handler) *Dispatcher {
	copied := make(map[string]Handler, len(handlers))
	for eventType, handler := range handlers {
		copied[eventType] = handler
	}
	return &Dispatcher{handlers: copied}
}

func NewDefaultDispatcher() *Dispatcher {
	noop := NoopHandler{}

	return NewDispatcher(map[string]Handler{
		"OrderCreated":                   noop,
		"OrderStatusUpdated":             noop,
		"OrderCancelled":                 noop,
		"OrderPaid":                      noop,
		"PaymentConfirmed":               noop,
		"PaymentRejected":                noop,
		"StockReserved":                  noop,
		"StockReduced":                   noop,
		"StockAdjusted":                  noop,
		"StockReservationReleased":       noop,
		"CashierSessionOpened":           noop,
		"CashierSessionClosed":           noop,
		"POSTransactionCreated":          noop,
		"ExpenseChanged":                 noop,
		EventFinanceSummaryUpdateRequest: noop,
		EventNotificationRequested:       noop,
	})
}

func (d *Dispatcher) Handle(ctx context.Context, event Event) error {
	handler, ok := d.handlers[event.EventType]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownEvent, event.EventType)
	}

	return handler.Handle(ctx, event)
}

type NoopHandler struct{}

func (NoopHandler) Handle(context.Context, Event) error {
	return nil
}
