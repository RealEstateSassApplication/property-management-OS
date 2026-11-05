#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT_DIR"

require() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required local dependency: $1" >&2
    exit 1
  fi
}

require docker
require go
require npm

if [ ! -f .env ]; then
  cp .env.example .env
  echo "Created .env from .env.example"
fi

set -a
# shellcheck disable=SC1091
. ./.env
set +a

echo "Starting PostgreSQL..."
docker compose up -d postgres

printf "Waiting for PostgreSQL"
i=0
until docker compose exec -T postgres pg_isready -U "${POSTGRES_USER:-property_os}" -d "${POSTGRES_DB:-property_os}" >/dev/null 2>&1; do
  i=$((i + 1))
  if [ "$i" -ge 30 ]; then
    echo "\nPostgreSQL did not become ready." >&2
    exit 1
  fi
  printf "."
  sleep 1
done
printf " ready\n"

schema_exists=$(docker compose exec -T postgres psql -U "${POSTGRES_USER:-property_os}" -d "${POSTGRES_DB:-property_os}" -Atc "SELECT to_regclass('public.organizations') IS NOT NULL")

echo "Verifying migration history and applying outstanding upgrades..."
make migrate-all

if [ "$schema_exists" != "t" ]; then
  echo "Loading deterministic development data..."
  make seed
else
  echo "Preserving existing development records (no automatic reseed)."
fi

echo "Downloading Go modules..."
(cd apps/api && go mod download)

echo "Installing web dependencies..."
npm install

echo ""
echo "Local setup complete."
echo "Run: make dev"
echo "Web: http://localhost:3000"
echo "API health: http://localhost:8080/healthz"
echo "API readiness: http://localhost:8080/readyz"
