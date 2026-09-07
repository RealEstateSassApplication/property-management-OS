#!/usr/bin/env sh
set -eu

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required" >&2
  exit 1
fi

BACKUP_DIR="${BACKUP_DIR:-./backups}"
TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
BACKUP_FILE="${BACKUP_FILE:-${BACKUP_DIR}/property-os-${TIMESTAMP}.dump}"

mkdir -p "$(dirname "$BACKUP_FILE")"

echo "Creating PostgreSQL backup at $BACKUP_FILE"
pg_dump \
  --dbname="$DATABASE_URL" \
  --format=custom \
  --no-owner \
  --no-privileges \
  --file="$BACKUP_FILE"

pg_restore --list "$BACKUP_FILE" >/dev/null

echo "Backup created and archive structure verified: $BACKUP_FILE"
