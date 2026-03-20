package repo

import (
	"context"
	"fmt"
	"time"
)

// RecalcMetricsForSKU recalculates so_ngay_ton and luong_ban_binh_quan_ngay
// for a single SKU and updates inventory_main.
func (r *PostgresRepo) RecalcMetricsForSKU(ctx context.Context, maHang string) error {
	soNgayTon := calcSoNgayTon(ctx, r, maHang)
	lbbq := calcLBBQ(ctx, r, maHang)

	_, err := r.Pool.Exec(ctx,
		`UPDATE inventory_main
		 SET so_ngay_ton = $1, luong_ban_binh_quan_ngay = $2
		 WHERE ma_hang = $3`,
		soNgayTon, lbbq, maHang)
	if err != nil {
		return fmt.Errorf("update metrics for %s: %w", maHang, err)
	}
	return nil
}

// RecalcMetricsForSKUs recalculates metrics for multiple SKUs (best-effort).
func (r *PostgresRepo) RecalcMetricsForSKUs(ctx context.Context, maHangs []string) error {
	for _, mh := range maHangs {
		if err := r.RecalcMetricsForSKU(ctx, mh); err != nil {
			// best-effort: log but continue
			fmt.Printf("WARN: recalc metrics for %s failed: %v\n", mh, err)
		}
	}
	return nil
}

// calcSoNgayTon returns the age in days of the oldest remaining lot.
// so_ngay_ton = NOW() - MIN(inventory_lots.received_at) WHERE qty_remaining > 0
// Returns 0 if no lots with remaining stock.
func calcSoNgayTon(ctx context.Context, r *PostgresRepo, maHang string) float64 {
	var receivedAt *time.Time
	err := r.Pool.QueryRow(ctx,
		`SELECT MIN(received_at) FROM inventory_lots
		 WHERE ma_hang = $1 AND qty_remaining > 0`,
		maHang).Scan(&receivedAt)
	if err != nil || receivedAt == nil {
		return 0
	}

	days := time.Since(*receivedAt).Hours() / 24
	if days < 0 {
		return 0
	}
	// Round to 2 decimal places
	return float64(int(days*100)) / 100
}

// calcLBBQ returns the average daily outbound quantity for the latest outbound month.
// LBBQ = total outbound in latest outbound month / number of days in that month
// Returns 0 if no outbound data.
func calcLBBQ(ctx context.Context, r *PostgresRepo, maHang string) float64 {
	// Find the latest outbound month for this SKU
	var latestMonth *time.Time
	err := r.Pool.QueryRow(ctx,
		`SELECT date_trunc('month', MAX(ngay_nhan_hang))
		 FROM outbound_items
		 WHERE ma_hang = $1`,
		maHang).Scan(&latestMonth)
	if err != nil || latestMonth == nil {
		return 0
	}

	// Sum outbound qty in that month
	var totalQty float64
	err = r.Pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(so_luong), 0)
		 FROM outbound_items
		 WHERE ma_hang = $1
		   AND ngay_nhan_hang >= $2
		   AND ngay_nhan_hang < $2 + INTERVAL '1 month'`,
		maHang, latestMonth).Scan(&totalQty)
	if err != nil || totalQty == 0 {
		return 0
	}

	// Days in that month
	nextMonth := latestMonth.AddDate(0, 1, 0)
	daysInMonth := nextMonth.Sub(*latestMonth).Hours() / 24

	if daysInMonth <= 0 {
		return 0
	}

	lbbq := totalQty / daysInMonth
	// Round to 4 decimal places
	return float64(int(lbbq*10000)) / 10000
}
