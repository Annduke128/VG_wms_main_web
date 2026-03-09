# PIPELINE.md — Data Pipeline

Tài liệu mô tả chi tiết luồng dữ liệu (data pipeline) của hệ thống WMS Web v1, từ nguồn dữ liệu đầu vào đến lưu trữ và hiển thị.

---

## Tổng quan Pipeline

```
                    ┌─────────────────────────────────────────────────────────────┐
                    │                     DATA SOURCES                            │
                    │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
                    │  │Products  │  │Inventory │  │Inbound   │  │Outbound  │   │
                    │  │.xlsx     │  │.xlsx     │  │.xlsx     │  │.xlsx     │   │
                    │  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘   │
                    └───────┼─────────────┼─────────────┼─────────────┼──────────┘
                            │             │             │             │
                            ▼             ▼             ▼             ▼
┌───────────────────────────────────────────────────────────────────────────────────┐
│  INGESTION LAYER                                                                  │
│                                                                                   │
│  POST /api/import/{type}                                                          │
│       │                                                                           │
│       ▼                                                                           │
│  ┌──────────────────┐     ┌─────────────────┐     ┌────────────────────┐         │
│  │ ImportService     │────→│ Redis Queue     │────→│ Worker Process     │         │
│  │ .EnqueueImport() │     │ (QueueImport)   │     │ .ProcessImport()   │         │
│  └──────────────────┘     └─────────────────┘     └────────┬───────────┘         │
│                                                             │                     │
│                                                             ▼                     │
│                                                   ┌──────────────────┐            │
│                                                   │ XLSX Parser      │            │
│                                                   │ (excelize v2)    │            │
│                                                   │ Date: MM/DD/YYYY │            │
│                                                   └────────┬─────────┘            │
│                                                             │                     │
│                                                             ▼                     │
│                                                   ┌──────────────────┐            │
│                                                   │ Validate + Map   │            │
│                                                   │ rows → entities  │            │
│                                                   └────────┬─────────┘            │
└────────────────────────────────────────────────────────────┼──────────────────────┘
                                                             │
                            ┌────────────────────────────────┼──────────────────┐
                            │                                │                  │
                            ▼                                ▼                  ▼
                    ┌──────────────┐                ┌──────────────┐   ┌──────────────┐
                    │  products    │                │inventory_main│   │inbound_items │
                    │  (UPSERT)    │                │  (UPSERT)    │   │outbound_items│
                    └──────────────┘                └──────────────┘   │  (INSERT)    │
                                                                      └──────────────┘
```

---

## 1. Excel Import Pipeline

### Luồng xử lý

```
User upload .xlsx
  │
  ▼
HTTP Handler (multipart/form-data, max 50MB)
  │
  ▼
ImportService.EnqueueImport()
  ├─ Tạo record trong import_batches (status: pending)
  ├─ Tạo record trong async_jobs (status: pending)
  └─ LPUSH job vào Redis queue "import"
  │
  ▼
Response: 202 Accepted + job_id
  │
  ▼
Worker (BRPOP từ Redis queue "import")
  │
  ▼
ImportService.ProcessImport()
  ├─ Đọc file .xlsx bằng excelize
  ├─ Parse theo loại:
  │   ├─ ParseProducts()   → []Product
  │   ├─ ParseInventory()  → []InventoryMain
  │   ├─ ParseInbound()    → []InboundItem   (date: MM/DD/YYYY)
  │   └─ ParseOutbound()   → []OutboundItem  (date: MM/DD/YYYY)
  ├─ Validate dữ liệu (required fields, data types)
  ├─ Bulk upsert/insert vào PostgreSQL
  │   ├─ products       → UPSERT (ON CONFLICT ma_hang)
  │   ├─ inventory_main → UPSERT (ON CONFLICT ma_hang)
  │   ├─ inbound_items  → INSERT
  │   └─ outbound_items → INSERT
  ├─ Cập nhật import_batches (row_count, status: done/failed)
  └─ Cập nhật async_jobs (status: done/failed, error nếu có)
```

### File format

| File      | Cột chính                                                                                                                                                                        |
| --------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Products  | `ma_hang`, `ten_san_pham`, `ma_bu`, `ma_cat`, `ma_nhom_hang`, `nhom_hang`, `don_vi_tinh`, `quy_cach`, `don_gia`, `vat`, `gia_niv`, `gia_nhap`, `hoa_hong`                        |
| Inventory | `ma_hang`, `ten_san_pham`, `so_ton`, `so_nhap`, `so_xuat`, `tien_ton`, `tien_nhap`, `tien_xuat`, `so_ngay_ton`, `luong_ban_binh_quan_ngay`                                       |
| Inbound   | `ma_hang`, `ten_san_pham`, `don_vi_tinh`, `quy_cach`, `so_luong`, `doanh_so`, `chiet_khau`, `so_luong_tra_lai`, `doanh_thu`, `von`, `lai_gop`, `ti_le_lai_gop`, `ngay_nhan_hang` |
| Outbound  | Cùng cấu trúc với Inbound                                                                                                                                                        |

### Tracking trạng thái

```
Frontend poll: GET /api/jobs/:id
  │
  ▼
async_jobs.status:
  pending → processing → done
                       → failed (kèm error message)
```

---

## 2. Kanban Transition Pipeline

Luồng dữ liệu khi thẻ Kanban chuyển trạng thái — đây là pipeline chính cho nghiệp vụ nhập/xuất hàng.

### Inbound Pipeline (Nhập hàng)

```
                    ┌──────────────────────────────────────────────────┐
                    │            KANBAN INBOUND STAGES                  │
                    │                                                   │
                    │  Cần nhập → Đã lên đơn → Đã duyệt → Đã về hàng │
                    └──────────────────────────────────┬───────────────┘
                                                       │
                                          Transition to "Đã về hàng"
                                                       │
                                                       ▼
                    ┌──────────────────────────────────────────────────┐
                    │  SIDE EFFECTS (trong 1 DB transaction)            │
                    │                                                   │
                    │  1. INSERT inbound_items                          │
                    │     ├─ ma_hang, so_luong, ngay_nhan_hang, ...    │
                    │     └─ Dữ liệu từ thẻ Kanban                    │
                    │                                                   │
                    │  2. UPDATE inventory_main                         │
                    │     ├─ so_ton  += so_luong                       │
                    │     └─ so_nhap += so_luong                       │
                    │                                                   │
                    │  3. INSERT inventory_movements                    │
                    │     ├─ type = 'IN'                               │
                    │     ├─ qty  = so_luong                           │
                    │     └─ created_at = NOW()                        │
                    │                                                   │
                    │  4. INSERT kanban_events                          │
                    │     ├─ from_stage = 'da_duyet'                   │
                    │     ├─ to_stage   = 'da_ve_hang'                 │
                    │     └─ user_id, created_at                       │
                    └──────────────────────────────────────────────────┘
```

### Outbound Pipeline (Xuất hàng)

```
                    ┌──────────────────────────────────────────┐
                    │        KANBAN OUTBOUND STAGES             │
                    │                                           │
                    │  Cần đẩy → Đã chốt đơn → Đã giao        │
                    └────────────────────────┬─────────────────┘
                                             │
                                Transition to "Đã giao"
                                             │
                                             ▼
                    ┌──────────────────────────────────────────┐
                    │  SIDE EFFECTS (trong 1 DB transaction)    │
                    │                                           │
                    │  1. INSERT outbound_items                 │
                    │     └─ ma_hang, so_luong, ...             │
                    │                                           │
                    │  2. UPDATE inventory_main                 │
                    │     ├─ so_ton  -= so_luong                │
                    │     ├─ so_xuat += so_luong                │
                    │     └─ ⚠ Cho phép so_ton < 0 → ALERT     │
                    │                                           │
                    │  3. INSERT inventory_movements             │
                    │     ├─ type = 'OUT'                       │
                    │     └─ qty  = so_luong                    │
                    │                                           │
                    │  4. INSERT kanban_events                   │
                    │     ├─ from_stage = 'da_chot_don'         │
                    │     └─ to_stage   = 'da_giao'             │
                    └──────────────────────────────────────────┘

        ┌──────────────────────────────────────────────────────┐
        │  NEGATIVE STOCK ALERT                                 │
        │                                                       │
        │  IF inventory_main.so_ton < 0 AFTER update:          │
        │    → Return alert flag in API response                │
        │    → Frontend hiển thị cảnh báo trên KanbanBoard      │
        │    → Không block transaction (vẫn cho phép xuất)      │
        └──────────────────────────────────────────────────────┘
```

---

## 3. Inventory Grid Pipeline

Luồng dữ liệu khi user tương tác với bảng tồn kho (đọc + ghi).

### Read Pipeline (Server-side Grid)

```
User scroll/filter/sort trên InventoryGrid
  │
  ▼
POST /api/inventory/grid
  │
  Body: { startRow, endRow, sortModel, filterModel }
  │
  ▼
GridBuilder.BuildQuery()
  │
  ├─ Parse filterModel → WHERE clauses
  │   ├─ contains    → WHERE col ILIKE '%value%'
  │   ├─ equals      → WHERE col = 'value'
  │   ├─ startsWith  → WHERE col ILIKE 'value%'
  │   ├─ endsWith    → WHERE col ILIKE '%value'
  │   ├─ inRange     → WHERE col BETWEEN min AND max
  │   └─ set (multi) → WHERE col IN ('a','b','c')
  │
  ├─ Parse sortModel → ORDER BY clauses
  │   └─ Column whitelist validation (chống SQL injection)
  │
  ├─ Pagination → OFFSET startRow LIMIT (endRow - startRow)
  │
  └─ Thực thi 2 queries song song:
      ├─ SELECT * FROM inventory_main WHERE ... ORDER BY ... LIMIT ...
      └─ SELECT COUNT(*) FROM inventory_main WHERE ...
  │
  ▼
Response: { rowsData: [...], totalRowCount: N }
  │
  ▼
InventoryGrid render dữ liệu (virtualized, 60 FPS)
```

### Write Pipeline (Inline Edit)

```
User edit cell trên InventoryGrid
  │
  ▼
Optimistic UI: Cập nhật giao diện ngay
  │
  ▼
PATCH /api/inventory/:ma_hang
  │
  Body: { field: value, ... }
  │
  ▼
InventoryRepo.UpdateInventoryItem()
  │
  ├─ UPDATE inventory_main SET ... WHERE ma_hang = $1
  └─ Return updated row
  │
  ▼
  ├─ Success → Confirm UI state
  └─ Error   → Rollback UI to previous value + hiển thị lỗi
```

### Bulk Update Pipeline (Async)

```
User trigger bulk update (vd: recalculate all)
  │
  ▼
POST /api/inventory/bulk-update
  │
  ▼
InventoryService.BulkUpdate()
  ├─ Tạo async_jobs (status: pending)
  ├─ LPUSH job vào Redis queue "bulk_update"
  └─ Return 202 + job_id
  │
  ├──→ UI lock affected rows (visual indicator)
  │
  ▼
Worker (BRPOP từ Redis queue "bulk_update")
  │
  ├─ Process từng batch
  ├─ UPDATE inventory_main ...
  └─ Cập nhật async_jobs (done/failed)
  │
  ▼
Frontend poll GET /api/jobs/:id
  │
  ├─ status: processing → tiếp tục poll
  ├─ status: done → unlock rows + refetch visible range
  └─ status: failed → unlock rows + hiển thị lỗi
```

---

## 4. Movement Logging Pipeline

Mọi thay đổi tồn kho đều được ghi log vào `inventory_movements` — đây là audit trail chính.

```
┌───────────────────────────────────────────────────────────────┐
│  MOVEMENT SOURCES                                              │
│                                                                │
│  1. Kanban Inbound → "Đã về hàng"                             │
│     └─ INSERT inventory_movements (type='IN', qty=+N)         │
│                                                                │
│  2. Kanban Outbound → "Đã giao"                               │
│     └─ INSERT inventory_movements (type='OUT', qty=N)         │
│                                                                │
│  3. (Future) Manual adjustment                                 │
│     └─ INSERT inventory_movements (type='ADJ', qty=±N)        │
│                                                                │
│  4. (Future) Import reconciliation                             │
│     └─ INSERT inventory_movements (type='IMPORT', qty=N)      │
└────────────────────────────────┬──────────────────────────────┘
                                 │
                                 ▼
                    ┌─────────────────────────┐
                    │  inventory_movements     │
                    │                          │
                    │  movement_id  (PK, UUID) │
                    │  ma_hang      (FK)       │
                    │  qty          (integer)  │
                    │  type         (IN/OUT)   │
                    │  created_at   (timestamp)│
                    └─────────────────────────┘
                                 │
                                 ▼ (Future)
                    ┌─────────────────────────┐
                    │  ClickHouse Sync         │
                    │  (analytics, reporting)  │
                    └─────────────────────────┘
```

---

## 5. Queue Architecture

Redis đóng vai trò message broker cho tất cả async operations.

```
┌─────────────────────────────────────────────────────────────────┐
│  REDIS QUEUES                                                    │
│                                                                  │
│  Queue: "import"                                                 │
│  ├─ Producer: ImportService.EnqueueImport()                     │
│  ├─ Consumer: Worker (cmd/worker/main.go)                       │
│  ├─ Payload:  { job_id, file_path, import_type }                │
│  └─ Pattern:  LPUSH (enqueue) / BRPOP (dequeue, blocking)      │
│                                                                  │
│  Queue: "bulk_update"                                            │
│  ├─ Producer: InventoryService.BulkUpdate()                     │
│  ├─ Consumer: Worker (cmd/worker/main.go)                       │
│  ├─ Payload:  { job_id, update_spec }                           │
│  └─ Pattern:  LPUSH / BRPOP                                     │
│                                                                  │
│  Job Lifecycle:                                                  │
│  pending → processing → done                                    │
│                       → failed (+ error message)                │
│                                                                  │
│  Tracking: async_jobs table (PostgreSQL)                         │
│  Polling:  GET /api/jobs/:id                                     │
└─────────────────────────────────────────────────────────────────┘
```

### Worker Process

```
cmd/worker/main.go
  │
  ├─ Goroutine 1: BRPOP "import" queue
  │   └─ ImportService.ProcessImport() per job
  │
  ├─ Goroutine 2: BRPOP "bulk_update" queue
  │   └─ InventoryService.ProcessBulkUpdate() per job
  │
  └─ Graceful shutdown on SIGINT/SIGTERM
      └─ Finish current job → exit
```

---

## 6. Data Flow Summary

### Tổng hợp các luồng dữ liệu

```
┌──────────────────────────────────────────────────────────────────────────┐
│                                                                          │
│  EXCEL FILES ──→ Import Pipeline ──→ PostgreSQL (products, inventory,   │
│                                       inbound_items, outbound_items)     │
│                                                                          │
│  KANBAN UI   ──→ Transition Pipeline ──→ PostgreSQL (inventory_main,   │
│                                           movements, kanban_events)      │
│                                                                          │
│  GRID UI     ──→ Query Pipeline    ──→ PostgreSQL → Grid Response       │
│              ──→ Edit Pipeline     ──→ PostgreSQL (optimistic UI)       │
│              ──→ Bulk Pipeline     ──→ Redis → Worker → PostgreSQL      │
│                                                                          │
│  PostgreSQL  ──→ (Future) ClickHouse Sync ──→ Analytics & Reports      │
│                                                                          │
│  PostgreSQL  ──→ (Future) ML Pipeline ──→ Forecast Inbound             │
│                   (avg sales velocity → predict restock needs)           │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

### Tính toàn vẹn dữ liệu

| Nguyên tắc                 | Cách thực hiện                                                               |
| -------------------------- | ---------------------------------------------------------------------------- |
| **Transactional writes**   | Kanban side-effects chạy trong 1 PostgreSQL transaction                      |
| **FIFO ordering**          | `inbound_items.ngay_nhan_hang` dùng để xác định thứ tự nhập trước xuất trước |
| **Audit trail**            | Mọi biến động kho ghi vào `inventory_movements`                              |
| **Negative stock allowed** | Không block xuất hàng khi tồn âm, chỉ alert                                  |
| **Async isolation**        | Import & bulk update chạy async qua Redis, không block API response          |
| **Idempotent imports**     | Products & inventory dùng UPSERT (ON CONFLICT), an toàn khi re-import        |
| **Job tracking**           | Mọi async operation có `async_jobs` record để track status                   |

### Business Rules Pipeline

```
rule_config table:
  ├─ optimal_days = 14
  └─ gap_ratio    = 0.10

Áp dụng:
  ├─ Kanban "Cần nhập": SKU có so_ton < (luong_ban_binh_quan_ngay × optimal_days)
  ├─ Kanban "Cần đẩy": SKU có so_ngay_ton > threshold hoặc slow-moving
  └─ Gap alert: Khi chênh lệch tồn kho thực tế vs. tối ưu > gap_ratio × optimal_stock
```

---

## 7. Future Pipeline Extensions

### ClickHouse Analytics (Planned)

```
PostgreSQL (inventory_movements)
  │
  ▼ (CDC or periodic sync)
ClickHouse (analytics tables)
  │
  ├─ Aggregated sales velocity per SKU
  ├─ Inventory turnover reports
  ├─ Slow-moving stock identification
  └─ Historical trend analysis
```

### ML Forecast Pipeline (Deferred)

```
ClickHouse (historical data)
  │
  ▼
Feature extraction
  ├─ luong_ban_binh_quan_ngay (7d, 14d, 30d rolling)
  ├─ Seasonality patterns
  └─ Lead time per supplier
  │
  ▼
Forecast model
  ├─ Predict optimal restock quantity
  └─ Predict restock timing
  │
  ▼
Kanban automation
  └─ Auto-create "Cần nhập" cards when predicted stockout
```

---

_Cập nhật lần cuối: 2026-03-09_
