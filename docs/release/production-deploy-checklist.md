# Production Deploy Checklist

Use this checklist only after staging has passed. Production here means a controlled pilot production environment, not broad public launch.

## 1. Preconditions

```txt
[ ] RELEASE_CHECKLIST.md is complete or exceptions are explicitly approved.
[ ] Staging deploy checklist passed on the same release candidate.
[ ] Backup and restore were tested on staging.
[ ] Rollback image tags or previous release commit are known.
[ ] Pilot tenants know this is an MVP pilot with manual payment and online-first POS.
```

## 2. Production security checks

```txt
[ ] SSH uses keys, not password login.
[ ] Firewall allows only SSH, HTTP, and HTTPS from the public internet.
[ ] PostgreSQL has no public host port.
[ ] TLS is active for frontend and API domains.
[ ] `CORS_ALLOWED_ORIGINS` is explicit; no wildcard credentials.
[ ] `JWT_SECRET`, database password, and admin credentials are strong and unique.
[ ] `.env.production` is not committed and has restricted file permissions.
[ ] Upload storage path and public URL are configured.
```

## 3. Backup before deploy

```bash
./deploy/scripts/backup-db.sh
```

Verify the file before changing the app:

```bash
gzip -t backups/database/daily/umkm_os_db_YYYYMMDD_HHMMSS.sql.gz
```

## 4. Deploy

```bash
git fetch --all --tags
git checkout v1.0.0-pilot
chmod +x deploy/scripts/*.sh
./deploy/scripts/deploy.sh
./deploy/scripts/migrate.sh
./deploy/scripts/check-health.sh
```

Check services:

```bash
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml ps
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml logs --tail=100 api worker frontend
```

Logs must not include request bodies, tokens, passwords, payment proof content, customer sensitive data, or financial notes.

## 5. Production smoke test

```txt
[ ] API live/ready health checks pass.
[ ] Frontend home page loads.
[ ] robots.txt and sitemap.xml load.
[ ] Login works for an approved pilot owner.
[ ] Dashboard rejects missing X-Tenant-ID.
[ ] Public storefront loads without Authorization.
[ ] Checkout creates an order with manual payment.
[ ] POS can open session and complete a small test transaction if pilot tenant agrees.
[ ] Admin /api/v1/admin/me works only for super_admin.
```

## 6. Rollback

For app-only rollback, change image tags or checkout the previous known-good commit, then:

```bash
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml pull
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml up -d
./deploy/scripts/check-health.sh
```

Do not restore the database unless data corruption has been confirmed and approved. Database restore is a separate incident-response action.

## 7. Production sign-off

```txt
[ ] Deploy operator recorded release tag/commit.
[ ] Backup path is recorded.
[ ] Health check time is recorded.
[ ] Pilot support contact is online.
[ ] Known limitations are visible to stakeholders.
```
