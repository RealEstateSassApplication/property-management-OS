#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT_DIR"

if [ ! -f .env ]; then
  echo "Missing .env. Run: make dev-setup" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1091
. ./.env
set +a

cleanup() {
  status=$?
  trap - INT TERM EXIT
  [ -n "${API_PID:-}" ] && kill "$API_PID" 2>/dev/null || true
  [ -n "${WORKER_PID:-}" ] && kill "$WORKER_PID" 2>/dev/null || true
  [ -n "${WEB_PID:-}" ] && kill "$WEB_PID" 2>/dev/null || true
  wait 2>/dev/null || true
  exit "$status"
}
trap cleanup INT TERM EXIT

echo "Starting Property OS local development stack..."
echo "  web    http://localhost:3000"
echo "  api    http://localhost:8080"
echo "  health http://localhost:8080/healthz"
echo ""

(
  cd apps/api
  go run ./cmd/api
) &
API_PID=$!

(
  cd apps/api
  go run ./cmd/worker
) &
WORKER_PID=$!

npm run dev:web &
WEB_PID=$!

wait "$API_PID" "$WORKER_PID" "$WEB_PID"
