# Incident Response Ops Guide

This guide is the short operational version of `docs/incident_response_runbook_umkm_commerce_os_hybrid.md`.

## First response

Capture these facts before changing anything:

- What is broken?
- Which tenant, user, endpoint, order, product, or request is affected?
- When did it start?
- Was there a recent deploy or migration?
- Is data correctness at risk?
- Is there a `request_id` in the logs?

If data correctness or tenant isolation is at risk, treat it as high severity and consider stopping writes.

## App down

Check containers:

```bash
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml ps
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml logs --tail=200 api
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml logs --tail=200 frontend
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml logs --tail=200 caddy
```

Verify health:

```bash
./deploy/scripts/check-health.sh
```

If the API is stuck but the database is healthy:

```bash
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml restart api
```

## Database down

Check:

```bash
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml ps postgres
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml logs --tail=200 postgres
df -h
free -m
```

Do not delete Docker volumes. If disk is full, clean safe Docker images/logs first and preserve database volume data.

## Bad deploy rollback

For app-only rollback:

1. Set `API_IMAGE` and `FRONTEND_IMAGE` to previous known-good tags in `deploy/.env.production`.
2. Run:

   ```bash
   docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml pull
   docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml up -d
   ./deploy/scripts/check-health.sh
   ```

Prefer forward fixes for database migrations. Restore database only if data is damaged and a restore has been approved.

## Suspected tenant data leak

Treat as SEV-1.

1. Disable or block the affected endpoint if needed.
2. Preserve logs.
3. Identify affected tenants and data fields.
4. Patch the tenant filter or permission bug.
5. Run tenant isolation and permission tests:

   ```bash
   cd backend
   go test ./...
   ```

6. Document the incident and notify affected pilot tenants honestly.

Do not paste customer private data or secrets into tickets or chat.

## Checkout/POS stock inconsistency

Immediate checks:

```sql
SELECT *
FROM product_stock_snapshots
WHERE quantity_available <> quantity_on_hand - quantity_reserved;

SELECT *
FROM product_stock_snapshots
WHERE quantity_on_hand < 0
   OR quantity_reserved < 0
   OR quantity_available < 0;
```

Inspect source of truth:

```sql
SELECT *
FROM stock_movements
WHERE product_id = '<product_id>'
ORDER BY created_at ASC;
```

Repair preference:

- Use corrective stock movement when possible.
- Avoid silent stock edits unless you understand the full impact.
- Stop checkout/POS temporarily if corruption is spreading.

