package repo

import (
	"context"
	"fmt"
	"time"

	"wms-v1/internal/domain"
	"wms-v1/internal/importer"
)

func (r *PostgresRepo) CreateImportBatch(ctx context.Context, fileType, fileName string, totalRows int) (int64, error) {
	var id int64
	err := r.Pool.QueryRow(ctx,
		`INSERT INTO import_batches (file_type, file_name, total_rows, status)
		 VALUES ($1, $2, $3, 'processing')
		 RETURNING batch_id`,
		fileType, fileName, totalRows).Scan(&id)
	return id, err
}

func (r *PostgresRepo) UpdateImportBatch(ctx context.Context, batchID int64, success, errors int, status string, errJSON string) error {
	now := time.Now()
	_, err := r.Pool.Exec(ctx,
		`UPDATE import_batches SET success_rows=$1, error_rows=$2, status=$3, errors=$4, completed_at=$5 WHERE batch_id=$6`,
		success, errors, status, errJSON, now, batchID)
	return err
}

// UpsertProducts bulk upserts products
func (r *PostgresRepo) UpsertProducts(ctx context.Context, products []domain.Product) (int, error) {
	success := 0
	for _, p := range products {
		_, err := r.Pool.Exec(ctx,
			`INSERT INTO products (ma_hang, ten_san_pham, ma_bu, ma_cat, ma_nhom_hang, nhom_hang,
			  don_vi_tinh, quy_cach, don_gia, vat, gia_niv, gia_nhap, ngay_cap_nhat, hoa_hong)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
			 ON CONFLICT (ma_hang) DO UPDATE SET
			  ten_san_pham=EXCLUDED.ten_san_pham, ma_bu=EXCLUDED.ma_bu, ma_cat=EXCLUDED.ma_cat,
			  ma_nhom_hang=EXCLUDED.ma_nhom_hang, nhom_hang=EXCLUDED.nhom_hang,
			  don_vi_tinh=EXCLUDED.don_vi_tinh, quy_cach=EXCLUDED.quy_cach,
			  don_gia=EXCLUDED.don_gia, vat=EXCLUDED.vat, gia_niv=EXCLUDED.gia_niv,
			  gia_nhap=EXCLUDED.gia_nhap, ngay_cap_nhat=EXCLUDED.ngay_cap_nhat,
			  hoa_hong=EXCLUDED.hoa_hong`,
			p.MaHang, p.TenSanPham, p.MaBu, p.MaCat, p.MaNhomHang, p.NhomHang,
			p.DonViTinh, p.QuyCach, p.DonGia, p.Vat, p.GiaNiv, p.GiaNhap,
			p.NgayCapNhat, p.HoaHong)
		if err != nil {
			continue
		}
		success++
	}
	return success, nil
}

// UpsertInventory bulk upserts inventory
func (r *PostgresRepo) UpsertInventory(ctx context.Context, items []domain.InventoryMain) (int, error) {
	success := 0
	for _, item := range items {
		_, err := r.Pool.Exec(ctx,
			`INSERT INTO inventory_main (ma_hang, ten_san_pham, so_ton, so_nhap, so_xuat,
			  tien_ton, tien_nhap, tien_xuat, so_ngay_ton, luong_ban_binh_quan_ngay)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			 ON CONFLICT (ma_hang) DO UPDATE SET
			  ten_san_pham=EXCLUDED.ten_san_pham, so_ton=EXCLUDED.so_ton,
			  so_nhap=EXCLUDED.so_nhap, so_xuat=EXCLUDED.so_xuat,
			  tien_ton=EXCLUDED.tien_ton, tien_nhap=EXCLUDED.tien_nhap,
			  tien_xuat=EXCLUDED.tien_xuat, so_ngay_ton=EXCLUDED.so_ngay_ton,
			  luong_ban_binh_quan_ngay=EXCLUDED.luong_ban_binh_quan_ngay`,
			item.MaHang, item.TenSanPham, item.SoTon, item.SoNhap, item.SoXuat,
			item.TienTon, item.TienNhap, item.TienXuat, item.SoNgayTon,
			item.LuongBanBinhQuanNgay)
		if err != nil {
			continue
		}
		success++
	}
	return success, nil
}

// InsertInboundItems bulk inserts inbound items
func (r *PostgresRepo) InsertInboundItems(ctx context.Context, items []domain.InboundItem) (int, error) {
	success := 0
	for _, item := range items {
		_, err := r.Pool.Exec(ctx,
			`INSERT INTO inbound_items (ma_hang, ten_san_pham, don_vi_tinh, quy_cach,
			  so_luong, doanh_so, chiet_khau, so_luong_tra_lai, doanh_thu, von, lai_gop, ti_le_lai_gop, ngay_nhan_hang)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			item.MaHang, item.TenSanPham, item.DonViTinh, item.QuyCach,
			item.SoLuong, item.DoanhSo, item.ChietKhau, item.SoLuongTraLai,
			item.DoanhThu, item.Von, item.LaiGop, item.TiLeLaiGop, item.NgayNhanHang)
		if err != nil {
			continue
		}
		success++
	}
	return success, nil
}

// InsertOutboundItems bulk inserts outbound items
func (r *PostgresRepo) InsertOutboundItems(ctx context.Context, items []domain.InboundItem) (int, error) {
	success := 0
	for _, item := range items {
		_, err := r.Pool.Exec(ctx,
			`INSERT INTO outbound_items (ma_hang, ten_san_pham, don_vi_tinh, quy_cach,
			  so_luong, doanh_so, chiet_khau, so_luong_tra_lai, doanh_thu, von, lai_gop, ti_le_lai_gop, ngay_nhan_hang)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			item.MaHang, item.TenSanPham, item.DonViTinh, item.QuyCach,
			item.SoLuong, item.DoanhSo, item.ChietKhau, item.SoLuongTraLai,
			item.DoanhThu, item.Von, item.LaiGop, item.TiLeLaiGop, item.NgayNhanHang)
		if err != nil {
			continue
		}
		success++
	}
	return success, nil
}

// CreateAsyncJob creates a new async job
func (r *PostgresRepo) CreateAsyncJob(ctx context.Context, jobID, jobType, payload string) error {
	_, err := r.Pool.Exec(ctx,
		"INSERT INTO async_jobs (job_id, job_type, payload) VALUES ($1, $2, $3)",
		jobID, jobType, payload)
	return err
}

// ImportInventoryFull processes rows from the 15-column file:
// 1. Upsert products (with computed gia_nhap, gia_niv)
// 2. Upsert inventory_main (so_ton + tien_ton only; so_nhap/so_xuat NOT touched)
// 3. Insert inbound_items (with batch_code)
// 4. Upsert inventory_lots
// Best-effort: each row independent, skip on error.
func (r *PostgresRepo) ImportInventoryFull(ctx context.Context, rows []importer.InventoryFullRow) (int, error) {
	success := 0
	for _, row := range rows {
		// 1. Upsert product (ngay_cap_nhat uses value from parser directly)
		_, err := r.Pool.Exec(ctx,
			`INSERT INTO products (ma_hang, ten_san_pham, ma_bu, ma_cat, ma_nhom_hang, nhom_hang,
			  don_vi_tinh, quy_cach, don_gia, vat, gia_niv, gia_nhap, ngay_cap_nhat, hoa_hong)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
			 ON CONFLICT (ma_hang) DO UPDATE SET
			  ten_san_pham=EXCLUDED.ten_san_pham, ma_bu=EXCLUDED.ma_bu, ma_cat=EXCLUDED.ma_cat,
			  ma_nhom_hang=EXCLUDED.ma_nhom_hang, nhom_hang=EXCLUDED.nhom_hang,
			  don_vi_tinh=EXCLUDED.don_vi_tinh, quy_cach=EXCLUDED.quy_cach,
			  don_gia=EXCLUDED.don_gia, vat=EXCLUDED.vat, gia_niv=EXCLUDED.gia_niv,
			  gia_nhap=EXCLUDED.gia_nhap,
			  ngay_cap_nhat=EXCLUDED.ngay_cap_nhat,
			  hoa_hong=EXCLUDED.hoa_hong`,
			row.Product.MaHang, row.Product.TenSanPham, row.Product.MaBu, row.Product.MaCat,
			row.Product.MaNhomHang, row.Product.NhomHang, row.Product.DonViTinh, row.Product.QuyCach,
			row.Product.DonGia, row.Product.Vat, row.Product.GiaNiv, row.Product.GiaNhap,
			row.Product.NgayCapNhat, row.Product.HoaHong)
		if err != nil {
			continue
		}

		// 2. Upsert inventory_main (so_nhap/so_xuat not touched — computed from orders)
		_, err = r.Pool.Exec(ctx,
			`INSERT INTO inventory_main (ma_hang, ten_san_pham, so_ton, so_nhap, so_xuat,
			  tien_ton, tien_nhap, tien_xuat)
			 VALUES ($1,$2,$3,0,0,$4,0,0)
			 ON CONFLICT (ma_hang) DO UPDATE SET
			  ten_san_pham=EXCLUDED.ten_san_pham, so_ton=EXCLUDED.so_ton,
			  tien_ton=EXCLUDED.tien_ton`,
			row.Inventory.MaHang, row.Inventory.TenSanPham, row.Inventory.SoTon,
			row.Inventory.TienTon)
		if err != nil {
			continue
		}

		// 3. Insert inbound_items
		_, err = r.Pool.Exec(ctx,
			`INSERT INTO inbound_items (ma_hang, ten_san_pham, don_vi_tinh, quy_cach,
			  so_luong, doanh_so, chiet_khau, so_luong_tra_lai, doanh_thu, von,
			  lai_gop, ti_le_lai_gop, ngay_nhan_hang, batch_code)
			 VALUES ($1,$2,$3,$4,$5,0,0,0,0,0,0,0,$6,$7)`,
			row.Inbound.MaHang, row.Inbound.TenSanPham, row.Inbound.DonViTinh,
			row.Inbound.QuyCach, row.Inbound.SoLuong,
			row.Inbound.NgayNhanHang, row.Inbound.BatchCode)
		if err != nil {
			continue
		}

		// 4. Upsert inventory_lots
		_, err = r.Pool.Exec(ctx,
			`INSERT INTO inventory_lots (ma_hang, batch_code, received_at, qty_in, qty_out, qty_remaining)
			 VALUES ($1, $2, $3, $4, 0, $4)
			 ON CONFLICT (ma_hang, batch_code) DO UPDATE SET
			  qty_in = inventory_lots.qty_in + EXCLUDED.qty_in,
			  qty_remaining = inventory_lots.qty_remaining + EXCLUDED.qty_in`,
			row.Inbound.MaHang, row.Inbound.BatchCode, row.Inbound.NgayNhanHang, row.Inbound.SoLuong)
		if err != nil {
			continue
		}

		success++
	}
	return success, nil
}

// GetImportBatch retrieves import batch by ID
func (r *PostgresRepo) GetImportBatch(ctx context.Context, batchID int64) (*domain.ImportBatch, error) {
	var b domain.ImportBatch
	err := r.Pool.QueryRow(ctx,
		`SELECT batch_id, file_type, file_name, total_rows, success_rows, error_rows, status, errors, created_at, completed_at
		 FROM import_batches WHERE batch_id=$1`, batchID).Scan(
		&b.BatchID, &b.FileType, &b.FileName, &b.TotalRows, &b.SuccessRows,
		&b.ErrorRows, &b.Status, &b.Errors, &b.CreatedAt, &b.CompletedAt)
	if err != nil {
		return nil, fmt.Errorf("get import batch %d: %w", batchID, err)
	}
	return &b, nil
}

// GetLatestImportBatch retrieves the most recent import batch
func (r *PostgresRepo) GetLatestImportBatch(ctx context.Context) (*domain.ImportBatch, error) {
	var b domain.ImportBatch
	err := r.Pool.QueryRow(ctx,
		`SELECT batch_id, file_type, file_name, total_rows, success_rows, error_rows, status, errors, created_at, completed_at
		 FROM import_batches ORDER BY batch_id DESC LIMIT 1`).Scan(
		&b.BatchID, &b.FileType, &b.FileName, &b.TotalRows, &b.SuccessRows,
		&b.ErrorRows, &b.Status, &b.Errors, &b.CreatedAt, &b.CompletedAt)
	if err != nil {
		return nil, fmt.Errorf("get latest import batch: %w", err)
	}
	return &b, nil
}

// UpdateAsyncJob updates job status
func (r *PostgresRepo) UpdateAsyncJob(ctx context.Context, jobID, status, result, errMsg string) error {
	_, err := r.Pool.Exec(ctx,
		"UPDATE async_jobs SET status=$1, result=$2, error=$3, updated_at=NOW() WHERE job_id=$4",
		status, result, errMsg, jobID)
	return err
}

// GetAsyncJob retrieves a job by ID
func (r *PostgresRepo) GetAsyncJob(ctx context.Context, jobID string) (*domain.AsyncJob, error) {
	var job domain.AsyncJob
	err := r.Pool.QueryRow(ctx,
		"SELECT job_id, job_type, status, payload, result, error, created_at, updated_at FROM async_jobs WHERE job_id=$1",
		jobID).Scan(&job.JobID, &job.JobType, &job.Status, &job.Payload, &job.Result, &job.Error, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get job %s: %w", jobID, err)
	}
	return &job, nil
}
