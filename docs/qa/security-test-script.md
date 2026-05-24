# Sprint 11F Security Test Script

Security QA script for tenant isolation, permission, public data safety, upload hardening, and rate-limit checks.

## Safety rules

- Use only demo tenants and demo users.
- Do not paste real passwords, tokens, customer addresses, or production data into reports.
- Capture `request_id`, endpoint, role, and expected/actual result for every failure.

## 1. Tenant isolation

Create:

- Tenant A: `Toko Bunga Ayu`
- Tenant B: `Toko Kue Rani`
- Owner A, Owner B
- Product/order/inventory/finance/courier data in each tenant

Checks:

- Tenant A cannot read/update/delete Tenant B categories.
- Tenant A cannot read/update/delete Tenant B products.
- Tenant A cannot read Tenant B inventory stock or movements.
- Tenant A cannot adjust Tenant B product stock.
- Tenant A cannot list/read/cancel Tenant B orders.
- Tenant A cannot access Tenant B POS sessions/transactions.
- Tenant A cannot read Tenant B finance summary/expenses.
- Tenant A cannot update Tenant B courier zones or shipments.
- Tenant A dashboard summary/recent orders/low stock only show Tenant A data.

Expected:

- Cross-tenant dashboard attempts return 401/403/404 with safe error.
- No response includes business data from the wrong tenant.

## 2. Permission matrix

Use active users in Tenant A:

- owner
- manager
- staff
- cashier

Checks:

- Owner can perform sensitive tenant actions: product mutation, stock adjustment, payment review, POS, finance, courier.
- Manager can perform only actions allowed by the matrix.
- Staff cannot access finance/admin-only actions.
- Cashier can access POS actions allowed by the matrix and cannot access finance/order-sensitive actions if not allowed.
- Super admin cannot access tenant dashboard unless also a tenant member, if that remains the intended policy.

Expected:

- Backend permission middleware is source of truth.
- Frontend `PermissionGate` only hides/disables UX.

## 3. Admin route guard

Checks:

- No token -> `/api/v1/admin/me` rejected.
- Tenant owner token -> `/api/v1/admin/me` rejected.
- Tenant manager/staff/cashier token -> `/api/v1/admin/*` rejected.
- Super admin token -> `/api/v1/admin/me` accepted.
- Admin requests do not need `X-Tenant-ID`.
- Adding `X-Tenant-ID` to admin requests does not change authorization semantics.

Expected:

- Only platform role `super_admin` can access `/api/v1/admin/*`.

## 4. Public route data leak checks

Public endpoints to test without auth:

- `/api/v1/public/stores/{storeSlug}`
- `/api/v1/public/stores/{storeSlug}/categories`
- `/api/v1/public/stores/{storeSlug}/products`
- `/api/v1/public/stores/{storeSlug}/products/{productSlug}`
- `/api/v1/public/discovery/home`
- `/api/v1/public/discovery/stores`
- `/api/v1/public/discovery/products`
- `/api/v1/public/discovery/search`
- `/api/v1/public/stores/{storeSlug}/orders/{orderNumber}/tracking?phone=...`

Checks:

- Unpublished store is hidden.
- Suspended/cancelled tenant is hidden.
- Draft/inactive product is hidden.
- Non-discoverable product is hidden from discovery.
- Public product/discovery responses never include:
  - `cost_price`
  - tenant internal status
  - audit logs
  - owner private data
  - internal notes
  - local filesystem paths

Expected:

- Public endpoints do not require `Authorization` or `X-Tenant-ID`.
- Public endpoints only expose safe DTO fields.

## 5. Upload validation

Checks:

- Valid JPEG under size limit accepted.
- Valid PNG under size limit accepted.
- Valid WebP under size limit accepted.
- Oversized image rejected.
- Text file renamed as `.jpg` rejected.
- Path traversal filename like `../../x.jpg` cannot affect storage path.
- API response does not expose local filesystem path.

Expected:

- MIME sniffing and configured max size are enforced server-side.

## 6. Rate limit checks

Endpoints:

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/register`
- `POST /api/v1/public/stores/{storeSlug}/checkout`
- `POST /api/v1/pos/transactions`

Checks:

- Repeated invalid login eventually returns rate-limit error.
- Repeated register attempts eventually return rate-limit error.
- Repeated checkout attempts from same client are bounded.
- POS transaction endpoint remains protected by auth, tenant, permission, idempotency, and rate-limit where configured.

Expected:

- Rate limits trigger without logging passwords, tokens, request bodies, or sensitive customer data.

