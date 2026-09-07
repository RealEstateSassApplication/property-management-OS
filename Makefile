.PHONY: db-up db-down migrate-up migrate-people migrate-finance migrate-maintenance migrate-identity migrate-documents migrate-notifications migrate-agent-actions migrate-portals migrate-org-settings migrate-all migrate-down migrate-org-settings-down migrate-portals-down migrate-agent-actions-down migrate-notifications-down migrate-documents-down migrate-identity-down migrate-maintenance-down migrate-finance-down migrate-people-down seed api worker mcp web test-api build-api build-web

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

migrate-up:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000001_core.up.sql
migrate-people:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000002_people_leasing.up.sql
migrate-finance:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000003_owners_rent.up.sql
migrate-maintenance:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000004_maintenance_operations.up.sql
migrate-identity:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000005_user_identities.up.sql
migrate-documents:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000006_documents.up.sql
migrate-notifications:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000007_notification_outbox.up.sql
migrate-agent-actions:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000008_agent_actions.up.sql
migrate-portals:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000009_portal_access.up.sql
migrate-org-settings:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000010_organization_settings.up.sql
migrate-all: migrate-up migrate-people migrate-finance migrate-maintenance migrate-identity migrate-documents migrate-notifications migrate-agent-actions migrate-portals migrate-org-settings

migrate-org-settings-down:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000010_organization_settings.down.sql
migrate-portals-down:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000009_portal_access.down.sql
migrate-agent-actions-down:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000008_agent_actions.down.sql
migrate-notifications-down:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000007_notification_outbox.down.sql
migrate-documents-down:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000006_documents.down.sql
migrate-identity-down:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000005_user_identities.down.sql
migrate-maintenance-down:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000004_maintenance_operations.down.sql
migrate-finance-down:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000003_owners_rent.down.sql
migrate-people-down:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000002_people_leasing.down.sql
migrate-down: migrate-org-settings-down migrate-portals-down migrate-agent-actions-down migrate-notifications-down migrate-documents-down migrate-identity-down migrate-maintenance-down migrate-finance-down migrate-people-down
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000001_core.down.sql

seed:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < scripts/dev-seed.sql
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < scripts/dev-document-seed.sql
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < scripts/dev-notification-seed.sql
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < scripts/dev-agent-seed.sql
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < scripts/dev-portal-seed.sql

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
