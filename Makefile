.PHONY: dev dev-api dev-worker build migrate docker-up docker-down

# Start infrastructure
docker-up:
	docker compose up -d

docker-down:
	docker compose down

# Run migrations (requires golang-migrate)
migrate:
	migrate -path migrations -database "postgres://wms:wms_secret@localhost:5432/wms?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://wms:wms_secret@localhost:5432/wms?sslmode=disable" down

# Development
dev-api:
	go run ./cmd/api

dev-worker:
	go run ./cmd/worker

# Build
build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

# Frontend
web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build
