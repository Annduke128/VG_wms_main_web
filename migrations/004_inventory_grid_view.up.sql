-- Create a view that joins inventory_main with products
-- to expose don_gia, ma_bu, ma_nhom_hang for grid display and filtering
CREATE OR REPLACE VIEW inventory_grid AS
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
