# Logging Ops Guide

This guide defines what is safe and useful to log for staging and pilot production.

## Request ID usage

Every API request receives an `X-Request-ID` response header. If a client sends `X-Request-ID`, the backend reuses it; otherwise the backend generates one.

Use `request_id` to correlate:

- browser/API client error
- API container logs
- reverse proxy logs
- incident notes

## API logs

API logs are written to container stdout/stderr and can be read with:

```bash
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml logs --tail=200 api
```

HTTP request logs include:

- `request_id`
- `method`
- `path`
- `route` / `path_template`
- `status`
- `duration_ms`
- `ip`
- `user_agent`

Slow requests are logged as warnings. Do not enable SQL value logging in production.

## Worker logs

Worker logs are written to container stdout/stderr:

```bash
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml logs --tail=200 worker
```

Outbox event logs include:

- `event_id`
- `event_type`
- `attempt`
- `attempts`
- `status`
- `tenant_id`
- `aggregate_type`
- `aggregate_id`

Worker logs must not print full event payloads because payloads may contain customer or business data.

## Caddy logs

Caddy logs can be read with:

```bash
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml logs --tail=200 caddy
```

Use Caddy logs for TLS, reverse proxy, and upstream availability issues.

## Nginx logs

Nginx is optional in this repo. If used, check the configured access/error log paths or container logs:

```bash
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml logs --tail=200 nginx
```

## Sensitive fields that must not be logged

Never log:

- passwords or password hashes
- access tokens
- refresh tokens
- `Authorization` headers
- raw request bodies
- customer full addresses
- payment proof contents or private URLs
- financial notes
- `DATABASE_URL`
- `JWT_SECRET`
- database passwords
- raw outbox payloads

Safe identifiers to log when needed:

- `request_id`
- `user_id`
- `tenant_id`
- `store_id`
- `order_id`
- `product_id`
- `event_id`
- standardized error code

