# UMKM Commerce OS Hybrid

**UMKM Commerce OS Hybrid** is a multi-tenant SaaS commerce platform for Indonesian UMKM. The MVP pilot combines tenant storefronts, platform discovery, online checkout, inventory, order management, POS/kasir, basic finance, local courier, and super admin operations in one monorepo.

This repository contains the pilot-ready application foundation and release assets:

- `backend/` - Go modular monolith with PostgreSQL, chi, pgx, slog, idempotency, and outbox foundations.
- `frontend/` - Next.js App Router with TypeScript, Tailwind CSS, TanStack Query, and Zustand.
- `deploy/` - Docker Compose, Caddy, migration, backup, restore, and health-check assets for VPS deployment.
- `docs/qa/`, `docs/ops/`, `docs/pilot/`, and `docs/release/` - QA scripts, operations runbooks, pilot guides, and release checklists.

Pilot scope is intentionally controlled: manual payment, online-first POS, one-store checkout, basic finance estimates, basic courier, and no marketplace sync or payment gateway yet.

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

3. Run backend checks, migrations, and the API:

   ```powershell
   cd backend
   go test ./...
   go run ./cmd/migrate up
   # Optional non-production demo seed
   go run ./cmd/seed
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
| `make backend-run` | Run the backend API locally |
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

## Release candidate docs

Sprint 12E release readiness lives in these files:

| File | Purpose |
|---|---|
| `RELEASE_CHECKLIST.md` | Final release candidate gate covering tests, migrations, seed, smoke tests, deployment, backup/restore, rollback, and SEO. |
| `RELEASE_NOTES.md` | `v1.0.0-pilot` release note, pilot-ready areas, and operational notes. |
| `docs/release/staging-deploy-checklist.md` | Staging deployment rehearsal checklist. |
| `docs/release/production-deploy-checklist.md` | Controlled pilot production deployment checklist. |
| `docs/release/pilot-go-live-checklist.md` | Tenant onboarding and go-live readiness checklist. |
| `docs/release/known-limitations.md` | Explicit MVP limitations: manual payment, no offline POS, no payment gateway, no marketplace sync, no custom domain, and finance estimate caveats. |
| `docs/release/pilot-support-playbook.md` | Support triage, tenant feedback, stock/order inconsistency handling, and safe tenant pause flow. |

## Staging and pilot production deployment

Sprint 12A adds Docker Compose deployment assets under `deploy/` for a small VPS-based staging or pilot production environment. Caddy is the preferred MVP reverse proxy because it manages TLS certificates automatically.

### Required VPS packages

Install the minimum runtime tools on an Ubuntu/Debian VPS:

```bash
sudo apt update
sudo apt install -y ca-certificates curl git
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker "$USER"
```

Log out and back in after adding the Docker group. Then verify:

```bash
docker version
docker compose version
```

### Required environment variables

Copy the production env example and replace all placeholders:

```bash
cp deploy/.env.production.example deploy/.env.production
chmod 600 deploy/.env.production
```

Required values include:

- `APP_DOMAIN`, `API_DOMAIN`, and `ACME_EMAIL`
- `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`
- `DATABASE_URL`
- `JWT_SECRET`
- `CORS_ALLOWED_ORIGINS`
- `STORAGE_PUBLIC_URL`
- `NEXT_PUBLIC_API_BASE_URL` and `NEXT_PUBLIC_SITE_URL`

Do not commit `deploy/.env.production`. PostgreSQL is intentionally not exposed with a public host port in `deploy/docker-compose.prod.yml`.

### Build or push images

For a simple VPS pilot, Compose can build images directly from the repository:

```bash
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml build
```

If using a registry, set `API_IMAGE` and `FRONTEND_IMAGE` in `deploy/.env.production`, then push from CI or your workstation:

```bash
docker build -t "$API_IMAGE" ./backend
docker build \
  --build-arg NEXT_PUBLIC_API_BASE_URL="https://api.example.com" \
  --build-arg NEXT_PUBLIC_SITE_URL="https://app.example.com" \
  -t "$FRONTEND_IMAGE" ./frontend
docker push "$API_IMAGE"
docker push "$FRONTEND_IMAGE"
```

### First deploy

From the repository root on the VPS:

```bash
chmod +x deploy/scripts/*.sh
./deploy/scripts/deploy.sh
```

Run migrations after PostgreSQL is healthy:

```bash
./deploy/scripts/migrate.sh
```

Then check health:

```bash
./deploy/scripts/check-health.sh
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml ps
```

### Backups and restore

Create a compressed PostgreSQL backup before every deploy and daily during pilot:

```bash
./deploy/scripts/backup-db.sh
```

Restore is intentionally guarded:

```bash
./deploy/scripts/restore-db.sh backups/database/daily/umkm_os_db_YYYYMMDD_HHMMSS.sql.gz
```

Use `--yes` only for controlled non-interactive emergency automation after approval.

Always test restore on staging before touching production. Uploaded files are stored in the named Docker volume `uploads_data`; include that volume in VPS backup planning if local upload storage is used.

### Rollback basics

For application-only rollback:

1. Set `API_IMAGE` and `FRONTEND_IMAGE` in `deploy/.env.production` to the previous known-good tags.
2. Run:

   ```bash
   docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml pull
   docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml up -d
   ./deploy/scripts/check-health.sh
   ```

Do not restore the database for an app rollback unless the incident runbook says data is damaged and a restore has been approved.

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

When a slow request warning appears in API logs, use `request_id`, method, `path_template`, status, and `duration_ms` to choose the matching query family above.

## Repository structure

```txt
.
|-- backend/
|   |-- cmd/
|   |-- internal/
|   |-- migrations/
|   |-- seeds/
|   `-- tests/
|-- deploy/
|-- docs/
|   |-- ops/
|   |-- pilot/
|   |-- qa/
|   `-- release/
|-- frontend/
|   |-- app/
|   |-- components/
|   |-- features/
|   `-- lib/
|-- docker-compose.yml
|-- Makefile
`-- README.md
```

## MVP pilot scope and limitations

The current release candidate is intended for controlled pilot usage. It includes the core MVP flows, release documentation, and deployment assets, but it is not a broad public production launch.

Important limitations are tracked in `docs/release/known-limitations.md`, including:

- manual payment only
- no offline POS
- no payment gateway
- no marketplace sync
- no custom domain support
- finance `net_estimate` excludes detailed HPP/COGS

Do not put production secrets or real customer data in repository docs, QA scripts, or committed environment files.
