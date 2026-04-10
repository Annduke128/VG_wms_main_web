# Multi-Warehouse Phase 1 — Implementation Plan (v2)

> **For Claude:** REQUIRED SUB-SKILL: Use skill({ name: "executing-plans" }) to implement this plan task-by-task.

**Goal:** Add multi-warehouse support so every stock-related operation (inventory, lots, orders, combos, accessories, metrics, dashboard) is scoped to a specific warehouse.

**Architecture:** Add a `warehouses` table with auto-generated code and editable name. Add `warehouse_id` FK column to all stock tables (`inventory_main`, `inventory_lots`, `inventory_movements`, `inbound_items`, `outbound_items`, `inventory_thresholds`, `inventory_lbbq_history`, `combo_inventory`, `combo_transactions`, `combo_component_movements`, `accessory_inventory`, `accessory_movements`). Backfill existing data to a default warehouse. Update all repo/service/handler/worker layers to pass `warehouseID`. Frontend uses Zustand store for active warehouse with AG Grid purge on switch.

**Tech Stack:** Go 1.22+, Gin, pgx v5, PostgreSQL 15+, React 18, TypeScript, Zustand, AG Grid Enterprise, Vite

---

## Decisions

| Decision | Value |
|----------|-------|
| Warehouse code | Auto-generated: `WH-XXXX` via `warehouse_code_seq` |
| Warehouse name | Editable on UI |
| Default warehouse | `WH-0001` / "Kho mac dinh" (backfill all existing data) |
| Combo inventory | Per warehouse (`combo_inventory.warehouse_id`) |
| Accessory inventory | Per warehouse (`accessory_inventory.warehouse_id`) |
| Dashboard KPIs | Filtered by active warehouse |
| Master data (products, combo_master, combo_bom_*, accessories) | Shared across warehouses — NO warehouse_id |
| warehouse_id passing | Query parameter `?warehouse_id=N` on all stock-related endpoints |

---

## Must-Haves

**Goal:** Every stock operation scoped to a warehouse. UI allows switching warehouses.

### Observable Truths

1. User can create/edit/list warehouses
2. All inventory queries return data for the active warehouse only
3. Inbound/outbound orders are created against a specific warehouse
4. FIFO lot allocation happens within one warehouse
5. Combo create/cancel/out/return operates within one warehouse
6. Accessory stock operates within one warehouse
7. Dashboard KPIs/charts show data for the active warehouse
8. Metrics recalc operates per warehouse (no cross-warehouse aggregation)
9. Excel import targets a specific warehouse
10. Worker async jobs carry warehouse_id in payload

### Required Artifacts

| Artifact | Provides | Path |
|----------|----------|------|
| Migration 007 | Schema changes + backfill | `migrations/007_multi_warehouse.up.sql` |
| Warehouse domain type | Go struct | `internal/domain/entities.go` |
| Warehouse repo | CRUD queries | `internal/repo/warehouse.go` |
| Updated repos (8 files) | warehouse-scoped queries | `internal/repo/*.go` |
| Updated services (5 files) | warehouse-scoped business logic | `internal/service/*.go` |
| Updated handlers + routes | warehouse_id extraction | `internal/web/*.go` |
| Updated worker | warehouse_id in queue payloads | `cmd/worker/main.go` |
| Warehouse store (frontend) | Active warehouse state | `web/src/stores/warehouseStore.ts` |
| Warehouse selector component | UI dropdown | `web/src/components/WarehouseSelector.tsx` |
| Updated API client | warehouse_id params | `web/src/api/client.ts` |
| Updated TS types | Warehouse type | `web/src/types/warehouse.ts` |

### Key Links (Risk Areas)

| From | To | Via | Risk |
|------|-----|-----|------|
| Migration PK change | All ON CONFLICT clauses | composite PK | Must update every ON CONFLICT to include warehouse_id |
| Migration PK change | All WHERE clauses in repos | FK/PK reference | Every bare `WHERE ma_hang = $N` must add `AND warehouse_id = $N` |
| Worker queue payloads | Repo functions | JSON payload | Must add warehouse_id to payload struct |
| AG Grid datasource | API endpoints | query params | Must reset grid on warehouse switch |
| combo.go service inline SQL | inventory_main | direct SQL | Lines 151, 259, 271 have bare WHERE — must add warehouse filter |

---

## Task Dependencies

```
Task 1 (Migration): needs nothing, creates schema
Task 2 (Domain types): needs nothing, creates Go structs
Task 3 (Warehouse repo): needs Task 1+2, creates warehouse CRUD
Task 4 (inventory.go): needs Task 1+2, modifies repo
Task 5 (import.go): needs Task 1+2, modifies repo
Task 6 (orders.go): needs Task 1+2, modifies repo
Task 7 (inventory_metrics.go): needs Task 1+2, modifies repo
Task 8 (combo.go repo): needs Task 1+2, modifies repo
Task 9 (accessory.go repo): needs Task 1+2, modifies repo
Task 10 (dashboard.go repo): needs Task 1+2, modifies repo
Task 11 (reset.go): needs Task 1, modifies repo
Task 12 (inventory service): needs Task 4+7, modifies service
Task 13 (order service): needs Task 6+7, modifies service
Task 14 (combo service): needs Task 8+9+7, modifies service
Task 15 (dashboard service): needs Task 10, modifies service
Task 16 (importer service): needs Task 5+7, modifies service
Task 17 (handlers + routes): needs Tasks 3+12-16, modifies web layer
Task 18 (worker): needs Tasks 4+5+7+12+16, modifies worker
Task 19 (TS types + API client): needs Task 17, modifies frontend
Task 20 (Warehouse store + selector): needs Task 19, creates frontend
Task 21 (Inventory view): needs Task 20, modifies frontend
Task 22 (Orders/Combo/Dashboard views): needs Task 20, modifies frontend
Task 23 (Verification): needs all above

Wave 1: Tasks 1, 2 (parallel — schema + types)
Wave 2: Tasks 3-11 (parallel — all repo changes)
Wave 3: Tasks 12-16 (parallel — all service changes)
Wave 4: Tasks 17, 18 (parallel — handlers + worker)
Wave 5: Tasks 19-22 (sequential — frontend)
Wave 6: Task 23 (verification)
```

---

## Task 1: Migration 007 — Multi-Warehouse Schema

**Files:**
- Create: `migrations/007_multi_warehouse.up.sql`
- Create: `migrations/007_multi_warehouse.down.sql`

**Step 1: Create up migration**

```sql
-- migrations/007_multi_warehouse.up.sql

BEGIN;

-- ============================================================
-- 1. warehouses table + sequence for auto-generated code
-- ============================================================
CREATE SEQUENCE warehouse_code_seq START WITH 2;
-- Start at 2 because WH-0001 is the default warehouse inserted below

CREATE TABLE warehouses (
    id          BIGSERIAL PRIMARY KEY,
    code        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    address     TEXT NOT NULL DEFAULT '',
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default warehouse
INSERT INTO warehouses (id, code, name) VALUES (1, 'WH-0001', 'Kho mac dinh');

-- ============================================================
-- 2. Add warehouse_id to stock tables + backfill
-- ============================================================

-- 2a. inventory_main: change PK from (ma_hang) to (ma_hang, warehouse_id)
ALTER TABLE inventory_main ADD COLUMN warehouse_id BIGINT;
UPDATE inventory_main SET warehouse_id = 1;
ALTER TABLE inventory_main ALTER COLUMN warehouse_id SET NOT NULL;
ALTER TABLE inventory_main ADD CONSTRAINT fk_inventory_main_warehouse
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id);
ALTER TABLE inventory_main DROP CONSTRAINT inventory_main_pkey;
ALTER TABLE inventory_main ADD PRIMARY KEY (ma_hang, warehouse_id);

-- 2b. inventory_lots: change unique constraint
-- IMPORTANT: The existing unique is an INDEX named idx_lots_ma_hang_batch, not a CONSTRAINT
ALTER TABLE inventory_lots ADD COLUMN warehouse_id BIGINT;
UPDATE inventory_lots SET warehouse_id = 1;
ALTER TABLE inventory_lots ALTER COLUMN warehouse_id SET NOT NULL;
ALTER TABLE inventory_lots ADD CONSTRAINT fk_inventory_lots_warehouse
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id);
DROP INDEX IF EXISTS idx_lots_ma_hang_batch;
ALTER TABLE inventory_lots ADD CONSTRAINT inventory_lots_ma_hang_batch_warehouse_key
    UNIQUE (ma_hang, batch_code, warehouse_id);
-- Replace partial index for FIFO queries (include warehouse_id for coverage)
DROP INDEX IF EXISTS idx_lots_remaining;
CREATE INDEX idx_lots_remaining
    ON inventory_lots (ma_hang, warehouse_id, qty_remaining)
    WHERE qty_remaining > 0;

-- 2c. inventory_movements
ALTER TABLE inventory_movements ADD COLUMN warehouse_id BIGINT;
UPDATE inventory_movements SET warehouse_id = 1;
ALTER TABLE inventory_movements ALTER COLUMN warehouse_id SET NOT NULL;
ALTER TABLE inventory_movements ADD CONSTRAINT fk_inventory_movements_warehouse
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id);
CREATE INDEX idx_movements_warehouse ON inventory_movements (warehouse_id);

-- 2d. inbound_items
ALTER TABLE inbound_items ADD COLUMN warehouse_id BIGINT;
UPDATE inbound_items SET warehouse_id = 1;
ALTER TABLE inbound_items ALTER COLUMN warehouse_id SET NOT NULL;
ALTER TABLE inbound_items ADD CONSTRAINT fk_inbound_items_warehouse
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id);
CREATE INDEX idx_inbound_items_warehouse ON inbound_items (warehouse_id);

-- 2e. outbound_items
ALTER TABLE outbound_items ADD COLUMN warehouse_id BIGINT;
UPDATE outbound_items SET warehouse_id = 1;
ALTER TABLE outbound_items ALTER COLUMN warehouse_id SET NOT NULL;
ALTER TABLE outbound_items ADD CONSTRAINT fk_outbound_items_warehouse
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id);
CREATE INDEX idx_outbound_items_warehouse ON outbound_items (warehouse_id);

-- 2f. inventory_thresholds
ALTER TABLE inventory_thresholds ADD COLUMN warehouse_id BIGINT;
UPDATE inventory_thresholds SET warehouse_id = 1;
ALTER TABLE inventory_thresholds ALTER COLUMN warehouse_id SET NOT NULL;
ALTER TABLE inventory_thresholds ADD CONSTRAINT fk_inventory_thresholds_warehouse
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id);
CREATE INDEX idx_thresholds_warehouse ON inventory_thresholds (ma_hang, warehouse_id);

-- 2g. inventory_lbbq_history: change unique constraint
ALTER TABLE inventory_lbbq_history ADD COLUMN warehouse_id BIGINT;
UPDATE inventory_lbbq_history SET warehouse_id = 1;
ALTER TABLE inventory_lbbq_history ALTER COLUMN warehouse_id SET NOT NULL;
ALTER TABLE inventory_lbbq_history ADD CONSTRAINT fk_lbbq_history_warehouse
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id);
-- Drop old unique (could be constraint or index — try both)
ALTER TABLE inventory_lbbq_history DROP CONSTRAINT IF EXISTS inventory_lbbq_history_ma_hang_month_start_key;
DROP INDEX IF EXISTS inventory_lbbq_history_ma_hang_month_start_key;
ALTER TABLE inventory_lbbq_history ADD CONSTRAINT inventory_lbbq_history_sku_wh_month_key
    UNIQUE (ma_hang, warehouse_id, month_start);

-- 2h. combo_inventory: change PK
ALTER TABLE combo_inventory ADD COLUMN warehouse_id BIGINT;
UPDATE combo_inventory SET warehouse_id = 1;
ALTER TABLE combo_inventory ALTER COLUMN warehouse_id SET NOT NULL;
ALTER TABLE combo_inventory ADD CONSTRAINT fk_combo_inventory_warehouse
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id);
ALTER TABLE combo_inventory DROP CONSTRAINT combo_inventory_pkey;
ALTER TABLE combo_inventory ADD PRIMARY KEY (ma_combo, warehouse_id);

-- 2i. combo_transactions
ALTER TABLE combo_transactions ADD COLUMN warehouse_id BIGINT;
UPDATE combo_transactions SET warehouse_id = 1;
ALTER TABLE combo_transactions ALTER COLUMN warehouse_id SET NOT NULL;
ALTER TABLE combo_transactions ADD CONSTRAINT fk_combo_transactions_warehouse
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id);
CREATE INDEX idx_combo_transactions_warehouse ON combo_transactions (warehouse_id);

-- 2j. combo_component_movements
ALTER TABLE combo_component_movements ADD COLUMN warehouse_id BIGINT;
UPDATE combo_component_movements SET warehouse_id = 1;
ALTER TABLE combo_component_movements ALTER COLUMN warehouse_id SET NOT NULL;
ALTER TABLE combo_component_movements ADD CONSTRAINT fk_combo_component_movements_warehouse
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id);

-- 2k. accessory_inventory: change PK
ALTER TABLE accessory_inventory ADD COLUMN warehouse_id BIGINT;
UPDATE accessory_inventory SET warehouse_id = 1;
ALTER TABLE accessory_inventory ALTER COLUMN warehouse_id SET NOT NULL;
ALTER TABLE accessory_inventory ADD CONSTRAINT fk_accessory_inventory_warehouse
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id);
ALTER TABLE accessory_inventory DROP CONSTRAINT accessory_inventory_pkey;
ALTER TABLE accessory_inventory ADD PRIMARY KEY (ma_phu_kien, warehouse_id);

-- 2l. accessory_movements
ALTER TABLE accessory_movements ADD COLUMN warehouse_id BIGINT;
UPDATE accessory_movements SET warehouse_id = 1;
ALTER TABLE accessory_movements ALTER COLUMN warehouse_id SET NOT NULL;
ALTER TABLE accessory_movements ADD CONSTRAINT fk_accessory_movements_warehouse
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id);
CREATE INDEX idx_accessory_movements_warehouse ON accessory_movements (warehouse_id);

-- ============================================================
-- 3. Recreate inventory_grid VIEW with warehouse_id
-- ============================================================
DROP VIEW IF EXISTS inventory_grid;
CREATE VIEW inventory_grid AS
SELECT
    im.ma_hang,
    im.ten_san_pham,
    im.so_ton,
    im.so_nhap,
    im.so_xuat,
    im.tien_ton,
    im.tien_nhap,
    im.tien_xuat,
    im.so_ngay_ton,
    im.luong_ban_binh_quan_ngay,
    im.so_ngay_ton_ban,
    im.warehouse_id,
    COALESCE(p.don_gia, 0)         AS don_gia,
    COALESCE(p.ma_bu, '')          AS ma_bu,
    COALESCE(p.ma_nhom_hang, '')   AS ma_nhom_hang
FROM inventory_main im
LEFT JOIN products p ON im.ma_hang = p.ma_hang;

-- ============================================================
-- 4. Additional indexes for common query patterns
-- ============================================================
CREATE INDEX idx_inventory_main_warehouse ON inventory_main (warehouse_id);
CREATE INDEX idx_inventory_lots_warehouse ON inventory_lots (warehouse_id);

COMMIT;
```

**Step 2: Create down migration**

```sql
-- migrations/007_multi_warehouse.down.sql
-- WARNING: This migration is DESTRUCTIVE if multiple warehouses exist.
-- It keeps only data from warehouse_id=1 (default warehouse) and drops the rest.

BEGIN;

-- 1. Delete non-default warehouse data (DESTRUCTIVE)
DELETE FROM combo_component_movements WHERE warehouse_id != 1;
DELETE FROM combo_transactions WHERE warehouse_id != 1;
DELETE FROM combo_inventory WHERE warehouse_id != 1;
DELETE FROM accessory_movements WHERE warehouse_id != 1;
DELETE FROM accessory_inventory WHERE warehouse_id != 1;
DELETE FROM inventory_movements WHERE warehouse_id != 1;
DELETE FROM inventory_lbbq_history WHERE warehouse_id != 1;
DELETE FROM inventory_thresholds WHERE warehouse_id != 1;
DELETE FROM inbound_items WHERE warehouse_id != 1;
DELETE FROM outbound_items WHERE warehouse_id != 1;
DELETE FROM inventory_lots WHERE warehouse_id != 1;
DELETE FROM inventory_main WHERE warehouse_id != 1;

-- 2. Restore inventory_main PK
ALTER TABLE inventory_main DROP CONSTRAINT inventory_main_pkey;
ALTER TABLE inventory_main ADD PRIMARY KEY (ma_hang);

-- 3. Restore inventory_lots unique constraint
ALTER TABLE inventory_lots DROP CONSTRAINT IF EXISTS inventory_lots_ma_hang_batch_warehouse_key;
DROP INDEX IF EXISTS idx_lots_remaining;
CREATE UNIQUE INDEX idx_lots_ma_hang_batch ON inventory_lots(ma_hang, batch_code);
CREATE INDEX idx_lots_remaining ON inventory_lots(ma_hang, qty_remaining) WHERE qty_remaining > 0;

-- 4. Restore combo_inventory PK
ALTER TABLE combo_inventory DROP CONSTRAINT combo_inventory_pkey;
ALTER TABLE combo_inventory ADD PRIMARY KEY (ma_combo);

-- 5. Restore accessory_inventory PK
ALTER TABLE accessory_inventory DROP CONSTRAINT accessory_inventory_pkey;
ALTER TABLE accessory_inventory ADD PRIMARY KEY (ma_phu_kien);

-- 6. Restore inventory_lbbq_history unique
ALTER TABLE inventory_lbbq_history DROP CONSTRAINT IF EXISTS inventory_lbbq_history_sku_wh_month_key;
ALTER TABLE inventory_lbbq_history ADD CONSTRAINT inventory_lbbq_history_ma_hang_month_start_key
    UNIQUE (ma_hang, month_start);

-- 7. Drop warehouse_id columns (cascade drops FKs)
ALTER TABLE inventory_main DROP COLUMN warehouse_id;
ALTER TABLE inventory_lots DROP COLUMN warehouse_id;
ALTER TABLE inventory_movements DROP COLUMN warehouse_id;
ALTER TABLE inbound_items DROP COLUMN warehouse_id;
ALTER TABLE outbound_items DROP COLUMN warehouse_id;
ALTER TABLE inventory_thresholds DROP COLUMN warehouse_id;
ALTER TABLE inventory_lbbq_history DROP COLUMN warehouse_id;
ALTER TABLE combo_inventory DROP COLUMN warehouse_id;
ALTER TABLE combo_transactions DROP COLUMN warehouse_id;
ALTER TABLE combo_component_movements DROP COLUMN warehouse_id;
ALTER TABLE accessory_inventory DROP COLUMN warehouse_id;
ALTER TABLE accessory_movements DROP COLUMN warehouse_id;

-- 8. Drop indexes (some dropped with columns above, but be explicit)
DROP INDEX IF EXISTS idx_inventory_main_warehouse;
DROP INDEX IF EXISTS idx_inventory_lots_warehouse;
DROP INDEX IF EXISTS idx_movements_warehouse;
DROP INDEX IF EXISTS idx_inbound_items_warehouse;
DROP INDEX IF EXISTS idx_outbound_items_warehouse;
DROP INDEX IF EXISTS idx_thresholds_warehouse;
DROP INDEX IF EXISTS idx_combo_transactions_warehouse;
DROP INDEX IF EXISTS idx_accessory_movements_warehouse;

-- 9. Recreate original inventory_grid VIEW
DROP VIEW IF EXISTS inventory_grid;
CREATE VIEW inventory_grid AS
SELECT
    im.ma_hang,
    im.ten_san_pham,
    im.so_ton,
    im.so_nhap,
    im.so_xuat,
    im.tien_ton,
    im.tien_nhap,
    im.tien_xuat,
    im.so_ngay_ton,
    im.luong_ban_binh_quan_ngay,
    im.so_ngay_ton_ban,
    COALESCE(p.don_gia, 0) AS don_gia,
    COALESCE(p.ma_bu, '') AS ma_bu,
    COALESCE(p.ma_nhom_hang, '') AS ma_nhom_hang
FROM inventory_main im
LEFT JOIN products p ON im.ma_hang = p.ma_hang;

-- 10. Drop warehouses table + sequence
DROP TABLE IF EXISTS warehouses;
DROP SEQUENCE IF EXISTS warehouse_code_seq;

COMMIT;
```

**Step 3: Verify migration syntax**

Run: `make migrate && make migrate-status`
Expected: Version 7, clean (not dirty)

**Step 4: Commit**

```bash
git add migrations/007_multi_warehouse.up.sql migrations/007_multi_warehouse.down.sql
git commit -m "feat(multi-wh): add migration 007 — warehouses table + warehouse_id on all stock tables"
```

---

## Task 2: Domain Types — Warehouse Struct + Updated Entities

**Files:**
- Modify: `internal/domain/entities.go`
- Modify: `internal/domain/orders.go`
- Modify: `internal/domain/combo.go`
- Modify: `internal/domain/grid.go`
- Modify: `internal/domain/dashboard.go`

**Step 1: Add Warehouse struct to entities.go**

Add after the `ImportBatch` struct:

```go
// Warehouse represents a physical warehouse location
type Warehouse struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateWarehouseRequest is the request body for creating a warehouse
type CreateWarehouseRequest struct {
	Name    string `json:"name" binding:"required"`
	Address string `json:"address"`
}

// UpdateWarehouseRequest is the request body for updating a warehouse
type UpdateWarehouseRequest struct {
	Name    *string `json:"name"`
	Address *string `json:"address"`
}
```

**Step 2: Add WarehouseID to existing stock structs in entities.go**

Add `WarehouseID int64 `json:"warehouse_id"`` field to:
- `InventoryMain`
- `InboundItem`
- `OutboundItem`
- `InventoryLot`
- `InventoryThreshold`
- `InventoryMovement`

**Step 3: Add WarehouseID to request types in orders.go**

Add `WarehouseID int64 `json:"warehouse_id" binding:"required"`` to:
- `CreateInboundRequest`
- `CreateOutboundRequest`

**Step 4: Add WarehouseID to combo types in combo.go**

Add `WarehouseID int64 `json:"warehouse_id"`` to:
- `ComboInventory`
- `ComboTransaction`
- `AccessoryInventory`
- `AccessoryMovement`

Add `WarehouseID int64 `json:"warehouse_id" binding:"required"`` to:
- `CreateComboRequest`
- `CancelComboRequest`
- `ComboOutRequest`
- `ComboReturnRequest`
- `AccessoryStockInRequest`

**Step 5: Add WarehouseID to grid.go**

Add `WarehouseID int64 `json:"warehouse_id"`` to `BulkUpdateRequest`.

**Step 6: Add WarehouseID to dashboard.go**

Add `WarehouseID int64 `json:"warehouse_id" binding:"required"`` to `ThresholdRequest`.

**Step 7: Verify compilation**

Run: `go build ./...`
Expected: No errors (structs are additive, no callers changed yet)

**Step 8: Commit**

```bash
git add internal/domain/
git commit -m "feat(multi-wh): add Warehouse domain type + WarehouseID to all stock entities"
```

---

## Task 3: Warehouse Repo — CRUD Operations

**Files:**
- Create: `internal/repo/warehouse.go`

**Step 1: Create warehouse repo file**

```go
package repo

import (
	"context"
	"fmt"

	"github.com/vg-leanmfg/wms/internal/domain"
)

func (r *PostgresRepo) ListWarehouses(ctx context.Context) ([]domain.Warehouse, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT id, code, name, address, is_active, created_at, updated_at
		FROM warehouses ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list warehouses: %w", err)
	}
	defer rows.Close()

	var warehouses []domain.Warehouse
	for rows.Next() {
		var w domain.Warehouse
		if err := rows.Scan(&w.ID, &w.Code, &w.Name, &w.Address, &w.IsActive, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan warehouse: %w", err)
		}
		warehouses = append(warehouses, w)
	}
	return warehouses, rows.Err()
}

func (r *PostgresRepo) GetWarehouse(ctx context.Context, id int64) (*domain.Warehouse, error) {
	var w domain.Warehouse
	err := r.Pool.QueryRow(ctx, `
		SELECT id, code, name, address, is_active, created_at, updated_at
		FROM warehouses WHERE id = $1`, id).
		Scan(&w.ID, &w.Code, &w.Name, &w.Address, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get warehouse %d: %w", id, err)
	}
	return &w, nil
}

func (r *PostgresRepo) CreateWarehouse(ctx context.Context, req domain.CreateWarehouseRequest) (*domain.Warehouse, error) {
	var w domain.Warehouse
	err := r.Pool.QueryRow(ctx, `
		INSERT INTO warehouses (code, name, address)
		VALUES ('WH-' || LPAD(nextval('warehouse_code_seq')::TEXT, 4, '0'), $1, $2)
		RETURNING id, code, name, address, is_active, created_at, updated_at`,
		req.Name, req.Address).
		Scan(&w.ID, &w.Code, &w.Name, &w.Address, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create warehouse: %w", err)
	}
	return &w, nil
}

func (r *PostgresRepo) UpdateWarehouse(ctx context.Context, id int64, req domain.UpdateWarehouseRequest) (*domain.Warehouse, error) {
	var w domain.Warehouse
	err := r.Pool.QueryRow(ctx, `
		UPDATE warehouses SET
			name = COALESCE($2, name),
			address = COALESCE($3, address),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, code, name, address, is_active, created_at, updated_at`,
		id, req.Name, req.Address).
		Scan(&w.ID, &w.Code, &w.Name, &w.Address, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update warehouse %d: %w", id, err)
	}
	return &w, nil
}

// WarehouseExists checks if a warehouse exists and is active.
func (r *PostgresRepo) WarehouseExists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM warehouses WHERE id = $1 AND is_active = TRUE)`,
		id).Scan(&exists)
	return exists, err
}

// InitWarehouseInventory seeds zero-stock rows for all existing combos and
// accessories in a new warehouse. Call after CreateWarehouse.
func (r *PostgresRepo) InitWarehouseInventory(ctx context.Context, warehouseID int64) error {
	// Seed combo_inventory for all active combo masters
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO combo_inventory (ma_combo, warehouse_id, so_ton, so_nhap, so_xuat, so_tra)
		SELECT ma_combo, $1, 0, 0, 0, 0 FROM combo_master WHERE active = TRUE
		ON CONFLICT (ma_combo, warehouse_id) DO NOTHING`, warehouseID)
	if err != nil {
		return fmt.Errorf("init combo inventory for warehouse %d: %w", warehouseID, err)
	}

	// Seed accessory_inventory for all accessories
	_, err = r.Pool.Exec(ctx, `
		INSERT INTO accessory_inventory (ma_phu_kien, warehouse_id, so_ton)
		SELECT ma_phu_kien, $1, 0 FROM accessories
		ON CONFLICT (ma_phu_kien, warehouse_id) DO NOTHING`, warehouseID)
	if err != nil {
		return fmt.Errorf("init accessory inventory for warehouse %d: %w", warehouseID, err)
	}

	return nil
}
```

**Step 2: Verify compilation**

Run: `go build ./...`

**Step 3: Commit**

```bash
git add internal/repo/warehouse.go
git commit -m "feat(multi-wh): add warehouse repo CRUD + InitWarehouseInventory"
```

---

## Task 4: Update `internal/repo/inventory.go` — Warehouse-Scoped Queries

**Files:**
- Modify: `internal/repo/inventory.go`
- Modify: `internal/grid/builder.go` (if grid SQL is built dynamically)

**Changes:**

**4a. Grid query:** The grid reads from `inventory_grid` VIEW which now has `warehouse_id`. Add `AND warehouse_id = $N` to the WHERE clause builder. If grid builder in `internal/grid/` builds SQL dynamically, add a mandatory base filter for warehouse_id. If grid uses FilterModel, inject `warehouse_id` filter at handler level before passing to repo.

**4b. `UpdateInventoryItem` (line 72):** Currently uses dynamic format string:
```go
query := fmt.Sprintf("UPDATE inventory_main SET %s WHERE ma_hang = $%d", setClauses, paramIdx)
```
Change to:
```go
query := fmt.Sprintf("UPDATE inventory_main SET %s WHERE ma_hang = $%d AND warehouse_id = $%d", setClauses, paramIdx, paramIdx+1)
args = append(args, warehouseID)
```
Function signature: add `warehouseID int64` parameter.

**4c. `GetFilterOptions`:** Add `WHERE warehouse_id = $1` to filter option queries.

**4d. `ExportRows`:** Add `AND warehouse_id = $N` to export query WHERE clause.

**Verify:** `go build ./internal/repo/...`

**Commit:**
```bash
git add internal/repo/inventory.go internal/grid/
git commit -m "feat(multi-wh): scope inventory repo queries to warehouse_id"
```

---

## Task 5: Update `internal/repo/import.go` — Warehouse-Scoped Upserts

**Files:**
- Modify: `internal/repo/import.go`

**Changes (5 ON CONFLICT statements):**

**5a. `UpsertProduct` (line 38):** `products` is master data — NO change needed.

**5b. `UpsertInventory` (line 64):**
- Add `warehouse_id` to INSERT column list
- Change `ON CONFLICT (ma_hang)` to `ON CONFLICT (ma_hang, warehouse_id)`
- Re-index all `$N` parameter references
- Function signature: add `warehouseID int64`

**5c. Second `UpsertInventory` variant (line 142):** Same pattern as 5b.

**5d. `ImportInventoryFull` (line 162):** Same pattern as 5b.

**5e. `UpsertLot` (line 188):**
- Add `warehouse_id` to INSERT column list
- Change `ON CONFLICT (ma_hang, batch_code)` to `ON CONFLICT (ma_hang, batch_code, warehouse_id)`
- Function signature: add `warehouseID int64`

**Verify:** `go build ./internal/repo/...`

**Commit:**
```bash
git add internal/repo/import.go
git commit -m "feat(multi-wh): scope import repo upserts to warehouse_id"
```

---

## Task 6: Update `internal/repo/orders.go` — Warehouse-Scoped Orders + FIFO + Thresholds

**Files:**
- Modify: `internal/repo/orders.go`

**Changes (9+ bare WHERE clauses):**

**6a. `UpsertLotInbound` (line 35):**
- `ON CONFLICT (ma_hang, batch_code)` becomes `ON CONFLICT (ma_hang, batch_code, warehouse_id)`
- Add `warehouse_id` to INSERT columns
- Function signature: add `warehouseID int64`

**6b. FIFO query `GetLotsForFIFO` (line 50):**
- `WHERE ma_hang = $1 AND qty_remaining > 0` becomes `WHERE ma_hang = $1 AND warehouse_id = $2 AND qty_remaining > 0`
- Function signature: add `warehouseID int64`

**6c. `DeductLot` (line 68):** Uses `WHERE id = $2` — unique PK. No change needed.

**6d. `UpdateInventoryMainInbound` (line 102):**
- `WHERE ma_hang = $2` becomes `WHERE ma_hang = $2 AND warehouse_id = $3`
- Function signature: add `warehouseID int64`

**6e. INSERT fallback for new inventory (line 110-113):**
- Add `warehouse_id` to INSERT column list:
```sql
INSERT INTO inventory_main (ma_hang, warehouse_id, ten_san_pham, so_ton, so_nhap, so_xuat)
SELECT $1, $3, ten_san_pham, $2, $2, 0 FROM products WHERE ma_hang = $1
```

**6f. `UpdateInventoryMainOutbound` (line 126):**
- `WHERE ma_hang = $2` becomes `WHERE ma_hang = $2 AND warehouse_id = $3`

**6g. `InsertMovement` (line 138):**
- Add `warehouse_id` to INSERT column list

**6h. `InsertInboundItem` / `InsertOutboundItem`:**
- Add `warehouse_id` to INSERT column list

**6i. Threshold queries (lines 245, 270, 310):**
- `GetThresholds`: add `AND warehouse_id = $N`
- `SaveThreshold` (line 270) expire old: `WHERE ma_hang = $1 AND effective_to IS NULL` becomes `AND warehouse_id = $N`
- `GetActiveThresholds`: add `AND warehouse_id = $N`
- All function signatures: add `warehouseID int64`

**6j. `ListOrders`:** Add `WHERE warehouse_id = $N` filter to inbound/outbound item queries.

**Verify:** `go build ./internal/repo/...`

**Commit:**
```bash
git add internal/repo/orders.go
git commit -m "feat(multi-wh): scope orders/FIFO/thresholds repo to warehouse_id"
```

---

## Task 7: Update `internal/repo/inventory_metrics.go` — Warehouse-Scoped Metrics

**Files:**
- Modify: `internal/repo/inventory_metrics.go`

**Changes (every function affected):**

**7a. `GetAllSKUs` (line 12):**
```sql
SELECT DISTINCT ma_hang FROM inventory_main WHERE warehouse_id = $1
```
Function signature: `GetAllSKUs(ctx, warehouseID int64) ([]string, error)`

**7b. Add `GetAllSKUsAllWarehouses`:**
```go
type SKUWarehouse struct {
    MaHang      string
    WarehouseID int64
}

func (r *PostgresRepo) GetAllSKUsAllWarehouses(ctx context.Context) ([]SKUWarehouse, error) {
    // SELECT DISTINCT ma_hang, warehouse_id FROM inventory_main
}
```

**7c. `RecalcMetricsForSKU`:**
Function signature: `RecalcMetricsForSKU(ctx, maHang string, warehouseID int64) error`

All sub-functions must add `AND warehouse_id = $N`:
- `calcSoNgayTon` (line ~97-98): lot query adds `AND warehouse_id = $2`
- `calcLBBQ` (line ~121): outbound query adds `AND warehouse_id = $2`
- LBBQ monthly calc (line ~132): adds `AND warehouse_id = $N`
- UPDATE inventory_main (line ~63): `WHERE ma_hang = $5` becomes `WHERE ma_hang = $5 AND warehouse_id = $6`
- INSERT into `inventory_lbbq_history` (line ~74): change `ON CONFLICT (ma_hang, month_start)` to `ON CONFLICT (ma_hang, warehouse_id, month_start)`, add `warehouse_id` to INSERT columns

**7d. `RecalcMetricsForSKUs`:**
Function signature: `RecalcMetricsForSKUs(ctx, skus []string, warehouseID int64) error`
Iterates and calls `RecalcMetricsForSKU(ctx, sku, warehouseID)`.

**Verify:** `go build ./internal/repo/...`

**Commit:**
```bash
git add internal/repo/inventory_metrics.go
git commit -m "feat(multi-wh): scope metrics recalc to warehouse_id"
```

---

## Task 8: Update `internal/repo/combo.go` — Warehouse-Scoped Combo Operations

**Files:**
- Modify: `internal/repo/combo.go`

**Master data queries (NO change needed):**
- `ListComboMasters`, `GetComboDetail`, `DeleteComboMaster` — read shared `combo_master` + BOM tables
- `SaveComboMaster` BOM upserts (lines 88-130) for `combo_master`, `combo_bom_semi`, `combo_bom_accessory`

**Changes:**

**8a. `SaveComboMaster` combo_inventory seed (line 100):**
Change from:
```sql
INSERT INTO combo_inventory (ma_combo) VALUES ($1) ON CONFLICT (ma_combo) DO NOTHING
```
To seed for ALL active warehouses:
```sql
INSERT INTO combo_inventory (ma_combo, warehouse_id, so_ton, so_nhap, so_xuat, so_tra)
SELECT $1, id, 0, 0, 0, 0 FROM warehouses WHERE is_active = TRUE
ON CONFLICT (ma_combo, warehouse_id) DO NOTHING
```

**8b. `UpdateComboInventory` (line 176):**
- `WHERE ma_combo = $1` becomes `WHERE ma_combo = $1 AND warehouse_id = $N`
- Function signature: add `warehouseID int64`

**8c. `GetComboInventoryForUpdate` (line 190):**
- `WHERE ma_combo = $1 FOR UPDATE` becomes `WHERE ma_combo = $1 AND warehouse_id = $2 FOR UPDATE`
- Function signature: add `warehouseID int64`

**8d. `GetComboInventory` (list):**
- Add `WHERE warehouse_id = $1`
- Function signature: add `warehouseID int64`

**8e. `ListComboTransactions`:**
- Add `WHERE warehouse_id = $1`
- Function signature: add `warehouseID int64`

**8f. INSERT into `combo_transactions`:** Add `warehouse_id` column.

**8g. INSERT into `combo_component_movements`:** Add `warehouse_id` column.

**Verify:** `go build ./internal/repo/...`

**Commit:**
```bash
git add internal/repo/combo.go
git commit -m "feat(multi-wh): scope combo repo operations to warehouse_id"
```

---

## Task 9: Update `internal/repo/accessory.go` — Warehouse-Scoped Accessory Operations

**Files:**
- Modify: `internal/repo/accessory.go`

**Master data (NO change):**
- `ListAccessories` — reads shared `accessories` table
- `CreateAccessory` INSERT into `accessories` — shared

**Changes:**

**9a. `CreateAccessory` INSERT into `accessory_inventory` (line 40):**
Change to seed ALL active warehouses:
```sql
INSERT INTO accessory_inventory (ma_phu_kien, warehouse_id, so_ton)
SELECT $1, id, 0 FROM warehouses WHERE is_active = TRUE
ON CONFLICT (ma_phu_kien, warehouse_id) DO NOTHING
```

**9b. `UpdateAccessoryStock` (line 73):**
- `WHERE ma_phu_kien = $1` becomes `WHERE ma_phu_kien = $1 AND warehouse_id = $N`
- Function signature: add `warehouseID int64`

**9c. `GetAccessoryStockForUpdate` (line 89):**
- `WHERE ma_phu_kien = $1 FOR UPDATE` becomes `WHERE ma_phu_kien = $1 AND warehouse_id = $2 FOR UPDATE`
- Function signature: add `warehouseID int64`

**9d. `GetAccessoryInventory` (list):**
- Add `WHERE warehouse_id = $1`
- Function signature: add `warehouseID int64`

**9e. INSERT into `accessory_movements`:** Add `warehouse_id` column.

**9f. `AccessoryStockIn`:** Add `warehouse_id` to UPDATE/INSERT. Function signature: add `warehouseID int64`.

**Verify:** `go build ./internal/repo/...`

**Commit:**
```bash
git add internal/repo/accessory.go
git commit -m "feat(multi-wh): scope accessory repo operations to warehouse_id"
```

---

## Task 10: Update `internal/repo/dashboard.go` — Warehouse-Scoped Dashboard

**Files:**
- Modify: `internal/repo/dashboard.go`

**All functions need `warehouseID int64` parameter.**

**10a. `GetSummary` (~line 19):**
All 4 KPI queries: add `WHERE im.warehouse_id = $1` or `AND im.warehouse_id = $1` to inventory_main queries. Threshold JOINs also need `AND it.warehouse_id = $1`.

**10b. `GetCharts` (~line 74):**
- Inbound by week: `WHERE warehouse_id = $1 AND ngay_nhan_hang >= ...`
- Outbound by week: `WHERE warehouse_id = $1 AND ngay_nhan_hang >= ...`
- Inventory vs Optimal: `AND im.warehouse_id = $1 AND it.warehouse_id = $1`

**10c. `GetAlerts` (~line 111):**
- `AND im.warehouse_id = $1 AND it.warehouse_id = $1`

**10d. `GetZeroSales` (~line 194):**
- Subquery: `WHERE warehouse_id = $1` on outbound_items
- Main: `WHERE im.warehouse_id = $1` on inventory_main

**10e. `GetRestockAlerts` (~line 240):**
- Same pattern: `WHERE warehouse_id = $1`

**10f. Threshold queries (lines 278, 320):**
- `AND warehouse_id = $N`

**Verify:** `go build ./internal/repo/...`

**Commit:**
```bash
git add internal/repo/dashboard.go
git commit -m "feat(multi-wh): scope dashboard repo queries to warehouse_id"
```

---

## Task 11: Update `internal/repo/reset.go` — Documentation

**Files:**
- Modify: `internal/repo/reset.go`

**Changes:** Add documentation comment only. No functional change.

```go
// ResetAll truncates all business data tables.
// NOTE: warehouses table is intentionally preserved.
// To remove warehouses, run migration rollback to version 006.
```

**Commit:**
```bash
git add internal/repo/reset.go
git commit -m "docs(multi-wh): document warehouses preservation in ResetAll"
```

---

## Task 12: Update `internal/service/inventory.go` — Pass WarehouseID

**Files:**
- Modify: `internal/service/inventory.go`

**All methods get `warehouseID int64` parameter:**

- `GridQuery(ctx, req, warehouseID)` -> repo call
- `UpdateItem(ctx, maHang, warehouseID, fields)` -> `UpdateInventoryItem` + `RecalcMetricsForSKU`
- `BulkUpdate(ctx, req)` — req.WarehouseID included in queue payload
- `GetFilterOptions(ctx, warehouseID)`
- `ExportRows(ctx, req, warehouseID)`

**Verify:** `go build ./internal/service/...`

**Commit:**
```bash
git add internal/service/inventory.go
git commit -m "feat(multi-wh): pass warehouseID through inventory service"
```

---

## Task 13: Update `internal/service/orders.go` — Pass WarehouseID

**Files:**
- Modify: `internal/service/orders.go`

**Changes:**

- `CreateInbound(ctx, req)` — `req.WarehouseID` passed to: `UpsertLotInbound`, `UpdateInventoryMainInbound`, `InsertMovement`, `RecalcMetricsForSKU`
- `CreateOutbound(ctx, req)` — `req.WarehouseID` passed to: `GetLotsForFIFO`, `DeductLot`, `UpdateInventoryMainOutbound`, `InsertMovement`, `RecalcMetricsForSKU`
- `ListOrders(ctx, filter, warehouseID)`
- `GetLots(ctx, maHang, warehouseID)`
- `GetThresholds(ctx, maHang, warehouseID)`
- `SaveThreshold(ctx, req)` — `req.WarehouseID`

**Verify:** `go build ./internal/service/...`

**Commit:**
```bash
git add internal/service/orders.go
git commit -m "feat(multi-wh): pass warehouseID through order service"
```

---

## Task 14: Update `internal/service/combo.go` — Pass WarehouseID + Fix Inline SQL

**Files:**
- Modify: `internal/service/combo.go`

**CRITICAL: This file has inline SQL that must be updated.**

**14a. Inline SQL fixes:**
- Line 151: `UPDATE inventory_main SET so_ton = so_ton - $1 WHERE ma_hang = $2`
  Add: `AND warehouse_id = $3` and pass `req.WarehouseID`
- Line 259: `UPDATE inventory_main SET so_ton = so_ton + $1 WHERE ma_hang = $2`
  Add: `AND warehouse_id = $3`
- Line 271: `WHERE ma_hang = $1 ORDER BY received_at DESC`
  Add: `AND warehouse_id = $2`

**14b. Service method updates:**
- `CreateCombo(ctx, req)` — pass `req.WarehouseID` to all repo calls
- `CancelCombo(ctx, req)` — same
- `ComboOut(ctx, req)` — same
- `ComboReturn(ctx, req)` — same
- `GetComboInventory(ctx, warehouseID)`
- `ListComboTransactions(ctx, page, limit, warehouseID)`
- `SaveComboMaster(ctx, req)` — seeding now handled by repo

**Verify:** `go build ./internal/service/...`

**Commit:**
```bash
git add internal/service/combo.go
git commit -m "feat(multi-wh): scope combo service to warehouse_id (incl inline SQL)"
```

---

## Task 15: Update `internal/service/dashboard.go` — Pass WarehouseID

**Files:**
- Modify: `internal/service/dashboard.go`

**All methods get `warehouseID int64`:**
- `GetSummary(ctx, warehouseID)`
- `GetCharts(ctx, weeks, warehouseID)`
- `GetAlerts(ctx, warehouseID)`
- `GetZeroSales(ctx, warehouseID)`
- `GetRestockAlerts(ctx, warehouseID)`

**Verify:** `go build ./internal/service/...`

**Commit:**
```bash
git add internal/service/dashboard.go
git commit -m "feat(multi-wh): pass warehouseID through dashboard service"
```

---

## Task 16: Update `internal/service/importer.go` — Pass WarehouseID

**Files:**
- Modify: `internal/service/importer.go`

**Changes:**
- `ProcessImport(ctx, payload)` — payload gains WarehouseID
- All `UpsertInventory` calls: pass `warehouseID`
- All `UpsertLot` calls: pass `warehouseID`
- All `RecalcMetricsForSKUs` calls: pass `warehouseID`

**Verify:** `go build ./internal/service/...`

**Commit:**
```bash
git add internal/service/importer.go
git commit -m "feat(multi-wh): pass warehouseID through importer service"
```

---

## Task 17: Update Handlers + Routes — Extract WarehouseID + Warehouse Endpoints

**Files:**
- Modify: `internal/web/handlers.go`
- Modify: `internal/web/routes.go`

**17a. Add `getWarehouseID` helper:**

```go
func getWarehouseID(c *gin.Context) (int64, error) {
	whStr := c.Query("warehouse_id")
	if whStr == "" {
		return 0, fmt.Errorf("warehouse_id query parameter is required")
	}
	id, err := strconv.ParseInt(whStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid warehouse_id: %s", whStr)
	}
	return id, nil
}
```

**17b. Add Repo field to Handlers struct** (for warehouse CRUD which doesn't go through a service):

Add `Repo *repo.PostgresRepo` field to Handlers struct and constructor.

**17c. Add warehouse handlers:**
- `ListWarehouses(c)` — `GET /api/warehouses`
- `CreateWarehouse(c)` — `POST /api/warehouses` (calls CreateWarehouse + InitWarehouseInventory)
- `UpdateWarehouse(c)` — `PATCH /api/warehouses/:id`
- `GetWarehouse(c)` — `GET /api/warehouses/:id`

**17d. Update ALL stock-related handlers to extract warehouseID:**

Every handler needs:
```go
whID, err := getWarehouseID(c)
if err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
}
```
Then pass `whID` to service calls.

List of handlers to update: `InventoryGrid`, `UpdateInventoryItem`, `BulkUpdateInventory`, `InventoryFilterOptions`, `ExportInventory`, `ImportFile`, `DashboardSummary`, `DashboardCharts`, `ZeroSales`, `RestockAlerts`, `InventoryLots`, `InventoryAlerts`, `ListOrders`, `CreateOrder`, `GetThresholds`, `SaveThreshold`, `RecalcAllMetrics`, all Combo handlers, all Accessory handlers.

**17e. Add warehouse routes:**

```go
api.GET("/warehouses", h.ListWarehouses)
api.POST("/warehouses", h.CreateWarehouse)
api.GET("/warehouses/:id", h.GetWarehouse)
api.PATCH("/warehouses/:id", h.UpdateWarehouse)
```

**Verify:** `go build ./...`

**Commit:**
```bash
git add internal/web/handlers.go internal/web/routes.go
git commit -m "feat(multi-wh): add warehouse endpoints + warehouseID extraction in all handlers"
```

---

## Task 18: Update Worker — WarehouseID in Queue Payloads

**Files:**
- Modify: `cmd/worker/main.go`
- Modify: `internal/queue/redis.go` (if payload structs defined here)

**18a. Update payload structs:**

All payloads gain `WarehouseID int64 `json:"warehouse_id"``:
- Import payload
- BulkUpdate payload
- Recalc payload (0 means all warehouses)

**18b. Update worker consumers:**

- BulkUpdate consumer: `pg.UpdateInventoryItem(ctx, item.MaHang, payload.WarehouseID, item.Fields)` + `pg.RecalcMetricsForSKU(ctx, item.MaHang, payload.WarehouseID)`
- Import consumer: pass `payload.WarehouseID` to importer service
- Recalc consumer:
  - If `payload.WarehouseID > 0`: `pg.GetAllSKUs(ctx, payload.WarehouseID)` then recalc each for that warehouse
  - If `payload.WarehouseID == 0`: `pg.GetAllSKUsAllWarehouses(ctx)` then recalc each (ma_hang, warehouse_id) pair

**Verify:** `go build ./...` (full project should compile now)

**Commit:**
```bash
git add cmd/worker/main.go internal/queue/
git commit -m "feat(multi-wh): add warehouseID to all queue payloads + worker consumers"
```

---

## Task 19: Frontend — TypeScript Types + API Client

**Files:**
- Create: `web/src/types/warehouse.ts`
- Modify: `web/src/api/client.ts`
- Modify: `web/src/types/inventory.ts`
- Modify: `web/src/types/combo.ts`

**19a. Create warehouse types:**

```typescript
export interface Warehouse {
  id: number
  code: string
  name: string
  address: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateWarehouseRequest {
  name: string
  address?: string
}

export interface UpdateWarehouseRequest {
  name?: string
  address?: string
}
```

**19b. Add `warehouse_id: number` to:** `InventoryMain`, `ComboInventory`, `ComboTransaction`, `AccessoryInventory`.

**19c. Update API client:**

Add warehouse CRUD functions. Update ALL existing functions to accept and pass `warehouseId` as query parameter:

```typescript
inventoryGrid: (body: unknown, warehouseId: number) =>
  axios.post(`/api/inventory/grid?warehouse_id=${warehouseId}`, body).then(r => r.data),
// Apply same pattern to all 35 functions
```

**Verify:** `cd web && npx tsc --noEmit`

**Commit:**
```bash
git add web/src/types/ web/src/api/client.ts
git commit -m "feat(multi-wh): add warehouse TS types + warehouseId to all API calls"
```

---

## Task 20: Frontend — Warehouse Store + Selector Component

**Files:**
- Create: `web/src/stores/warehouseStore.ts`
- Create: `web/src/components/WarehouseSelector.tsx`
- Modify: `web/src/components/Sidebar.tsx`

**20a. Zustand store with persist:**

```typescript
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { Warehouse } from '../types/warehouse'

interface WarehouseState {
  warehouses: Warehouse[]
  activeWarehouseId: number | null
  setWarehouses: (warehouses: Warehouse[]) => void
  setActiveWarehouse: (id: number) => void
}

export const useWarehouseStore = create<WarehouseState>()(
  persist(
    (set) => ({
      warehouses: [],
      activeWarehouseId: null,
      setWarehouses: (warehouses) => set({ warehouses }),
      setActiveWarehouse: (id) => set({ activeWarehouseId: id }),
    }),
    { name: 'wms-warehouse' }
  )
)
```

**20b. WarehouseSelector component:** Fetches warehouses on mount, renders `<select>` dropdown.

**20c. Add to Sidebar.tsx:** Render `<WarehouseSelector />` at top of sidebar.

**Verify:** `cd web && npx tsc --noEmit`

**Commit:**
```bash
git add web/src/stores/ web/src/components/WarehouseSelector.tsx web/src/components/Sidebar.tsx
git commit -m "feat(multi-wh): add warehouse Zustand store + selector component"
```

---

## Task 21: Frontend — Update Inventory View + AG Grid Reset

**Files:**
- Modify: `web/src/views/Inventory.tsx`
- Modify: `web/src/components/InventoryGrid.tsx`
- Modify: `web/src/components/ImportPanel.tsx`

**21a. AG Grid datasource:** Pass `activeWarehouseId` to `api.inventoryGrid()`.

**21b. Reset grid on warehouse switch:** Use React key pattern:
```tsx
<InventoryGrid key={activeWarehouseId} warehouseId={activeWarehouseId!} />
```
Or use `gridApi.refreshServerSide({ purge: true })` in a useEffect watching `activeWarehouseId`.

**21c. ImportPanel, export, filter options:** All pass `activeWarehouseId`.

**Verify:** `cd web && npx tsc --noEmit && npm run build`

**Commit:**
```bash
git add web/src/views/Inventory.tsx web/src/components/InventoryGrid.tsx web/src/components/ImportPanel.tsx
git commit -m "feat(multi-wh): scope Inventory view to active warehouse + grid reset on switch"
```

---

## Task 22: Frontend — Update Orders, ComboWarehouse, Overview, Settings Views

**Files:**
- Modify: `web/src/views/Orders.tsx`
- Modify: `web/src/views/ComboWarehouse.tsx`
- Modify: `web/src/views/Overview.tsx`
- Modify: `web/src/components/KpiCards.tsx`
- Modify: `web/src/components/Charts.tsx`
- Modify: `web/src/views/Settings.tsx`

**22a. Overview (Dashboard):**
- KpiCards: pass `activeWarehouseId` to dashboard API calls
- Charts: same
- Re-fetch on warehouse change

**22b. Orders:**
- `ListOrders`: pass `activeWarehouseId`
- `CreateOrder`: include `warehouse_id` in request body

**22c. ComboWarehouse:**
- All combo/accessory API calls pass `activeWarehouseId`

**22d. Settings:**
- Threshold management: pass `activeWarehouseId`
- Add warehouse management section (CRUD)

**Verify:** `cd web && npx tsc --noEmit && npm run build`

**Commit:**
```bash
git add web/src/views/ web/src/components/
git commit -m "feat(multi-wh): scope all views to active warehouse"
```

---

## Task 23: End-to-End Verification

**Dependencies:** ALL previous tasks

**Step 1:** `go build ./...` — No errors

**Step 2:** `make docker-up && make migrate && make migrate-status` — Version 7, clean

**Step 3:** `make test` — All pass

**Step 4:** `cd web && npx tsc --noEmit && npm run build` — No errors

**Step 5: Manual smoke test:**
```
make dev-api   # Terminal 1
make web-dev   # Terminal 2
```

Checklist:
- [ ] Warehouse selector visible in sidebar
- [ ] Default "Kho mac dinh (WH-0001)" selected
- [ ] Inventory grid loads for default warehouse
- [ ] Create second warehouse
- [ ] Switch to second warehouse — grid empty
- [ ] Import Excel for second warehouse — data appears
- [ ] Switch back — original data intact
- [ ] Dashboard KPIs change per warehouse
- [ ] Create inbound for warehouse 2 — stock updates only in warehouse 2
- [ ] Create combo in warehouse 1 — stock deducted only from warehouse 1
- [ ] Recalc-all works without cross-warehouse contamination

**Step 6:** Down migration test: `migrate down 1` — Version 6, clean (data loss for non-default warehouses expected)

---

## Appendix: Files Modified Summary

| File | Type | Changes |
|------|------|---------|
| `migrations/007_multi_warehouse.up.sql` | Create | Full schema migration |
| `migrations/007_multi_warehouse.down.sql` | Create | Rollback (destructive) |
| `internal/domain/entities.go` | Modify | Warehouse struct + WarehouseID |
| `internal/domain/orders.go` | Modify | WarehouseID on requests |
| `internal/domain/combo.go` | Modify | WarehouseID on combo types |
| `internal/domain/grid.go` | Modify | WarehouseID on BulkUpdateRequest |
| `internal/domain/dashboard.go` | Modify | WarehouseID on ThresholdRequest |
| `internal/repo/warehouse.go` | Create | CRUD + InitWarehouseInventory |
| `internal/repo/inventory.go` | Modify | warehouse_id in all queries |
| `internal/repo/import.go` | Modify | warehouse_id in all upserts |
| `internal/repo/orders.go` | Modify | warehouse_id in FIFO/movements/thresholds |
| `internal/repo/inventory_metrics.go` | Modify | warehouse-scoped metrics |
| `internal/repo/combo.go` | Modify | warehouse_id in combo stock ops |
| `internal/repo/accessory.go` | Modify | warehouse_id in accessory ops |
| `internal/repo/dashboard.go` | Modify | warehouse_id in all aggregations |
| `internal/repo/reset.go` | Modify | Documentation only |
| `internal/service/inventory.go` | Modify | Pass warehouseID |
| `internal/service/orders.go` | Modify | Pass warehouseID |
| `internal/service/combo.go` | Modify | Pass warehouseID + fix inline SQL |
| `internal/service/dashboard.go` | Modify | Pass warehouseID |
| `internal/service/importer.go` | Modify | Pass warehouseID |
| `internal/web/handlers.go` | Modify | getWarehouseID + all handlers |
| `internal/web/routes.go` | Modify | Warehouse CRUD routes |
| `cmd/worker/main.go` | Modify | WarehouseID in payloads |
| `internal/queue/redis.go` | Modify | Payload structs |
| `web/src/types/warehouse.ts` | Create | Warehouse TS types |
| `web/src/types/inventory.ts` | Modify | Add warehouse_id |
| `web/src/types/combo.ts` | Modify | Add warehouse_id |
| `web/src/api/client.ts` | Modify | All functions + warehouse CRUD |
| `web/src/stores/warehouseStore.ts` | Create | Zustand store |
| `web/src/components/WarehouseSelector.tsx` | Create | Dropdown component |
| `web/src/components/Sidebar.tsx` | Modify | Add selector |
| `web/src/components/InventoryGrid.tsx` | Modify | Warehouse-scoped + purge |
| `web/src/components/ImportPanel.tsx` | Modify | Pass warehouseId |
| `web/src/components/KpiCards.tsx` | Modify | Pass warehouseId |
| `web/src/components/Charts.tsx` | Modify | Pass warehouseId |
| `web/src/views/Inventory.tsx` | Modify | Pass warehouseId |
| `web/src/views/Orders.tsx` | Modify | Pass warehouseId |
| `web/src/views/ComboWarehouse.tsx` | Modify | Pass warehouseId |
| `web/src/views/Overview.tsx` | Modify | Pass warehouseId |
| `web/src/views/Settings.tsx` | Modify | Warehouse mgmt + warehouseId |
| `internal/grid/builder.go` | Modify | Base filter for warehouse_id |

**Total: 41 files (6 create, 35 modify)**
