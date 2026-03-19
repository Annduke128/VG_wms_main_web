.PHONY: dev dev-api dev-worker build migrate docker-up docker-down \
	lint test seed clean web-install web-dev web-build docker-prod

# ============================================
# Infrastructure
# ============================================

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-prod:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build

# ============================================
# Database
# ============================================

MIGRATE_DSN ?= postgres://wms:wms_secret@localhost:5432/wms?sslmode=disable

migrate:
	migrate -path migrations -database "$(MIGRATE_DSN)" up

migrate-down:
	migrate -path migrations -database "$(MIGRATE_DSN)" down

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

seed:
	psql "$(MIGRATE_DSN)" -f scripts/seed.sql

# ============================================
# Development
# ============================================

dev-api:
	go run ./cmd/api

dev-worker:
	go run ./cmd/worker

dev-air:
	air

# ============================================
# Quality
# ============================================

lint:
	golangci-lint run ./...

test:
	go test -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

# ============================================
# Build
# ============================================

build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/api ./cmd/api
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/worker ./cmd/worker

clean:
	rm -rf bin/ tmp/ coverage.out coverage.html

# ============================================
# Frontend
# ============================================

web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

web-lint:
	cd web && npm run lint

# ============================================
# All-in-one
# ============================================

setup: docker-up migrate seed web-install
	@echo "Setup complete. Run 'make dev-api' and 'make web-dev' to start."
