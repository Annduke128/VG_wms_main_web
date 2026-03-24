-- 003_pricing_metrics.up.sql
-- 1. products.quy_cach TEXT → NUMERIC(15,4)
--    Drop default first (text default can't auto-cast to numeric)
ALTER TABLE products
  ALTER COLUMN quy_cach DROP DEFAULT;

ALTER TABLE products
  ALTER COLUMN quy_cach TYPE NUMERIC(15,4)
  USING (
    CASE
      WHEN quy_cach ~ '^[0-9]+\.?[0-9]*$' THEN quy_cach::NUMERIC(15,4)
      ELSE 1
    END
  );

ALTER TABLE products
  ALTER COLUMN quy_cach SET DEFAULT 1;

-- 2. inventory_main.luong_ban_binh_quan_ngay → nullable
ALTER TABLE inventory_main
  ALTER COLUMN luong_ban_binh_quan_ngay DROP NOT NULL;
ALTER TABLE inventory_main
  ALTER COLUMN luong_ban_binh_quan_ngay DROP DEFAULT;
UPDATE inventory_main
  SET luong_ban_binh_quan_ngay = NULL
  WHERE luong_ban_binh_quan_ngay = 0;

-- 3. Add so_ngay_ton_ban (nullable, computed = so_ton / LBBQ)
ALTER TABLE inventory_main
  ADD COLUMN IF NOT EXISTS so_ngay_ton_ban NUMERIC(10,2);

-- 4. LBBQ history table (one row per SKU per month)
CREATE TABLE IF NOT EXISTS inventory_lbbq_history (
  id          BIGSERIAL PRIMARY KEY,
  ma_hang     TEXT NOT NULL REFERENCES products(ma_hang),
  month_start DATE NOT NULL,
  lbbq        NUMERIC(15,2),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (ma_hang, month_start)
);
