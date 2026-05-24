#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
REPO_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

ENV_FILE="${ENV_FILE:-deploy/.env.production}"

if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
fi

HEALTH_URL="${API_HEALTH_URL:-}"
if [ -z "$HEALTH_URL" ] && [ "${API_DOMAIN:-}" ]; then
  HEALTH_URL="https://$API_DOMAIN/health/ready"
fi
if [ -z "$HEALTH_URL" ]; then
  HEALTH_URL="http://localhost:8080/health/ready"
fi

echo "Checking API health: $HEALTH_URL"
curl --fail --silent --show-error --max-time 15 "$HEALTH_URL" >/dev/null
echo "API health check passed"
