# Sprint 11F Race and Idempotency Test Script

Use this script for manual/integration verification of stock locks, idempotency, payment/cancel races, and outbox worker concurrency.

## Preconditions

- Use a disposable PostgreSQL database.
- Run migrations from a clean state.
- Seed two tenants if isolation is involved.
- Use unique idempotency keys unless explicitly testing same-key replay.

## 1. Checkout last stock race

Setup:

- Product stock available = `1`.
- Store is published.
- Product is active and visible.

Run:

- Send two concurrent `POST /api/v1/public/stores/{storeSlug}/checkout` requests.
- Use different `Idempotency-Key` values.
- Both request quantity `1`.

Expected:

- Exactly one checkout succeeds.
- Exactly one checkout fails with `INSUFFICIENT_STOCK` or equivalent domain error.
- Final stock snapshot:
  - `quantity_reserved` increased once
  - `quantity_available` is `0`
- Only one active order/reservation exists for the product.

## 2. Checkout idempotency

Same payload:

- Send checkout request with key `qa-checkout-same-key`.
- Send the exact same payload with same key again.

Expected:

- Same completed response is returned.
- Only one order is created.

Different payload:

- Reuse key `qa-checkout-conflict-key`.
- Change quantity/customer payload.

Expected:

- Response is `IDEMPOTENCY_CONFLICT`.
- No second order is created.

## 3. POS last stock race

Setup:

- Product stock available = `1`.
- Cashier session is open.

Run:

- Send two concurrent `POST /api/v1/pos/transactions` requests with different idempotency keys.
- Both request quantity `1`.

Expected:

- Exactly one transaction succeeds.
- Exactly one transaction fails due to insufficient stock or equivalent domain error.
- Final stock snapshot:
  - `quantity_on_hand` decreased once
  - `quantity_available` is `0`
- Exactly one `stock_movements` row with type `pos_sale` is created.

## 4. POS idempotency

Same payload:

- Send POS transaction with key `qa-pos-same-key`.
- Send exact same payload and key again.

Expected:

- Same transaction response is replayed.
- Only one POS transaction exists.

Different payload:

- Reuse key `qa-pos-conflict-key`.
- Change item quantity or amount paid.

Expected:

- Response is `IDEMPOTENCY_CONFLICT`.
- No second transaction is created.

## 5. Payment confirm double request

Setup:

- Order is pending payment.
- One pending payment confirmation exists.

Run:

- Send two confirm-payment requests for the same order/confirmation as close together as possible.

Expected:

- Payment is marked paid once.
- Order status moves to paid once.
- Payment/log/outbox records remain consistent.
- Stock is not reduced/released twice.

## 6. Payment confirm vs cancel

Setup:

- Order is pending payment with reserved stock.

Run concurrently:

- Confirm payment.
- Cancel order.

Expected:

- Only one final state wins.
- Losing request returns invalid transition/conflict.
- Reservation status remains consistent.
- Stock snapshot remains consistent.
- No double stock release/reduction.

## 7. Outbox worker concurrency

Setup:

- Create multiple pending outbox events.

Run:

- Start two worker processes or call polling concurrently.

Expected:

- `FOR UPDATE SKIP LOCKED` or equivalent prevents the same event from being processed twice.
- Succeeded events are marked completed once.
- Failed events increment attempts safely.

## 8. Rollback checks

For checkout, POS transaction, payment confirmation, and cancel:

- Force a failure after partial writes if practical in integration tests.

Expected:

- Transaction rolls back.
- No partial order/item/reservation/payment/stock movement state remains.
- Outbox event is not inserted unless the domain write commits.

