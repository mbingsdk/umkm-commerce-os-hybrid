# Sprint 11E Frontend Hardening QA Checklist

Use this checklist until a frontend test runner is introduced. Run it with the API and frontend pointed at a local seeded tenant.

## Request header safety

- Public storefront pages (`/s/[storeSlug]`, product detail, cart, checkout, success, tracking) must not send `Authorization` or `X-Tenant-ID`.
- Public discovery pages (`/`, `/explore`, `/stores`, `/products`, `/search`) must not send `Authorization` or `X-Tenant-ID`.
- Admin pages (`/admin/*`) must send `Authorization` and must not send `X-Tenant-ID`.
- Tenant dashboard pages (`/dashboard/*`) must send both `Authorization` and `X-Tenant-ID`.

## Dashboard states

- Dashboard home, products, categories, inventory, orders, POS, finance, courier zones, shipments, and admin pages show loading states while fetching.
- Empty list pages show clear empty states with safe next actions.
- Failed queries show retryable error states.
- Forbidden/permission-limited pages show a safe explanation and do not expose hidden data.

## Public page states

- Store not found or unpublished store renders not-found/error state without requiring login.
- Product not found renders not-found/error state without requiring login.
- Discovery empty search/list results show empty states.
- Checkout handles empty cart, wrong-store cart, validation errors, insufficient stock, and idempotency conflict.
- Tracking handles missing input, wrong phone/order pair, loading, and not-found/error states.

## Double-submit guards

- Login and register cannot submit twice while pending.
- Create store onboarding cannot submit twice while pending.
- Category/product create/update/delete actions cannot double-submit.
- Checkout cannot create duplicate submits while pending.
- Order status, payment review, and cancel dialogs cannot submit twice while pending.
- POS open/close session and transaction submit cannot submit twice while pending.
- Stock adjustment and threshold update cannot submit twice while pending.
- Expense create/update and courier/shipment/admin mutations cannot submit twice while pending.

## Permission-aware navigation and actions

- Tenant sidebar hides or disables finance, POS, courier, inventory, and sensitive actions according to the active tenant permissions.
- Admin UI is visually separated from tenant dashboard navigation.
- Frontend permission checks are treated as UX only; backend remains source of truth.

## Mobile and POS usability

- Storefront, product detail, cart, checkout, order success, and tracking are usable on small screens.
- Cart quantity controls are large enough for touch.
- POS product search is keyboard-friendly and focused on entry.
- POS product tiles have large tap targets.
- POS payment submit stays disabled for empty cart, underpaid cash, QRIS mismatch, missing permission, or pending submit.

