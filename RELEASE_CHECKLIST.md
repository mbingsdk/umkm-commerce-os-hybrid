# Release Checklist - v1.0.0-pilot

Use this checklist before tagging or deploying the pilot release candidate.

## Code quality

```txt
[ ] Backend unit/module tests pass: cd backend && go test ./...
[ ] Frontend lint passes: cd frontend && npm.cmd run lint
[ ] Frontend typecheck passes: cd frontend && npm.cmd run typecheck
[ ] Frontend production build passes: cd frontend && npm.cmd run build
[ ] No unexpected generated files or binaries are committed.
```

## Database and migrations

```txt
[ ] Migration from empty database succeeds.
[ ] Migration from existing staging database succeeds.
[ ] Migration status is clean after deploy.
[ ] No applied production migration was edited in-place.
[ ] Backup completed before migration.
```

Suggested commands:

```powershell
cd backend
go run ./cmd/migrate up
go run ./cmd/migrate status
```

## Demo seed

```txt
[ ] Demo seed runs on local/staging only.
[ ] APP_ENV=production blocks demo seed by default.
[ ] Toko Bunga Ayu demo tenant/store is created.
[ ] Demo products, stock snapshots, stock movements, and courier zones exist.
[ ] Re-running seed does not duplicate users/products/slugs/initial movements.
```

Suggested command:

```powershell
cd backend
go run ./cmd/seed
```

## API smoke and QA

```txt
[ ] API collection smoke test passes using docs/qa/umkm-commerce-os.postman_collection.json.
[ ] Pilot manual flow passes using docs/qa/pilot-test-script.md.
[ ] Security manual flow passes using docs/qa/security-test-script.md.
[ ] Race/manual concurrency flow is reviewed using docs/qa/race-test-script.md.
```

## Security tests

```txt
[ ] Tenant isolation tests pass.
[ ] Permission matrix tests pass.
[ ] Admin route guard tests pass.
[ ] Public routes do not expose cost_price/internal fields.
[ ] Upload invalid MIME and oversized file tests pass.
[ ] Login/checkout rate limit tests pass.
```

## Race/integration tests

Use a disposable PostgreSQL test database only.

```powershell
cd backend
$env:RUN_DB_INTEGRATION = "1"
$env:TEST_DATABASE_URL = "postgres://postgres:postgres@localhost:5432/umkm_os_test?sslmode=disable"
go test -tags=integration ./internal/integration -count=1 -v
```

Checklist:

```txt
[ ] Checkout last-stock race: only one request succeeds.
[ ] POS last-stock race: only one transaction succeeds.
[ ] Checkout/POS idempotency replay/conflict behavior passes.
[ ] Payment confirm double request does not double-apply effects.
[ ] Payment confirm vs cancel race leaves one consistent final state.
[ ] Outbox worker concurrency prevents double processing.
```

## Deployment dry run

```txt
[ ] deploy/.env.production.example copied to staging env and placeholders replaced.
[ ] docker compose production config validates.
[ ] Images build locally or pull from registry.
[ ] API, worker, frontend, postgres, and Caddy start.
[ ] PostgreSQL is not exposed publicly.
[ ] API /health/live and /health/ready pass.
```

Suggested command:

```bash
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml config
```

## Backup, restore, and rollback

```txt
[ ] Manual backup script creates non-empty gzip file.
[ ] gzip -t backup file passes.
[ ] Restore tested on staging, not directly on production.
[ ] Application rollback tested by switching image tags.
[ ] Restore procedure requires explicit backup file and confirmation.
```

Docs:

```txt
docs/ops/backup-restore.md
docs/release/staging-deploy-checklist.md
docs/release/production-deploy-checklist.md
```

## Public SEO smoke test

```txt
[ ] /robots.txt loads.
[ ] /sitemap.xml loads.
[ ] Sitemap does not include /dashboard, /admin, login, checkout, cart, order success, or tracking pages.
[ ] Store page has title/description/Open Graph metadata.
[ ] Product detail has title/description/Open Graph metadata.
[ ] Product JSON-LD is valid and safe.
[ ] Store JSON-LD is valid and safe.
[ ] Public responses never include cost_price.
```

## Release sign-off

```txt
[ ] Known limitations are reviewed with pilot stakeholders.
[ ] Pilot support channel is ready.
[ ] Incident response playbook is ready.
[ ] Backup/restore procedure has an owner.
[ ] v1.0.0-pilot release notes are reviewed.
```

