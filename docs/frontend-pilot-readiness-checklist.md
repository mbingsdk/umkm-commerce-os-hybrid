# Frontend Pilot Readiness Checklist

Checklist ini dipakai untuk smoke test manual sebelum pilot UMKM. Fokusnya frontend, header request, UX state, dan batasan MVP.

## 1. Routes to manually test

### Public discovery / marketing

- `/`
- `/explore`
- `/stores`
- `/products`
- `/search?q=bouquet`
- `/category/bouquet`
- `/city/makassar`
- `/pricing`

### Public storefront

- `/s/{storeSlug}`
- `/s/{storeSlug}/products`
- `/s/{storeSlug}/categories/{categorySlug}`
- `/s/{storeSlug}/products/{productSlug}`
- `/s/{storeSlug}/about`
- `/s/{storeSlug}/contact`
- `/s/{storeSlug}/cart`
- `/s/{storeSlug}/checkout`
- `/s/{storeSlug}/orders/success?order_number={orderNumber}`
- `/s/{storeSlug}/orders/{orderNumber}/payment-confirmation`
- `/s/{storeSlug}/track-order`

### Tenant dashboard

- `/dashboard`
- `/dashboard/settings/store`
- `/dashboard/products`
- `/dashboard/categories`
- `/dashboard/inventory`
- `/dashboard/inventory/products/{productId}`
- `/dashboard/orders`
- `/dashboard/orders/{orderId}`
- `/dashboard/pos`
- `/dashboard/pos/history`
- `/dashboard/finance`
- `/dashboard/finance/expenses`
- `/dashboard/courier/zones`
- `/dashboard/shipments`
- `/dashboard/shipments/{shipmentId}`

### Super admin

- `/admin`
- `/admin/tenants`
- `/admin/tenants/{tenantId}`
- `/admin/plans`
- `/admin/discovery/featured`
- `/admin/audit-logs`

## 2. Public header checks

Public pages must not send:

- `Authorization`
- `X-Tenant-ID`

Check these flows in browser devtools Network tab:

- Storefront browse and product detail
- Cart and checkout submit
- Payment confirmation submit
- Tracking lookup
- Discovery pages and search
- Pricing page
- Category/city SEO pages
- Storefront about/contact pages

Allowed public headers include `Content-Type`, `Accept`, and `Idempotency-Key` for checkout/payment confirmation.

## 3. Dashboard/admin header checks

Dashboard tenant requests must send:

- `Authorization: Bearer ...`
- `X-Tenant-ID: ...`

Admin requests must send:

- `Authorization: Bearer ...`

Admin requests must not send:

- `X-Tenant-ID`

Check the API helpers:

- Dashboard modules should use `apiFetch()` default tenant-scoped behavior.
- Admin module should use `tenantScoped: false`.
- Public modules should use direct public fetch helpers or `auth: false` / `tenantScoped: false` if `apiFetch` is reused.

## 4. UI state checklist

Verify loading, empty, error, and forbidden states where applicable:

- Dashboard summary, products, categories, inventory, orders, POS, finance, courier, shipments, store settings
- Admin tenants, tenant detail, plans, featured discovery, audit logs
- Storefront home, product listing, product detail, cart, checkout, order success, payment confirmation, tracking
- Discovery home, stores, products, search, category, city

## 5. Critical submit checklist

Submit buttons must be disabled while pending:

- Login and register
- Create store
- Product/category create/update
- Stock adjustment and threshold update
- Checkout
- Public payment confirmation
- Order status, payment review, and cancel actions
- POS transaction, open session, close session
- Expense create/update/delete
- Courier zone create/update/delete
- Shipment create/status update
- Admin tenant status/plan, admin plan, featured discovery mutations

## 6. Mobile smoke checklist

- Storefront header navigation is reachable on mobile.
- Storefront home, product detail, cart, checkout, payment confirmation, and tracking are readable at mobile width.
- Dashboard mobile nav appears below `lg` and hides forbidden menu items.
- Desktop sidebar stays available on `lg` and above.
- POS product search and payment controls have large tap targets.
- Admin tables remain scannable using horizontal scroll or compact cards.

## 7. Copy and limitation checklist

- Indonesian copy is clear and user-facing.
- No visible sprint/scaffold/foundation labels.
- No fake/demo production copy.
- Manual payment is described honestly.
- Payment gateway, offline POS, AI, marketplace sync, and automatic payment confirmation are not promised as live MVP features.
- Public responses and UI do not expose `cost_price`, internal notes, tenant internals, audit logs, or private owner/customer data.

## 8. Known MVP frontend limitations

- Payment is manual transfer confirmation reviewed by the store owner.
- POS is online-first.
- Public discovery routes are SEO-friendly but do not include advanced ranking/recommendation.
- Storefront checkout is one store per transaction.
- Admin guard in frontend is UX only; backend remains the source of truth.
