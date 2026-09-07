#!/usr/bin/env sh
set -eu

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required" >&2
  exit 1
fi

BACKUP_FILE="${1:-${BACKUP_FILE:-}}"
if [ -z "$BACKUP_FILE" ] || [ ! -f "$BACKUP_FILE" ]; then
  echo "Usage: RESTORE_CONFIRM=YES DATABASE_URL=... sh scripts/restore-postgres.sh <backup.dump>" >&2
  exit 1
fi

if [ "${RESTORE_CONFIRM:-}" != "YES" ]; then
  echo "Restore is destructive. Set RESTORE_CONFIRM=YES to continue." >&2
  exit 1
fi

pg_restore --list "$BACKUP_FILE" >/dev/null

echo "Restoring $BACKUP_FILE into configured DATABASE_URL"
pg_restore \
  --dbname="$DATABASE_URL" \
  --clean \
  --if-exists \
  --no-owner \
  --no-privileges \
  "$BACKUP_FILE"

echo "Restore completed. Run migrations and application invariants before serving traffic."
