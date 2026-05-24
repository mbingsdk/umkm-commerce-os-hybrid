# UMKM Commerce OS Hybrid

Initial monorepo foundation for **UMKM Commerce OS Hybrid**, a multi-tenant SaaS commerce platform for Indonesian UMKM.

This repository currently contains only the Sprint 0 foundation:

- `backend/` -> Go modular-monolith scaffold
- `frontend/` -> Next.js App Router scaffold
- `docker-compose.yml` -> local PostgreSQL development service
- root tooling for common local commands

Business features such as auth, checkout, POS, payment confirmation, finance, discovery, and admin are intentionally **not implemented yet**.

## Prerequisites

- Go 1.23+
- Node.js 20+
- npm
- Docker Desktop or another Docker Compose-compatible runtime
- `make` if you want to use the convenience commands below

## Local setup

1. Copy the example environment files:

   ```powershell
   Copy-Item backend/.env.example backend/.env
   Copy-Item frontend/.env.example frontend/.env.local
   ```

2. Start PostgreSQL:

   ```powershell
   docker compose up -d postgres
   ```

3. Verify the backend scaffold:

   ```powershell
   cd backend
   go test ./...
   go run ./cmd/api
   ```

4. Install frontend dependencies and start the dev server:

   ```powershell
   cd frontend
   npm install
   npm run dev
   ```

5. Open the frontend at:

   ```txt
   http://localhost:3000
   ```

## Common commands

| Command | Purpose |
|---|---|
| `make db-up` | Start local PostgreSQL |
| `make db-down` | Stop local services |
| `make db-logs` | Follow PostgreSQL logs |
| `make backend-test` | Run Go tests |
| `make backend-build` | Compile backend scaffold |
| `make backend-run` | Run the backend placeholder entrypoint |
| `make frontend-install` | Install frontend dependencies |
| `make frontend-dev` | Start Next.js dev server |
| `make frontend-build` | Build frontend |
| `make frontend-lint` | Run frontend lint |
| `make frontend-typecheck` | Run TypeScript typecheck |

## Slow backend integration tests

Race-condition tests that depend on PostgreSQL row locks are guarded by the `integration` build tag and must use a disposable test database. The harness applies pending migrations and truncates public tables except `schema_migrations`.

```powershell
cd backend
$env:RUN_DB_INTEGRATION = "1"
$env:TEST_DATABASE_URL = "postgres://postgres:postgres@localhost:5432/umkm_os_test?sslmode=disable"
go test -tags=integration ./internal/integration -count=1
```

Daily lightweight backend tests remain:

```powershell
cd backend
go test ./...
```

## Frontend quality checks

Run these before a UI or frontend API-client change:

```powershell
cd frontend
npm.cmd run lint
npm.cmd run typecheck
npm.cmd run build
```

## Manual QA and API smoke collection

Sprint 11F adds manual QA scripts under `docs/qa/`:

| File | Purpose |
|---|---|
| `docs/qa/pilot-test-script.md` | End-to-end pilot flow from owner onboarding through storefront, checkout, POS, finance, courier, discovery, and admin suspend/activate. |
| `docs/qa/security-test-script.md` | Tenant isolation, permission matrix, admin guard, public data leak, upload validation, and rate-limit checks. |
| `docs/qa/race-test-script.md` | Last-stock checkout/POS, idempotency, payment/cancel race, outbox worker concurrency, and rollback checks. |
| `docs/qa/umkm-commerce-os.postman_collection.json` | Placeholder-based Postman collection for local/staging smoke tests. No real credentials are stored. |

Suggested local QA order:

```powershell
docker compose up -d postgres
cd backend
go run ./cmd/migrate up
go test ./...
cd ..
.\scripts\qa\seed-demo-data.ps1
```

Then import `docs/qa/umkm-commerce-os.postman_collection.json` into Postman and run folders in order:

```txt
Health -> Auth -> Tenant + Store -> Catalog -> Public Storefront + Checkout
-> Order + Payment -> Inventory + POS -> Finance -> Courier + Shipment
-> Discovery -> Admin
```

For frontend verification, run:

```powershell
cd frontend
npm.cmd run lint
npm.cmd run typecheck
npm.cmd run build
```

The demo seed script creates non-production data only:

```txt
Toko Bunga Ayu, Makassar, demo bouquet products, courier zones, and an open cashier session.
```

By default it refuses non-local URLs. Use `-AllowNonLocal` only for disposable staging.

## Query/index audit helpers

For slow endpoint investigation, run `EXPLAIN (ANALYZE, BUFFERS)` locally against a dev database with representative tenant/store data. Do not paste production customer data or full query parameter values into logs or tickets.

Recommended first-pass queries to inspect:

- public product listing/detail: products + categories + product_images + stock snapshots
- discovery stores/products/search: stores, tenants, products, categories, featured discovery
- dashboard product/order lists: tenant/store/status/search/date filters
- inventory stock list and POS product search
- finance summary: paid online orders, completed POS transactions, non-deleted expenses
- admin tenant list: tenants, plan, primary store, owner lookup, count snippets

Example local workflow:

```sql
EXPLAIN (ANALYZE, BUFFERS)
SELECT p.id, p.name, p.slug
FROM products p
WHERE p.tenant_id = '00000000-0000-0000-0000-000000000000'
  AND p.store_id = '00000000-0000-0000-0000-000000000000'
  AND p.status = 'active'
  AND p.deleted_at IS NULL
ORDER BY p.created_at DESC, p.id DESC
LIMIT 21;
```

When a slow request warning appears in API logs, use `request_id`, method, `path_template`, status, and `latency_ms` to choose the matching query family above.

## Repository structure

```txt
.
├─ backend/
│  ├─ cmd/
│  ├─ internal/
│  ├─ migrations/
│  ├─ seeds/
│  └─ tests/
├─ frontend/
│  ├─ app/
│  ├─ components/
│  ├─ features/
│  └─ lib/
├─ docker-compose.yml
├─ Makefile
└─ README.md
```

## Scope boundary for this foundation

The scaffold reserves clear places for later implementation, but it does not yet add:

- tenant/auth logic
- database migrations
- business repositories or services
- checkout/order/POS/payment flows
- finance/discovery/admin behavior

Those belong to the later sprints defined in `PLANS.md` and the project docs.
