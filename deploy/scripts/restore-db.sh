#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
REPO_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

COMPOSE_FILE="${COMPOSE_FILE:-deploy/docker-compose.prod.yml}"
ENV_FILE="${ENV_FILE:-deploy/.env.production}"
YES="false"

case "${1:-}" in
  --yes|-y)
    YES="true"
    shift
    ;;
esac

BACKUP_FILE="${1:-}"

if [ -z "$BACKUP_FILE" ]; then
  echo "Usage: $0 [--yes] path/to/backup.sql.gz" >&2
  exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
  echo "Missing $ENV_FILE. Copy deploy/.env.production.example first." >&2
  exit 1
fi

if [ ! -f "$BACKUP_FILE" ]; then
  echo "Backup file not found: $BACKUP_FILE" >&2
  exit 1
fi

gzip -t "$BACKUP_FILE"

cat >&2 <<'EOF'
WARNING: database restore is destructive.
- Do not run this blindly on production.
- Test the backup on staging first whenever possible.
- API and worker will be stopped to prevent writes during restore.
EOF

if [ "$YES" != "true" ]; then
  printf "Type RESTORE to continue: " >&2
  read -r CONFIRM
  if [ "$CONFIRM" != "RESTORE" ]; then
    echo "Restore cancelled." >&2
    exit 1
  fi
fi

echo "Stopping API and worker before restore to prevent writes..."
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" stop api worker

gzip -dc "$BACKUP_FILE" | docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" exec -T postgres \
  sh -c 'psql -U "$POSTGRES_USER" "$POSTGRES_DB"'

docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" start api worker
echo "Restore completed from: $BACKUP_FILE"
