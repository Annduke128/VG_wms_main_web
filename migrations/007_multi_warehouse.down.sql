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
