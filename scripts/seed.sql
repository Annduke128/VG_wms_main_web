-- ============================================
-- WMS v1 — Sample Seed Data
-- Usage: psql $POSTGRES_DSN -f scripts/seed.sql
-- ============================================

BEGIN;

-- Rule config (defaults)
INSERT INTO rule_config (key, value) VALUES
  ('optimal_days', '14'),
  ('gap_ratio', '0.10')
ON CONFLICT (key) DO NOTHING;

-- Sample products
INSERT INTO products (ma_hang, ten_san_pham, ma_bu, ma_cat, ma_nhom_hang, nhom_hang, don_vi_tinh, quy_cach, don_gia, vat, gia_niv, gia_nhap, hoa_hong) VALUES
  ('SP001', 'Sữa tươi Vinamilk 1L',       'BU01', 'CAT01', 'NH01', 'Sữa',       'Thùng', '12 hộp/thùng', 180000, 10, 165000, 160000, 5),
  ('SP002', 'Mì Hảo Hảo tôm chua cay',    'BU01', 'CAT02', 'NH02', 'Mì gói',    'Thùng', '30 gói/thùng',  95000,  8,  88000,  85000, 3),
  ('SP003', 'Nước mắm Chinsu 500ml',       'BU02', 'CAT03', 'NH03', 'Gia vị',    'Thùng', '24 chai/thùng',220000, 10, 200000, 195000, 4),
  ('SP004', 'Dầu ăn Tường An 1L',          'BU02', 'CAT03', 'NH03', 'Gia vị',    'Thùng', '12 chai/thùng',320000, 10, 295000, 290000, 3),
  ('SP005', 'Bia Tiger lon 330ml',          'BU03', 'CAT04', 'NH04', 'Đồ uống',   'Thùng', '24 lon/thùng', 350000, 10, 320000, 315000, 6),
  ('SP006', 'Coca Cola lon 330ml',          'BU03', 'CAT04', 'NH04', 'Đồ uống',   'Thùng', '24 lon/thùng', 220000, 10, 200000, 195000, 5),
  ('SP007', 'Bột giặt OMO 4.5kg',          'BU04', 'CAT05', 'NH05', 'Hóa phẩm',  'Bao',   '1 bao',        165000,  8, 155000, 150000, 4),
  ('SP008', 'Giấy vệ sinh Pulppy 12 cuộn', 'BU04', 'CAT05', 'NH05', 'Hóa phẩm',  'Lốc',   '12 cuộn/lốc',   72000,  8,  66000,  64000, 3),
  ('SP009', 'Gạo ST25 5kg',                'BU05', 'CAT06', 'NH06', 'Lương thực', 'Bao',   '5kg/bao',      125000,  5, 118000, 115000, 2),
  ('SP010', 'Đường Biên Hòa 1kg',          'BU05', 'CAT06', 'NH06', 'Lương thực', 'Bao',   '1kg/bao',       22000,  5,  20000,  19000, 2)
ON CONFLICT (ma_hang) DO NOTHING;

-- Sample inventory
INSERT INTO inventory_main (ma_hang, ten_san_pham, so_ton, so_nhap, so_xuat, tien_ton, tien_nhap, tien_xuat, so_ngay_ton, luong_ban_binh_quan_ngay) VALUES
  ('SP001', 'Sữa tươi Vinamilk 1L',       150, 500, 350, 27000000,  90000000, 63000000,  7, 50),
  ('SP002', 'Mì Hảo Hảo tôm chua cay',    300, 800, 500, 28500000,  76000000, 47500000, 10, 30),
  ('SP003', 'Nước mắm Chinsu 500ml',        80, 200, 120, 17600000,  44000000, 26400000, 14, 10),
  ('SP004', 'Dầu ăn Tường An 1L',           45, 100,  55, 14400000,  32000000, 17600000, 21,  5),
  ('SP005', 'Bia Tiger lon 330ml',          200, 600, 400, 70000000, 210000000,140000000,  5, 80),
  ('SP006', 'Coca Cola lon 330ml',          120, 350, 230, 26400000,  77000000, 50600000,  8, 40),
  ('SP007', 'Bột giặt OMO 4.5kg',           60, 150,  90,  9900000,  24750000, 14850000, 18,  7),
  ('SP008', 'Giấy vệ sinh Pulppy 12 cuộn',  90, 200, 110,  6480000,  14400000,  7920000, 12, 15),
  ('SP009', 'Gạo ST25 5kg',                100, 300, 200, 12500000,  37500000, 25000000,  9, 25),
  ('SP010', 'Đường Biên Hòa 1kg',          250, 500, 250,  5500000,  11000000,  5500000, 30,  8)
ON CONFLICT (ma_hang) DO NOTHING;

-- Sample inbound items
INSERT INTO inbound_items (ma_hang, ten_san_pham, don_vi_tinh, quy_cach, so_luong, doanh_so, chiet_khau, so_luong_tra_lai, doanh_thu, von, lai_gop, ti_le_lai_gop, ngay_nhan_hang) VALUES
  ('SP001', 'Sữa tươi Vinamilk 1L',    'Thùng', '12 hộp/thùng', 100, 18000000, 500000, 0, 17500000, 16000000, 1500000, 8.57,  '2026-03-01'),
  ('SP005', 'Bia Tiger lon 330ml',       'Thùng', '24 lon/thùng',  200, 70000000, 2000000, 5, 68000000, 63000000, 5000000, 7.35,  '2026-03-02'),
  ('SP002', 'Mì Hảo Hảo tôm chua cay', 'Thùng', '30 gói/thùng', 150, 14250000, 300000, 0, 13950000, 12750000, 1200000, 8.60,  '2026-03-03'),
  ('SP009', 'Gạo ST25 5kg',             'Bao',   '5kg/bao',       100, 12500000, 200000, 2, 12300000, 11500000,  800000, 6.50,  '2026-03-05');

-- Sample outbound items
INSERT INTO outbound_items (ma_hang, ten_san_pham, don_vi_tinh, quy_cach, so_luong, doanh_so, chiet_khau, so_luong_tra_lai, doanh_thu, von, lai_gop, ti_le_lai_gop, ngay_nhan_hang) VALUES
  ('SP001', 'Sữa tươi Vinamilk 1L',    'Thùng', '12 hộp/thùng',  50, 9000000, 200000, 0, 8800000, 8000000,  800000, 9.09,  '2026-03-02'),
  ('SP005', 'Bia Tiger lon 330ml',       'Thùng', '24 lon/thùng',  80, 28000000, 800000, 2, 27200000, 25200000, 2000000, 7.35, '2026-03-03'),
  ('SP006', 'Coca Cola lon 330ml',       'Thùng', '24 lon/thùng',  60, 13200000, 300000, 0, 12900000, 11700000, 1200000, 9.30, '2026-03-04');

-- Sample inventory movements
INSERT INTO inventory_movements (ma_hang, qty, type, created_at) VALUES
  ('SP001', 100, 'IN',  '2026-03-01 08:00:00'),
  ('SP005', 200, 'IN',  '2026-03-02 09:30:00'),
  ('SP001',  50, 'OUT', '2026-03-02 14:00:00'),
  ('SP002', 150, 'IN',  '2026-03-03 07:00:00'),
  ('SP005',  80, 'OUT', '2026-03-03 16:00:00'),
  ('SP006',  60, 'OUT', '2026-03-04 10:00:00'),
  ('SP009', 100, 'IN',  '2026-03-05 08:30:00');

COMMIT;
