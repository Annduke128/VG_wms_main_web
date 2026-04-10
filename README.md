# WMS Web v2 — Warehouse Management System

Hệ thống quản lý kho hàng (Warehouse Management System) dành cho đơn vị bán lẻ. Hỗ trợ **multi-warehouse** (nhiều kho), quản lý tồn kho, nhập/xuất hàng với FIFO theo batch (mã thùng), combo/phụ kiện, dashboard KPI + biểu đồ, cài đặt ngưỡng cảnh báo, và import dữ liệu từ Excel.

## Kiến trúc tổng quan

```
┌────────────────────────────────────────────────────────────────────┐
│  Frontend (React 18 + Vite + TypeScript)                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │ Overview │  │Inventory │  │ Orders   │  │  Combo   │  │ Settings │ │
│  │ KPI+Chart│  │GlideGrid │  │ FIFO+    │  │Warehouse │  │Threshold │ │
│  │ Alerts   │  │ +Lots    │  │ Create   │  │+Accessory│  │ Manual   │ │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘ │
│       │              │             │              │             │       │
│  ┌────────────────────────────────────────────────────────────────┐     │
│  │        WarehouseSelector (Zustand store + persist)            │     │
│  └───────────────────────────┬────────────────────────────────────┘     │
└──────────────────────────────┼─────────────────────────────────────────┘
        │              │             │             │
        ▼              ▼             ▼             ▼
┌────────────────────────────────────────────────────────────────────┐
│  Backend (Go + Gin)                                                │
│  ┌─────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │Handlers │→ │ Services │→ │   Repo   │→ │PostgreSQL│          │
│  └─────────┘  └────┬─────┘  └──────────┘  └──────────┘          │
│                     │                                              │
│                     ├──→ Redis Queue ──→ Worker (import/bulk)     │
│                     └──→ ClickHouse (analytics, future)           │
└────────────────────────────────────────────────────────────────────┘
```

## Tech Stack

| Layer     | Công nghệ                                                      |
| --------- | -------------------------------------------------------------- |
| Frontend  | React 18, TypeScript, Vite, Glide Data Grid, Recharts, Zustand |
| Backend   | Go, Gin, pgx v5                                                |
| Database  | PostgreSQL 16 (transactional)                                  |
| Analytics | ClickHouse 24 (future)                                         |
| Queue     | Redis 7 (async jobs)                                           |
| Excel     | excelize v2 (`.xlsx`, date `MM/DD/YYYY`)                       |

## Cấu trúc dự án

```
VG_wms_main_web/
├── cmd/
│   ├── api/main.go            # HTTP server (Gin)
│   └── worker/main.go         # Background worker (import, bulk-update)
├── internal/
│   ├── domain/                # Entities, grid, orders, dashboard, combo types
│   ├── grid/                  # SQL query builder (filter, sort, paginate)
│   ├── importer/              # Excel parser (4 loại file)
│   ├── queue/                 # Redis job queue
│   ├── repo/                  # PostgreSQL repositories (+ warehouse CRUD)
│   ├── service/               # Business logic layer
│   └── web/                   # HTTP handlers + routes + static serving
├── migrations/
│   ├── 001_init.up.sql              # Schema creation (core tables)
│   ├── 002_fifo_thresholds.up.sql   # FIFO lots + thresholds
│   ├── 003_pricing_metrics.up.sql   # Pricing + metrics columns
│   ├── 004_inventory_grid_view.up.sql # inventory_grid VIEW (join products)
│   ├── 005_combo_system.up.sql       # Combo/phụ kiện tables
│   ├── 006_drop_kanban_tables.up.sql  # Drop legacy kanban tables
│   ├── 007_multi_warehouse.up.sql     # Multi-warehouse: warehouses table + warehouse_id FK
│   └── *.down.sql                   # Rollback migrations
├── web/                       # Frontend React app
│   ├── src/
│   │   ├── api/client.ts      # API client (typed fetch)
│   │   ├── components/        # Sidebar, WarehouseSelector, KpiCards, Charts, InventoryGrid, etc.
│   │   ├── stores/            # Zustand stores (warehouseStore — persist active warehouse)
│   │   ├── views/             # Overview, Inventory, Orders, ComboWarehouse, Settings
│   │   └── types/             # TypeScript types (warehouse, grid, combo, dashboard, inventory)
│   └── package.json
├── docker-compose.yml         # Dev: PostgreSQL, Redis, ClickHouse (migrate via profile)
├── docker-compose.prod.yml    # Prod: + API, Worker
├── Dockerfile                 # Multi-stage: FE build + Go build
├── Dockerfile.worker          # Worker binary
├── .env.example
├── Makefile
└── go.mod
```

## Tính năng chính

### Multi-Warehouse (Nhiều kho)

- **WarehouseSelector** trong sidebar — chọn kho hoạt động, persist vào localStorage
- Tất cả data (inventory, orders, combo, dashboard) scoped theo `warehouse_id`
- AG Grid tự động reset/purge khi chuyển kho (React key pattern)
- Warehouse CRUD API: tạo, sửa, liệt kê kho
- `InitWarehouseInventory` — tự khởi tạo inventory cho warehouse mới từ master products

### Dashboard (Tổng quan)

- **4 KPI cards:** Tổng SKU, SKU tồn lâu, SKU thiếu hàng, Tổng tiền hàng
- **2 biểu đồ:** Nhập/Xuất theo tuần (4 tuần) + Tồn thực vs Tối ưu (Top SKU)
- **Cảnh báo:** Tồn lâu + Thiếu hàng (dựa trên thresholds)

### Kho hàng (Inventory)

- **Glide Data Grid** — virtualized, 20k+ dòng, 60 FPS, inline editing
- **Lots panel** — click vào dòng để xem chi tiết lô hàng (batch_code, qty, ngày nhập)
- Server-side pagination, sort, filter

### Nhập / Xuất (Orders)

- **Tạo đơn nhập** (inbound) với mã thùng (batch_code)
- **Tạo đơn xuất** (outbound) — tự động FIFO, phân bổ lô cũ nhất trước
- **batch_code in đậm** trong bảng xuất
- Danh sách đơn phân trang, lọc theo loại

### Settings (Cài đặt ngưỡng)

- Nhập thresholds thủ công: `min_qty`, `optimal_qty`, `max_age_days`
- Lưu với lịch sử (versioned, `effective_from/to`)
- Tra cứu lịch sử theo mã hàng

### Combo / Phụ kiện (ComboWarehouse)

- **Catalog tab** — quản lý BOM (bill of materials) cho combo + semi-finished
- **Action tab** — tạo combo (ghép), xuất combo, trả combo
- **Inventory tab** — tồn kho combo theo warehouse
- **Accessories tab** — quản lý phụ kiện kèm theo combo

### FIFO theo batch (mã thùng)

- Mỗi lần nhập hàng tạo/cập nhật `inventory_lots`
- Xuất hàng allocate FIFO (oldest `received_at` first)
- Có thể split thành nhiều dòng outbound nếu 1 lot không đủ

### Import Excel

- 4 loại: Products, Inventory, Inbound, Outbound
- Async processing qua Redis queue + Worker
- Poll trạng thái job

## Yêu cầu hệ thống

| Tool           | Version | Mục đích                                  |
| -------------- | ------- | ----------------------------------------- |
| Docker         | >= 24   | Tất cả services                           |
| Docker Compose | >= 2.20 | Orchestration                             |
| Go             | >= 1.23 | Backend (dev mode)                        |
| Node.js        | >= 18   | Frontend (dev mode)                       |
| golang-migrate | >= 4.17 | DB migration (dev mode, hoặc dùng docker) |

## Cài đặt & Chạy

### Quick Start (Docker)

```bash
# 1. Clone & cấu hình
git clone <repo-url>
cd VG_wms_main_web
cp .env.example .env

# 2. Khởi động infra (KHÔNG auto migrate)
docker compose up -d
# → PostgreSQL :5432, Redis :6379, ClickHouse :8123

# 3. Chạy migration (chọn 1 trong 2 cách)
make migrate                                              # Cách 1: CLI trên host (cần cài golang-migrate)
docker compose --profile migrate run --rm migrate         # Cách 2: Docker (không cần cài gì)

# 4. Chạy backend (dev mode)
make dev-api       # Terminal 1 — API :8080
make dev-worker    # Terminal 2 — Worker

# 5. Chạy frontend (dev mode)
make web-install   # Lần đầu
make web-dev       # → http://localhost:5173 (proxy /api → :8080)
```

### Production (Docker)

```bash
# Set env
export POSTGRES_PASSWORD=your_secure_password

# Chạy migration trước
docker compose --profile migrate run --rm migrate

# Build & run all services
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
# → API + FE served on :8080
```

## Cơ sở dữ liệu

### PostgreSQL — Bảng chính

| Bảng                   | Mục đích                                 |
| ---------------------- | ---------------------------------------- |
| `warehouses`           | Danh sách kho (multi-warehouse)          |
| `products`             | Danh mục sản phẩm (master data)          |
| `inventory_main`       | Tồn kho hiện tại (số tồn, nhập, xuất)    |
| `inbound_items`        | Chi tiết phiếu nhập (+ batch_code)       |
| `outbound_items`       | Chi tiết phiếu xuất (+ batch_code)       |
| `inventory_lots`       | Lô hàng FIFO (batch_code, qty_remaining) |
| `inventory_thresholds` | Ngưỡng cảnh báo (versioned + history)    |
| `inventory_movements`  | Log biến động kho (IN/OUT)               |
| `inventory_metrics`    | Metrics tự động tính (LBBQ, tồn bán...)  |
| `combo_masters`        | Định nghĩa combo (BOM)                   |
| `combo_transactions`   | Lịch sử ghép/xuất/trả combo              |
| `accessories`          | Phụ kiện combo                           |
| `rule_config`          | Cấu hình quy tắc                         |
| `import_batches`       | Lịch sử import Excel                     |
| `async_jobs`           | Trạng thái job bất đồng bộ               |

> **Multi-warehouse:** Tất cả business tables (trừ `warehouses`) có `warehouse_id` FK. Mọi query scoped theo warehouse.

## API Endpoints

### Dashboard

| Method | Path                     | Mô tả                                       |
| ------ | ------------------------ | ------------------------------------------- |
| GET    | `/api/dashboard/summary` | 4 KPI: SKU count, tồn lâu, thiếu, tổng tiền |
| GET    | `/api/dashboard/charts`  | 2 charts: nhập/xuất tuần + tồn vs optimal   |

### Inventory

| Method | Path                            | Mô tả                               |
| ------ | ------------------------------- | ----------------------------------- |
| POST   | `/api/inventory/grid`           | Truy vấn lưới (filter, sort, page)  |
| PATCH  | `/api/inventory/:ma_hang`       | Cập nhật 1 dòng inventory           |
| POST   | `/api/inventory/bulk-update`    | Cập nhật hàng loạt → 202 + job_id   |
| GET    | `/api/inventory/lots`           | Lots theo ma_hang (FIFO)            |
| GET    | `/api/inventory/alerts`         | Cảnh báo tồn lâu + thiếu hàng       |
| GET    | `/api/inventory/filter-options` | Danh sách BU + Nhóm hàng (dropdown) |
| POST   | `/api/inventory/export`         | Export Excel (chọn dòng + cột)      |
| POST   | `/api/inventory/recalc-all`     | Recalc tất cả metrics (async)       |

### Orders

| Method | Path          | Mô tả                              |
| ------ | ------------- | ---------------------------------- |
| GET    | `/api/orders` | Danh sách đơn (nhập + xuất UNION)  |
| POST   | `/api/orders` | Tạo đơn nhập/xuất (auto FIFO xuất) |

### Thresholds

| Method | Path              | Mô tả                          |
| ------ | ----------------- | ------------------------------ |
| GET    | `/api/thresholds` | Lịch sử threshold theo ma_hang |
| POST   | `/api/thresholds` | Lưu threshold mới (close cũ)   |

### Warehouses

| Method | Path                  | Mô tả                                |
| ------ | --------------------- | ------------------------------------ |
| GET    | `/api/warehouses`     | Danh sách tất cả warehouses          |
| POST   | `/api/warehouses`     | Tạo warehouse mới (+ init inventory) |
| GET    | `/api/warehouses/:id` | Chi tiết warehouse                   |
| PATCH  | `/api/warehouses/:id` | Cập nhật warehouse (tên, địa chỉ)    |

### Kanban & Import

| Method | Path                            | Mô tả                             |
| ------ | ------------------------------- | --------------------------------- |
| GET    | `/api/kanban/inbound`           | Danh sách thẻ nhập                |
| POST   | `/api/kanban/inbound`           | Tạo thẻ nhập mới                  |
| POST   | `/api/kanban/inbound/:id/move`  | Chuyển trạng thái nhập            |
| GET    | `/api/kanban/outbound`          | Danh sách thẻ xuất                |
| POST   | `/api/kanban/outbound`          | Tạo thẻ xuất mới                  |
| POST   | `/api/kanban/outbound/:id/move` | Chuyển trạng thái xuất            |
| POST   | `/api/import/{type}`            | Import Excel (.xlsx)              |
| GET    | `/api/jobs/:id`                 | Trạng thái job                    |
| POST   | `/api/admin/reset-all`          | Reset toàn bộ DB (trừ warehouses) |

## Biến môi trường

| Biến             | Mặc định                                                       | Mô tả                 |
| ---------------- | -------------------------------------------------------------- | --------------------- |
| `POSTGRES_DSN`   | `postgres://wms:wms_secret@localhost:5432/wms?sslmode=disable` | PostgreSQL connection |
| `REDIS_ADDR`     | `localhost:6379`                                               | Redis address         |
| `CLICKHOUSE_DSN` | `tcp://localhost:9000?database=wms`                            | ClickHouse (future)   |
| `PORT`           | `8080`                                                         | API server port       |
| `STATIC_DIR`     | `web/dist`                                                     | FE static files       |

## Makefile Commands

| Command               | Mô tả                                                   |
| --------------------- | ------------------------------------------------------- |
| `make docker-up`      | Khởi động PostgreSQL, Redis, ClickHouse (KHÔNG migrate) |
| `make docker-down`    | Dừng tất cả containers                                  |
| `make docker-prod`    | Build + chạy production (API+Worker+FE)                 |
| `make migrate`        | Chạy migration PostgreSQL (cần cài golang-migrate)      |
| `make migrate-status` | Kiểm tra version + dirty state                          |
| `make migrate-fix`    | Interactive fix dirty migration                         |
| `make dev-api`        | Chạy API server (dev mode)                              |
| `make dev-worker`     | Chạy background worker (dev mode)                       |
| `make build`          | Build binary production                                 |
| `make web-dev`        | Chạy frontend dev server (proxy :8080)                  |
| `make web-build`      | Build frontend production                               |

## Giới hạn hiện tại

- ClickHouse chưa tích hợp (dành cho analytics phase sau)
- Chưa có authentication/authorization (MVP)
- ML forecasting deferred

## License

Private — Internal use only.
