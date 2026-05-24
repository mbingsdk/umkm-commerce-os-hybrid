# Staging Deploy Checklist

Use staging as the rehearsal space for `v1.0.0-pilot`. It should be close to production, but disposable enough that destructive restore and race tests are safe.

## 1. Environment readiness

```txt
[ ] Staging VPS is separate from production.
[ ] Staging database is separate from production.
[ ] Staging domain/subdomain is configured.
[ ] Caddy or Nginx points to the staging frontend/API containers.
[ ] PostgreSQL is not exposed to the public internet.
[ ] `deploy/.env.production` exists only on the server and contains staging values.
[ ] No real production secrets are copied into staging.
```

Recommended staging domains:

```txt
APP_DOMAIN=staging.example.com
API_DOMAIN=api-staging.example.com
```

## 2. Pre-deploy checks

```txt
[ ] `go test ./...` passes in `backend/`.
[ ] Frontend lint/typecheck/build pass in `frontend/`.
[ ] `docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml config` succeeds.
[ ] Backup script is available and executable.
[ ] Rollback image tags or previous commit SHA are known.
```

## 3. Deploy rehearsal

```bash
chmod +x deploy/scripts/*.sh
./deploy/scripts/backup-db.sh
./deploy/scripts/deploy.sh
./deploy/scripts/migrate.sh
./deploy/scripts/check-health.sh
```

Confirm containers:

```bash
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml ps
```

## 4. Smoke tests

```txt
[ ] GET /health/live returns healthy.
[ ] GET /health/ready returns healthy.
[ ] Owner can login.
[ ] Tenant dashboard loads with X-Tenant-ID.
[ ] Public storefront loads without login.
[ ] Public discovery loads without login.
[ ] Checkout smoke flow creates one order.
[ ] POS smoke flow creates one transaction.
[ ] Worker starts and logs outbox handling safely.
```

Run the Postman collection from `docs/qa/umkm-commerce-os.postman_collection.json` with staging variables. Do not save real tokens in the collection file.

## 5. Restore rehearsal

```txt
[ ] Latest backup file exists.
[ ] `gzip -t` passes for the backup.
[ ] Restore is tested on staging only.
[ ] App health passes after restore.
```

Example:

```bash
./deploy/scripts/restore-db.sh --yes backups/database/daily/umkm_os_db_YYYYMMDD_HHMMSS.sql.gz
./deploy/scripts/check-health.sh
```

## 6. Staging sign-off

```txt
[ ] Release checklist critical items are complete.
[ ] Known limitations have been reviewed.
[ ] Pilot support owner is assigned.
[ ] No production secrets or real customer data were committed.
```
