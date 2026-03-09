# WMS Web v1 — Warehouse Management System

Hệ thống quản lý kho hàng (Warehouse Management System) dành cho đơn vị bán lẻ, hỗ trợ quản lý tồn kho, nhập/xuất hàng theo luồng Kanban, và import dữ liệu từ Excel.

## Kiến trúc tổng quan

```
┌────────────────────────────────────────────────────────────────────┐
│  Frontend (React 18 + Vite + TypeScript)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────┐               │
│  │ InventoryGrid│  │ KanbanBoard  │  │ImportPanel│               │
│  │ (Glide Grid) │  │ (Inbound/Out)│  │ (.xlsx)   │               │
│  └──────┬───────┘  └──────┬───────┘  └─────┬─────┘               │
└─────────┼──────────────────┼────────────────┼─────────────────────┘
          │ POST /grid       │ POST /move     │ POST /import
          ▼                  ▼                ▼
┌────────────────────────────────────────────────────────────────────┐
│  Backend (Go Fiber)                                                │
│  ┌─────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │Handlers │→ │ Services │→ │   Repo   │→ │PostgreSQL│          │
│  └─────────┘  └────┬─────┘  └──────────┘  └──────────┘          │
│                     │                                              │
│                     ├──→ Redis Queue ──→ Worker (import/bulk)     │
│                     └──→ ClickHouse (analytics, future)           │
└────────────────────────────────────────────────────────────────────┘
```

## Tech Stack

| Layer     | Công nghệ                                     |
| --------- | --------------------------------------------- |
| Frontend  | React 18, TypeScript, Vite 7, Glide Data Grid |
| Backend   | Go 1.23, Fiber v2, pgx v5                     |
| Database  | PostgreSQL 16 (transactional)                 |
| Analytics | ClickHouse 24 (future)                        |
| Queue     | Redis 7 (async jobs)                          |
| Excel     | excelize v2 (`.xlsx`, date `MM/DD/YYYY`)      |

## Cấu trúc dự án

```
VG_wms_main_web/
├── cmd/
│   ├── api/main.go            # HTTP server (Fiber)
│   └── worker/main.go         # Background worker (import, bulk-update)
├── internal/
│   ├── domain/                # Entities, grid types, kanban stages
│   ├── grid/                  # SQL query builder (filter, sort, paginate)
│   ├── importer/              # Excel parser (4 loại file)
│   ├── queue/                 # Redis job queue
│   ├── repo/                  # PostgreSQL repositories
│   ├── service/               # Business logic layer
│   └── web/                   # HTTP handlers + routes
├── migrations/
│   ├── 001_init.up.sql        # Schema creation (12 bảng)
│   └── 001_init.down.sql      # Rollback
├── web/                       # Frontend React app
│   ├── src/
│   │   ├── api/client.ts      # API client (typed fetch)
│   │   ├── components/        # InventoryGrid, KanbanBoard, ImportPanel
│   │   └── types/             # TypeScript types
│   └── package.json
├── docker-compose.yml         # PostgreSQL, Redis, ClickHouse
├── .env.example
├── Makefile
└── go.mod
```

## Yêu cầu hệ thống

- Go >= 1.23
- Node.js >= 18
- Docker & Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI

## Cài đặt & Chạy

### 1. Clone & cấu hình

```bash
git clone <repo-url>
cd VG_wms_main_web
cp .env.example .env
```

### 2. Khởi động infrastructure

```bash
make docker-up
# → PostgreSQL :5432, Redis :6379, ClickHouse :8123/:9000
```

### 3. Chạy migration

```bash
make migrate
```

### 4. Chạy backend

```bash
# Terminal 1 — API server
make dev-api
# → http://localhost:8080

# Terminal 2 — Background worker
make dev-worker
```

### 5. Chạy frontend

```bash
cd web
npm install
npm run dev
# → http://localhost:5173
```

### Build production

```bash
# Backend
make build
# → bin/api, bin/worker

# Frontend
make web-build
# → web/dist/
```

## Cơ sở dữ liệu

### PostgreSQL — 12 bảng chính

| Bảng                  | Mục đích                                 |
| --------------------- | ---------------------------------------- |
| `products`            | Danh mục sản phẩm (master data)          |
| `inventory_main`      | Tồn kho hiện tại (số tồn, nhập, xuất)    |
| `inbound_items`       | Chi tiết phiếu nhập                      |
| `outbound_items`      | Chi tiết phiếu xuất                      |
| `inventory_movements` | Log mọi biến động kho (IN/OUT)           |
| `kanban_inbound`      | Thẻ Kanban nhập hàng                     |
| `kanban_outbound`     | Thẻ Kanban xuất hàng                     |
| `kanban_events`       | Lịch sử chuyển trạng thái Kanban         |
| `rule_config`         | Cấu hình quy tắc (tồn tối ưu, gap ratio) |
| `import_batches`      | Lịch sử import Excel                     |
| `async_jobs`          | Trạng thái job bất đồng bộ               |

### Quy tắc nghiệp vụ

- **FIFO** theo `ngay_nhan_hang` trong `inbound_items`
- **Tồn kho tối ưu:** 14 ngày (`rule_config.optimal_days`)
- **Gap ratio:** 10% (`rule_config.gap_ratio`)
- **Tồn âm được phép** — hệ thống sẽ cảnh báo (alert)

## API Endpoints

### Inventory

| Method | Path                         | Mô tả                              |
| ------ | ---------------------------- | ---------------------------------- |
| POST   | `/api/inventory/grid`        | Truy vấn lưới (filter, sort, page) |
| PATCH  | `/api/inventory/:ma_hang`    | Cập nhật 1 dòng inventory          |
| POST   | `/api/inventory/bulk-update` | Cập nhật hàng loạt → 202 + job_id  |

### Kanban

| Method | Path                            | Mô tả                  |
| ------ | ------------------------------- | ---------------------- |
| GET    | `/api/kanban/inbound`           | Danh sách thẻ nhập     |
| POST   | `/api/kanban/inbound`           | Tạo thẻ nhập mới       |
| POST   | `/api/kanban/inbound/:id/move`  | Chuyển trạng thái nhập |
| GET    | `/api/kanban/outbound`          | Danh sách thẻ xuất     |
| POST   | `/api/kanban/outbound`          | Tạo thẻ xuất mới       |
| POST   | `/api/kanban/outbound/:id/move` | Chuyển trạng thái xuất |

### Import & Jobs

| Method | Path                    | Mô tả                     |
| ------ | ----------------------- | ------------------------- |
| POST   | `/api/import/products`  | Import sản phẩm từ .xlsx  |
| POST   | `/api/import/inventory` | Import tồn kho từ .xlsx   |
| POST   | `/api/import/inbound`   | Import nhập hàng từ .xlsx |
| POST   | `/api/import/outbound`  | Import xuất hàng từ .xlsx |
| GET    | `/api/jobs/:id`         | Kiểm tra trạng thái job   |

### Grid Request Format (POST)

```json
{
  "startRow": 0,
  "endRow": 200,
  "sortModel": [{ "colId": "so_ton", "sort": "desc" }],
  "filterModel": {
    "ten_san_pham": {
      "filterType": "text",
      "type": "contains",
      "filter": "sữa"
    }
  }
}
```

### Grid Response

```json
{
  "rowsData": [...],
  "totalRowCount": 15432
}
```

## Luồng Kanban

### Nhập hàng (Inbound)

```
Cần nhập → Đã lên đơn → Đã duyệt → Đã về hàng
```

Khi chuyển sang **"Đã về hàng":**

1. Insert `inbound_items` (chi tiết nhập)
2. Cộng `inventory_main.so_ton` + `so_nhap`
3. Log `inventory_movements` (type = IN)

### Xuất hàng (Outbound)

```
Cần đẩy → Đã chốt đơn → Đã giao
```

Khi chuyển sang **"Đã giao":**

1. Insert `outbound_items` (chi tiết xuất)
2. Trừ `inventory_main.so_ton` + cộng `so_xuat` (cho phép âm → alert)
3. Log `inventory_movements` (type = OUT)

## Import Excel

Hệ thống hỗ trợ import 4 loại file `.xlsx`:

| File      | Sheet/Nội dung    | Xử lý                                       |
| --------- | ----------------- | ------------------------------------------- |
| Products  | Danh mục sản phẩm | Upsert vào `products` (PK: `ma_hang`)       |
| Inventory | Tồn kho ban đầu   | Upsert vào `inventory_main`                 |
| Inbound   | Lịch sử nhập hàng | Insert `inbound_items`, date: `MM/DD/YYYY`  |
| Outbound  | Lịch sử xuất hàng | Insert `outbound_items`, date: `MM/DD/YYYY` |

- Upload qua API → đẩy vào Redis queue → Worker xử lý bất đồng bộ
- Tracking trạng thái qua `import_batches` + `async_jobs`
- Poll `GET /api/jobs/:id` để theo dõi tiến trình

## Frontend

### Inventory Grid

- **Glide Data Grid** — virtualized, hỗ trợ 20k+ dòng, 60 FPS scroll
- Server-side pagination (200 dòng/page), sort, filter
- Inline editing với **optimistic UI** — cập nhật ngay, rollback nếu backend lỗi
- Prefetch khi scroll gần vùng chưa tải

### Kanban Board

- Hiển thị thẻ theo cột trạng thái
- Nút "Move to next" để chuyển trạng thái
- Hỗ trợ cả luồng Inbound và Outbound
- Alert khi tồn kho âm (Outbound)

### Import Panel

- 4 ô upload riêng biệt cho 4 loại file
- Hiển thị trạng thái job (pending → processing → done/failed)
- Auto-poll cập nhật tiến trình

## Biến môi trường

| Biến             | Mặc định                                                       | Mô tả                 |
| ---------------- | -------------------------------------------------------------- | --------------------- |
| `POSTGRES_DSN`   | `postgres://wms:wms_secret@localhost:5432/wms?sslmode=disable` | PostgreSQL connection |
| `REDIS_ADDR`     | `localhost:6379`                                               | Redis address         |
| `CLICKHOUSE_DSN` | `tcp://localhost:9000?database=wms`                            | ClickHouse (future)   |
| `PORT`           | `8080`                                                         | API server port       |

## Makefile Commands

| Command            | Mô tả                                   |
| ------------------ | --------------------------------------- |
| `make docker-up`   | Khởi động PostgreSQL, Redis, ClickHouse |
| `make docker-down` | Dừng tất cả containers                  |
| `make migrate`     | Chạy migration PostgreSQL               |
| `make dev-api`     | Chạy API server (dev mode)              |
| `make dev-worker`  | Chạy background worker (dev mode)       |
| `make build`       | Build binary production                 |
| `make web-dev`     | Chạy frontend dev server                |
| `make web-build`   | Build frontend production               |

## Giới hạn hiện tại

- Đơn kho (single warehouse), không hỗ trợ multi-warehouse
- Không quản lý batch/lô hàng, không quản lý hạn sử dụng
- ClickHouse chưa tích hợp (dành cho analytics phase sau)
- Chưa có authentication/authorization
- ML forecasting deferred (dự kiến: dự báo nhập hàng dựa trên tốc độ bán trung bình)

## License

Private — Internal use only.
