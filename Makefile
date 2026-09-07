.PHONY: db-up db-down migrate-up migrate-down api web test-api build-api build-web

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

migrate-up:
	docker compose exec -T postgres psql -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000001_core.up.sql

migrate-down:
	docker compose exec -T postgres psql -U $${POSTGRES_USER:-property_os} -d $${POSTGRES_DB:-property_os} < migrations/000001_core.down.sql

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
