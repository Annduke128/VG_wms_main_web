-- 003_pricing_metrics.down.sql
DROP TABLE IF EXISTS inventory_lbbq_history;

ALTER TABLE inventory_main DROP COLUMN IF EXISTS so_ngay_ton_ban;

UPDATE inventory_main
  SET luong_ban_binh_quan_ngay = 0
  WHERE luong_ban_binh_quan_ngay IS NULL;
ALTER TABLE inventory_main
  ALTER COLUMN luong_ban_binh_quan_ngay SET DEFAULT 0;
ALTER TABLE inventory_main
  ALTER COLUMN luong_ban_binh_quan_ngay SET NOT NULL;

ALTER TABLE products
  ALTER COLUMN quy_cach TYPE TEXT
  USING quy_cach::TEXT;
