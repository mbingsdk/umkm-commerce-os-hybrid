#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
REPO_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

COMPOSE_FILE="${COMPOSE_FILE:-deploy/docker-compose.prod.yml}"
ENV_FILE="${ENV_FILE:-deploy/.env.production}"
BACKUP_DIR="${BACKUP_DIR:-backups/database/daily}"
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-14}"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
BACKUP_FILE="$BACKUP_DIR/umkm_os_db_$TIMESTAMP.sql.gz"
TMP_SQL="$BACKUP_DIR/.umkm_os_db_$TIMESTAMP.sql.tmp"
TMP_GZ="$BACKUP_FILE.tmp"

cleanup() {
  rm -f "$TMP_SQL" "$TMP_GZ"
}
trap cleanup EXIT INT TERM

if [ ! -f "$ENV_FILE" ]; then
  echo "Missing $ENV_FILE. Copy deploy/.env.production.example first." >&2
  exit 1
fi

mkdir -p "$BACKUP_DIR"

echo "Starting PostgreSQL backup..."
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" exec -T postgres \
  sh -c 'pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB"' > "$TMP_SQL"

if [ ! -s "$TMP_SQL" ]; then
  echo "Backup failed: pg_dump output is empty." >&2
  exit 1
fi

gzip -c "$TMP_SQL" > "$TMP_GZ"
gzip -t "$TMP_GZ"

mv "$TMP_GZ" "$BACKUP_FILE"
rm -f "$TMP_SQL"

if [ "$BACKUP_RETENTION_DAYS" -gt 0 ] 2>/dev/null; then
  find "$BACKUP_DIR" -type f -name 'umkm_os_db_*.sql.gz' -mtime +"$BACKUP_RETENTION_DAYS" -delete
fi

echo "Backup written: $BACKUP_FILE"
echo "Retention: keeping backups newer than $BACKUP_RETENTION_DAYS day(s) in $BACKUP_DIR"
