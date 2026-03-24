# AGENTS.md — WMS (Warehouse Management System)

Hướng dẫn cho AI agents làm việc với dự án này.

## Tổng quan

Hệ thống quản lý kho (WMS) cho doanh nghiệp Việt Nam. Go/Gin backend, React frontend, PostgreSQL, Redis queue.

## Kiến trúc

```
cmd/
├── api/main.go          # HTTP server (Gin) — port 8080
└── worker/main.go       # Background worker (Redis queue consumer)

internal/
├── domain/              # Entities, types (entities.go, grid.go, kanban.go, orders.go, dashboard.go)
├── repo/                # PostgreSQL queries (inventory, import, kanban, orders, dashboard, reset, inventory_metrics)
├── service/             # Business logic (inventory, importer, kanban, orders, dashboard)
├── web/                 # HTTP handlers + routes (Gin)
├── grid/                # AG Grid server-side builder
├── importer/            # Excel import (xlsx.go, template.go, inventory_full.go)
└── queue/               # Redis queue (redis.go)

web/                     # React frontend (Vite + TypeScript)
├── src/
│   ├── api/client.ts    # API client (axios)
│   ├── views/           # Pages: Overview, Inventory, Orders, Settings
│   ├── components/      # UI components: KanbanBoard, InventoryGrid, ImportPanel, Charts, etc.
│   └── types/           # TypeScript types matching Go domain
├── vite.config.ts
└── package.json

migrations/              # PostgreSQL migrations (golang-migrate format)
```

## Build & Deploy

### Build binaries

```bash
make build               # → bin/api, bin/worker (production, stripped)
```

Khi chạy `go build ./cmd/api` hoặc `go build ./cmd/worker` trực tiếp (không có `-o`), binary sẽ xuất ra root folder (`/api`, `/worker`). Các file này đã được thêm vào `.gitignore` — **không commit binary**.

Để deploy lên cloud, luôn dùng `make build` để output vào `bin/`.

### Development

```bash
make docker-up           # Start PostgreSQL + Redis (KHÔNG chạy migrate)
make migrate             # Run migrations (chạy riêng, chủ động)
make migrate-status      # Kiểm tra version + dirty state
make migrate-fix         # Interactive fix dirty migration
make setup               # docker-up + migrate + seed + web-install
make dev-api             # Run API server (air hot-reload nếu có)
make dev-worker          # Run worker
make web-dev             # Run frontend dev server
```

### Test & Lint

```bash
make test                # Go tests with coverage
make lint                # golangci-lint
make web-lint            # Frontend lint
make web-build           # Frontend production build
```

### Verify trước khi commit

```bash
go build ./...           # Backend compiles
cd web && npx tsc --noEmit && npm run build   # Frontend type-check + build
```

## Database

PostgreSQL với 13 business tables. Xem `migrations/` cho schema chi tiết.

**Tables chính:** products, inbound_orders, inbound_items, outbound_orders, outbound_items, inventory, inventory_metrics, kanban_inbound, kanban_outbound, pricing, customers, suppliers, product_categories.

**Reset:** `POST /api/admin/reset-all` (body: `{"confirm_text": "RESET ALL"}`) — TRUNCATE tất cả tables.

## Queue (Redis)

- `wms:queue:import` — Import Excel jobs
- `wms:queue:bulk_update` — Bulk inventory updates
- `wms:queue:recalc` — Recalc metrics cho tất cả SKUs

Worker consume từ cả 3 queues, tự động recalc metrics sau mỗi job.

## Import Excel

- Template: `MauTonKho.xlsx` (17 columns)
- Auto batch_code: `LOT-<maVach>-<yyyymmdd>` nếu không có
- Date formats hỗ trợ: dd/mm/yyyy, dd-mm-yyyy, dd-mm-yy, dd/mm/yy, d/m/yyyy, d-m-yyyy
- LBBQ=0 → NULL trong DB; UI hiện 0 với tooltip "CẦN ĐẨY HÀNG"

## Metrics tự động

`RecalcMetricsForSKU` chạy sau:

- Import (inbound, outbound, products)
- Grid edit (UpdateItem)
- Kanban moves (MoveInbound stage=da_ve_hang, MoveOutbound stage=da_giao)
- Recalc-all job (POST /api/inventory/recalc-all)

## Conventions

- Ngôn ngữ UI: Tiếng Việt
- Font: Be Vietnam Pro
- Tên cột DB/API: tiếng Việt snake_case (ma_hang, ten_san_pham, so_luong_ton, ...)
- so_ngay_ton_ban = so_ton / LBBQ (NULL nếu LBBQ NULL)
- Luôn hỏi trước khi commit hoặc thay đổi behavior

## API Endpoints chính

| Method | Path                          | Mô tả                                |
| ------ | ----------------------------- | ------------------------------------ |
| GET    | /api/inventory                | AG Grid SSR (filter, sort, paginate) |
| POST   | /api/inventory/import         | Upload Excel                         |
| POST   | /api/inventory/recalc-all     | Recalc tất cả metrics (async)        |
| GET    | /api/inventory/template       | Download MauTonKho.xlsx              |
| PUT    | /api/inventory/:id            | Update 1 item                        |
| POST   | /api/inventory/bulk-update    | Bulk update (queue)                  |
| GET    | /api/kanban/inbound           | Kanban inbound board                 |
| POST   | /api/kanban/inbound/:id/move  | Move inbound stage                   |
| GET    | /api/kanban/outbound          | Kanban outbound board                |
| POST   | /api/kanban/outbound/:id/move | Move outbound stage                  |
| GET    | /api/orders                   | Orders list                          |
| GET    | /api/dashboard                | Dashboard metrics                    |
| POST   | /api/admin/reset-all          | Reset toàn bộ DB                     |

## Files quan trọng

| File                                | Vai trò                                |
| ----------------------------------- | -------------------------------------- |
| cmd/api/main.go                     | API entry point, wiring dependencies   |
| cmd/worker/main.go                  | Worker entry point, queue consumers    |
| internal/web/routes.go              | Tất cả route definitions               |
| internal/web/handlers.go            | Tất cả HTTP handlers                   |
| internal/repo/inventory_metrics.go  | RecalcMetricsForSKU, GetAllSKUs        |
| internal/importer/inventory_full.go | Excel parsing logic                    |
| web/src/api/client.ts               | Frontend API client                    |
| .gitignore                          | Bao gồm /api, /worker (build binaries) |
