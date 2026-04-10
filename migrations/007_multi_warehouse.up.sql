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
