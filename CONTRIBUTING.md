# Contributing — WMS Web v1

Hướng dẫn phát triển và đóng góp cho dự án.

## Yêu cầu

| Tool             | Version | Cài đặt                                                                               |
| ---------------- | ------- | ------------------------------------------------------------------------------------- |
| Go               | >= 1.23 | https://go.dev/dl/                                                                    |
| Node.js          | >= 18   | https://nodejs.org/                                                                   |
| Docker + Compose | latest  | https://docs.docker.com/get-docker/                                                   |
| golang-migrate   | latest  | `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |
| Air (optional)   | latest  | `go install github.com/air-verse/air@latest`                                          |
| golangci-lint    | >= 1.60 | https://golangci-lint.run/usage/install/                                              |

## Setup lần đầu

```bash
# 1. Clone repo
git clone <repo-url>
cd VG_wms_main_web

# 2. Cấu hình
cp .env.example .env

# 3. Khởi động DB + Redis
make docker-up

# 4. Chạy migration
make migrate

# 5. (Optional) Seed data mẫu
make seed

# 6. Chạy backend
make dev-api       # Terminal 1
make dev-worker    # Terminal 2

# 7. Chạy frontend
make web-install
make web-dev       # Terminal 3
```

## Development Workflow

### Backend (Go)

```bash
# Hot reload với Air
air

# Hoặc chạy trực tiếp
make dev-api
make dev-worker

# Lint
golangci-lint run ./...

# Test
go test ./...

# Build
make build
```

### Frontend (React)

```bash
cd web

# Dev server (http://localhost:5173)
npm run dev

# Lint
npm run lint

# Type check
npx tsc --noEmit

# Build production
npm run build
```

### Database

```bash
# Tạo migration mới
migrate create -ext sql -dir migrations -seq <name>

# Chạy migration
make migrate

# Rollback
make migrate-down

# Seed data mẫu
make seed
```

## Cấu trúc code

### Backend

```
cmd/api/         → Fiber HTTP server entry point
cmd/worker/      → Background job worker entry point
internal/
  domain/        → Entities, types, business rules (KHÔNG import package khác)
  grid/          → SQL query builder cho grid API
  repo/          → PostgreSQL data access layer
  service/       → Business logic, orchestration
  queue/         → Redis job queue
  importer/      → Excel (.xlsx) file parser
  web/           → HTTP handlers + routes
```

**Quy tắc dependency:**

```
web → service → repo → domain
                     → queue
         importer ↗
```

`domain/` không import bất kỳ package nào trong `internal/`.

### Frontend

```
web/src/
  api/           → API client functions
  components/    → React components (Grid, Kanban, Import)
  types/         → TypeScript type definitions
  App.tsx        → Root component, tab navigation
```

## API Testing

### Với curl

```bash
# Grid query
curl -X POST http://localhost:8080/api/inventory/grid \
  -H 'Content-Type: application/json' \
  -d '{"startRow":0,"endRow":50,"sortModel":[],"filterModel":{}}'

# Update item
curl -X PATCH http://localhost:8080/api/inventory/SP001 \
  -H 'Content-Type: application/json' \
  -d '{"so_ton":200}'

# List kanban inbound
curl http://localhost:8080/api/kanban/inbound

# Import products
curl -X POST http://localhost:8080/api/import/products \
  -F 'file=@products.xlsx'
```

## Quy ước

### Git

- Branch: `feature/<name>`, `fix/<name>`, `chore/<name>`
- Commit message: `type(scope): description`
  - `feat`, `fix`, `refactor`, `test`, `chore`, `docs`
- Luôn chạy lint trước khi commit
- Không commit file `.env`

### Code Style

- **Go:** Theo `gofmt` + `golangci-lint` config (`.golangci.yml`)
- **TypeScript:** Theo ESLint config (`eslint.config.js`)
- **SQL:** Lowercase keywords, snake_case cho tên bảng/cột
- **Indentation:** Tab cho Go, 2 spaces cho TS/JSON/YAML (xem `.editorconfig`)

### Naming

| Ngôn ngữ  | Convention | Ví dụ                  |
| --------- | ---------- | ---------------------- |
| Go        | camelCase  | `inventoryService`     |
| Go export | PascalCase | `InventoryService`     |
| TS        | camelCase  | `fetchInventoryGrid()` |
| TS types  | PascalCase | `GridRequest`          |
| SQL       | snake_case | `inventory_main`       |
| API       | kebab-case | `/api/bulk-update`     |
