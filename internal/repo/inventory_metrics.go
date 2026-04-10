package repo

import (
	"context"
	"fmt"
	"math"
	"time"
)

type SKUWarehouse struct {
	MaHang      string
	WarehouseID int64
}

// GetAllSKUs returns all ma_hang from inventory_main for a warehouse.
func (r *PostgresRepo) GetAllSKUs(ctx context.Context, warehouseID int64) ([]string, error) {
	rows, err := r.Pool.Query(ctx, "SELECT DISTINCT ma_hang FROM inventory_main WHERE warehouse_id = $1", warehouseID)
	if err != nil {
		return nil, fmt.Errorf("get all SKUs: %w", err)
	}
	defer rows.Close()

	var skus []string
	for rows.Next() {
		var sku string
		if err := rows.Scan(&sku); err != nil {
			return nil, fmt.Errorf("scan sku: %w", err)
		}
		skus = append(skus, sku)
	}
	return skus, rows.Err()
}

func (r *PostgresRepo) GetAllSKUsAllWarehouses(ctx context.Context) ([]SKUWarehouse, error) {
	rows, err := r.Pool.Query(ctx, `SELECT DISTINCT ma_hang, warehouse_id FROM inventory_main`)
	if err != nil {
		return nil, fmt.Errorf("get all SKUs all warehouses: %w", err)
	}
	defer rows.Close()

	var result []SKUWarehouse
	for rows.Next() {
		var sw SKUWarehouse
		if err := rows.Scan(&sw.MaHang, &sw.WarehouseID); err != nil {
			return nil, fmt.Errorf("scan SKUWarehouse: %w", err)
		}
		result = append(result, sw)
	}
	return result, rows.Err()
}

// RecalcMetricsForSKU recalculates so_ngay_ton, luong_ban_binh_quan_ngay (nullable),
// so_ngay_ton_ban, and tien_ton/tien_nhap/tien_xuat for a single SKU.
func (r *PostgresRepo) RecalcMetricsForSKU(ctx context.Context, maHang string, warehouseID int64) error {
	soNgayTon := calcSoNgayTon(ctx, r, maHang, warehouseID)
	lbbq := calcLBBQ(ctx, r, maHang, warehouseID)

	// so_ngay_ton_ban = so_ton / LBBQ  (NULL if LBBQ is NULL)
	var soNgayTonBan *float64
	if lbbq != nil && *lbbq > 0 {
		// Fetch current so_ton
		var soTon float64
		err := r.Pool.QueryRow(ctx,
			`SELECT COALESCE(so_ton, 0) FROM inventory_main WHERE ma_hang = $1 AND warehouse_id = $2`,
			maHang, warehouseID).Scan(&soTon)
		if err == nil && soTon > 0 {
			v := math.Round(soTon / *lbbq * 100) / 100
			soNgayTonBan = &v
		}
	}

	// Compute tien_ton/tien_nhap/tien_xuat from products.don_gia
	var donGia float64
	_ = r.Pool.QueryRow(ctx,
		`SELECT COALESCE(don_gia, 0) FROM products WHERE ma_hang = $1`,
		maHang).Scan(&donGia)

	_, err := r.Pool.Exec(ctx,
		`UPDATE inventory_main
		 SET so_ngay_ton = $1,
		     luong_ban_binh_quan_ngay = $2,
		     so_ngay_ton_ban = $3,
		     tien_ton = so_ton * $4,
		     tien_nhap = so_nhap * $4,
		     tien_xuat = so_xuat * $4
		 WHERE ma_hang = $5 AND warehouse_id = $6`,
		soNgayTon, lbbq, soNgayTonBan, donGia, maHang, warehouseID)
	if err != nil {
		return fmt.Errorf("update metrics for %s: %w", maHang, err)
	}

	// Write LBBQ history (first recalc of each month, ON CONFLICT DO NOTHING)
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	_, _ = r.Pool.Exec(ctx,
		`INSERT INTO inventory_lbbq_history (ma_hang, warehouse_id, month_start, lbbq)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (ma_hang, warehouse_id, month_start) DO NOTHING`,
		maHang, warehouseID, monthStart, lbbq)

	return nil
}

// RecalcMetricsForSKUs recalculates metrics for multiple SKUs (best-effort).
func (r *PostgresRepo) RecalcMetricsForSKUs(ctx context.Context, maHangs []string, warehouseID int64) error {
	for _, mh := range maHangs {
		if err := r.RecalcMetricsForSKU(ctx, mh, warehouseID); err != nil {
			// best-effort: log but continue
			fmt.Printf("WARN: recalc metrics for %s failed: %v\n", mh, err)
		}
	}
	return nil
}

// calcSoNgayTon returns the age in days of the oldest remaining lot.
// so_ngay_ton = NOW() - MIN(inventory_lots.received_at) WHERE qty_remaining > 0
// Returns 0 if no lots with remaining stock.
func calcSoNgayTon(ctx context.Context, r *PostgresRepo, maHang string, warehouseID int64) float64 {
	var receivedAt *time.Time
	err := r.Pool.QueryRow(ctx,
		`SELECT MIN(received_at) FROM inventory_lots
		 WHERE ma_hang = $1 AND warehouse_id = $2 AND qty_remaining > 0`,
		maHang, warehouseID).Scan(&receivedAt)
	if err != nil || receivedAt == nil {
		return 0
	}

	days := time.Since(*receivedAt).Hours() / 24
	if days < 0 {
		return 0
	}
	return math.Round(days*100) / 100
}

// calcLBBQ returns the average daily outbound quantity for the latest outbound month.
// LBBQ = total outbound in latest outbound month / number of days in that month
// Returns nil if no outbound data (stored as NULL in DB).
// Rounds to 2 decimal places.
func calcLBBQ(ctx context.Context, r *PostgresRepo, maHang string, warehouseID int64) *float64 {
	// Find the latest outbound month for this SKU
	var latestMonth *time.Time
	err := r.Pool.QueryRow(ctx,
		`SELECT date_trunc('month', MAX(ngay_nhan_hang))
		 FROM outbound_items
		 WHERE ma_hang = $1 AND warehouse_id = $2`,
		maHang, warehouseID).Scan(&latestMonth)
	if err != nil || latestMonth == nil {
		return nil
	}

	// Sum outbound qty in that month
	var totalQty float64
	err = r.Pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(so_luong), 0)
		 FROM outbound_items
		 WHERE ma_hang = $1
		   AND warehouse_id = $2
		   AND ngay_nhan_hang >= $3
		   AND ngay_nhan_hang < $3 + INTERVAL '1 month'`,
		maHang, warehouseID, latestMonth).Scan(&totalQty)
	if err != nil || totalQty == 0 {
		return nil
	}

	// Days in that month
	nextMonth := latestMonth.AddDate(0, 1, 0)
	daysInMonth := nextMonth.Sub(*latestMonth).Hours() / 24

	if daysInMonth <= 0 {
		return nil
	}

	lbbq := math.Round(totalQty/daysInMonth*100) / 100
	return &lbbq
}
