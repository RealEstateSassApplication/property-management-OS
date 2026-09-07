.PHONY: db-up db-down migrate-up migrate-people migrate-finance migrate-maintenance migrate-identity migrate-all migrate-down migrate-identity-down migrate-maintenance-down migrate-finance-down migrate-people-down seed api web test-api build-api build-web

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

migrate-all: migrate-up migrate-people migrate-finance migrate-maintenance migrate-identity

migrate-identity-down:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000005_user_identities.down.sql

migrate-maintenance-down:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000004_maintenance_operations.down.sql

migrate-finance-down:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000003_owners_rent.down.sql

migrate-people-down:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000002_people_leasing.down.sql

migrate-down: migrate-identity-down migrate-maintenance-down migrate-finance-down migrate-people-down
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000001_core.down.sql

seed:
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < scripts/dev-seed.sql

api:
	cd apps/api && go run ./cmd/api

web:
	npm run dev:web

test-api:
	cd apps/api && go test ./...

build-api:
	cd apps/api && go build ./cmd/api

build-web:
	npm run build:web
