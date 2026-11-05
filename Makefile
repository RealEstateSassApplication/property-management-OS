.PHONY: dev-setup dev dev-stop db-up db-down migrate-all migrate-down seed api worker mcp web test-api build-api build-web

dev-setup:
	sh scripts/dev-setup.sh

dev:
	sh scripts/dev-run.sh

dev-stop:
	docker compose down

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

# The versioned runner is the ONLY supported way to apply migrations.
# It tracks checksums, serializes deploys and refuses unknown legacy schemas.
migrate-all:
	@set -a; if [ -f .env ]; then . ./.env; fi; set +a; cd apps/api && go run ./cmd/migrate -dir ../../migrations

# Legacy partial/down targets would bypass the migration ledger. Fail closed.
legacy-migration-targets := migrate-up migrate-people migrate-finance migrate-maintenance migrate-identity migrate-documents migrate-notifications migrate-agent-actions migrate-portals migrate-org-settings migrate-accounting migrate-inspections migrate-payment-providers migrate-push-devices migrate-payment-providers-down migrate-inspections-down migrate-accounting-down migrate-org-settings-down migrate-portals-down migrate-agent-actions-down migrate-notifications-down migrate-documents-down migrate-identity-down migrate-maintenance-down migrate-finance-down migrate-people-down migrate-push-devices-down migrate-down
.PHONY: $(legacy-migration-targets)
$(legacy-migration-targets):
	@echo "Unsupported direct migration target: use 'make migrate-all' or the reviewed restore procedure in docs/migrations.md" >&2
	@exit 1

seed:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < scripts/dev-seed.sql
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < scripts/dev-document-seed.sql
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < scripts/dev-notification-seed.sql
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < scripts/dev-agent-seed.sql
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < scripts/dev-portal-seed.sql
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < scripts/dev-accounting-seed.sql
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < scripts/dev-inspection-seed.sql
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < scripts/dev-payment-provider-seed.sql

api:
	cd apps/api && go run ./cmd/api
worker:
	cd apps/api && go run ./cmd/worker
mcp:
	cd apps/api && go run ./cmd/mcp
web:
	npm run dev:web
test-api:
	cd apps/api && go test ./...
build-api:
	cd apps/api && go build ./cmd/api ./cmd/worker ./cmd/mcp
build-web:
	npm run build:web
