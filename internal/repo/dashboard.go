package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"wms-v1/internal/domain"
)

// GetDashboardSummary returns the 4 KPI values
func (r *PostgresRepo) GetDashboardSummary(ctx context.Context) (*domain.DashboardSummary, error) {
	var summary domain.DashboardSummary

	// SKU count = distinct products with stock
	err := r.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM inventory_main WHERE so_ton > 0`).Scan(&summary.SKUCount)
	if err != nil {
		return nil, fmt.Errorf("sku count: %w", err)
	}

	// Tổng tiền hàng = SUM(so_ton * don_gia) from inventory_main JOIN products
	err = r.Pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(im.so_ton * p.don_gia), 0)
		 FROM inventory_main im
		 JOIN products p ON im.ma_hang = p.ma_hang`).Scan(&summary.TongTienHang)
	if err != nil {
		return nil, fmt.Errorf("tong tien hang: %w", err)
	}

	// SKU tồn lâu: so_ton >= optimal_qty AND so_ngay_ton >= max_age_days
	err = r.Pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT im.ma_hang)
		 FROM inventory_main im
		 JOIN inventory_thresholds it ON im.ma_hang = it.ma_hang
		   AND it.effective_from <= NOW()
		   AND (it.effective_to IS NULL OR it.effective_to > NOW())
		 WHERE im.so_ton >= it.optimal_qty
		   AND im.so_ngay_ton >= it.max_age_days`).Scan(&summary.SKUTonLau)
	if err != nil {
		return nil, fmt.Errorf("sku ton lau: %w", err)
	}

	// SKU thiếu hàng: so_ton < min_qty
	err = r.Pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT im.ma_hang)
		 FROM inventory_main im
		 JOIN inventory_thresholds it ON im.ma_hang = it.ma_hang
		   AND it.effective_from <= NOW()
		   AND (it.effective_to IS NULL OR it.effective_to > NOW())
		 WHERE im.so_ton < it.min_qty`).Scan(&summary.SKUThieuHang)
	if err != nil {
		return nil, fmt.Errorf("sku thieu hang: %w", err)
	}

	return &summary, nil
}

// GetInboundOutboundByWeek returns inbound/outbound totals per week for last N weeks
func (r *PostgresRepo) GetInboundOutboundByWeek(ctx context.Context, weeks int) ([]domain.InOutWeekData, error) {
	query := `
		WITH weeks AS (
			SELECT generate_series(
				date_trunc('week', NOW()) - ($1 - 1) * interval '1 week',
				date_trunc('week', NOW()),
				interval '1 week'
			) AS week_start
		),
		inbound_agg AS (
			SELECT date_trunc('week', ngay_nhan_hang) AS week_start,
			       COALESCE(SUM(so_luong), 0) AS total
			FROM inbound_items
			WHERE ngay_nhan_hang >= date_trunc('week', NOW()) - ($1 - 1) * interval '1 week'
			GROUP BY 1
		),
		outbound_agg AS (
			SELECT date_trunc('week', ngay_nhan_hang) AS week_start,
			       COALESCE(SUM(so_luong), 0) AS total
			FROM outbound_items
			WHERE ngay_nhan_hang >= date_trunc('week', NOW()) - ($1 - 1) * interval '1 week'
			GROUP BY 1
		)
		SELECT
			to_char(w.week_start, 'IYYY-"W"IW') AS week,
			to_char(w.week_start, 'YYYY-MM-DD') AS week_start,
			COALESCE(i.total, 0) AS inbound,
			COALESCE(o.total, 0) AS outbound
		FROM weeks w
		LEFT JOIN inbound_agg i ON i.week_start = w.week_start
		LEFT JOIN outbound_agg o ON o.week_start = w.week_start
		ORDER BY w.week_start
	`

	rows, err := r.Pool.Query(ctx, query, weeks)
	if err != nil {
		return nil, fmt.Errorf("inbound/outbound by week: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.InOutWeekData])
}

// GetInventoryVsOptimal returns top SKUs with their current stock vs optimal
func (r *PostgresRepo) GetInventoryVsOptimal(ctx context.Context, limit int) ([]domain.InventoryVsOptimalItem, error) {
	query := `
		SELECT im.ma_hang, im.ten_san_pham, im.so_ton,
		       COALESCE(it.optimal_qty, 0) AS optimal_qty
		FROM inventory_main im
		LEFT JOIN inventory_thresholds it ON im.ma_hang = it.ma_hang
		  AND it.effective_from <= NOW()
		  AND (it.effective_to IS NULL OR it.effective_to > NOW())
		WHERE im.so_ton > 0
		ORDER BY im.so_ton DESC
		LIMIT $1
	`

	rows, err := r.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("inventory vs optimal: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.InventoryVsOptimalItem])
}

// GetAlerts returns inventory alerts (tồn lâu + thiếu hàng)
func (r *PostgresRepo) GetAlerts(ctx context.Context) ([]domain.AlertItem, error) {
	query := `
		SELECT im.ma_hang, im.ten_san_pham, im.so_ton,
		       'ton_lau' AS alert_type,
		       im.so_ngay_ton,
		       it.max_age_days,
		       it.optimal_qty,
		       0 AS min_qty
		FROM inventory_main im
		JOIN inventory_thresholds it ON im.ma_hang = it.ma_hang
		  AND it.effective_from <= NOW()
		  AND (it.effective_to IS NULL OR it.effective_to > NOW())
		WHERE im.so_ton >= it.optimal_qty
		  AND im.so_ngay_ton >= it.max_age_days

		UNION ALL

		SELECT im.ma_hang, im.ten_san_pham, im.so_ton,
		       'thieu_hang' AS alert_type,
		       im.so_ngay_ton,
		       0 AS max_age_days,
		       0 AS optimal_qty,
		       it.min_qty
		FROM inventory_main im
		JOIN inventory_thresholds it ON im.ma_hang = it.ma_hang
		  AND it.effective_from <= NOW()
		  AND (it.effective_to IS NULL OR it.effective_to > NOW())
		WHERE im.so_ton < it.min_qty

		ORDER BY alert_type, ma_hang
	`

	rows, err := r.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get alerts: %w", err)
	}
	defer rows.Close()

	var alerts []domain.AlertItem
	for rows.Next() {
		var a domain.AlertItem
		err := rows.Scan(&a.MaHang, &a.TenSanPham, &a.SoTon, &a.AlertType,
			&a.SoNgayTon, &a.MaxAgeDays, &a.OptimalQty, &a.MinQty)
		if err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, a)
	}

	if alerts == nil {
		alerts = []domain.AlertItem{}
	}

	return alerts, nil
}

// GetZeroSalesSKUs returns SKUs with LBBQ=0 and so_ton>0 (zero sales)
func (r *PostgresRepo) GetZeroSalesSKUs(ctx context.Context) ([]domain.ZeroSalesItem, error) {
	query := `
		SELECT im.ma_hang, im.ten_san_pham, im.so_ton,
		       im.luong_ban_binh_quan_ngay,
		       COALESCE(to_char(date_trunc('month', sub.latest_out), 'MM/YYYY'), '') AS latest_outbound_month
		FROM inventory_main im
		LEFT JOIN (
			SELECT ma_hang, MAX(ngay_nhan_hang) AS latest_out
			FROM outbound_items
			GROUP BY ma_hang
		) sub ON im.ma_hang = sub.ma_hang
		WHERE im.so_ton > 0
		  AND im.luong_ban_binh_quan_ngay = 0
		ORDER BY im.so_ton DESC
	`

	rows, err := r.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get zero sales: %w", err)
	}
	defer rows.Close()

	var items []domain.ZeroSalesItem
	for rows.Next() {
		var item domain.ZeroSalesItem
		if err := rows.Scan(&item.MaHang, &item.TenSanPham, &item.SoTon,
			&item.LuongBanBinhQuanNgay, &item.LatestOutboundMonth); err != nil {
			return nil, fmt.Errorf("scan zero sales: %w", err)
		}
		items = append(items, item)
	}

	if items == nil {
		items = []domain.ZeroSalesItem{}
	}
	return items, nil
}

// GetRestockAlerts returns SKUs that were sold before, now so_ton=0, 1-7 days since last outbound
func (r *PostgresRepo) GetRestockAlerts(ctx context.Context) ([]domain.RestockAlertItem, error) {
	query := `
		SELECT im.ma_hang, im.ten_san_pham, im.so_ton,
		       CASE
		         WHEN sub.last_out IS NOT NULL
		           THEN EXTRACT(DAY FROM NOW() - sub.last_out)::int
		         ELSE -1
		       END AS ngay_since_last_out,
		       CASE
		         WHEN sub.last_out IS NOT NULL
		           THEN to_char(sub.last_out, 'YYYY-MM-DD')
		         ELSE 'Chưa có dữ liệu xuất hàng'
		       END AS last_outbound_date
		FROM inventory_main im
		LEFT JOIN (
			SELECT ma_hang, MAX(ngay_nhan_hang) AS last_out
			FROM outbound_items
			GROUP BY ma_hang
		) sub ON im.ma_hang = sub.ma_hang
		WHERE im.so_ton = 0
		  AND sub.last_out IS NOT NULL
		  AND EXTRACT(DAY FROM NOW() - sub.last_out) BETWEEN 1 AND 7
		ORDER BY EXTRACT(DAY FROM NOW() - sub.last_out) ASC
	`

	rows, err := r.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get restock alerts: %w", err)
	}
	defer rows.Close()

	var items []domain.RestockAlertItem
	for rows.Next() {
		var item domain.RestockAlertItem
		if err := rows.Scan(&item.MaHang, &item.TenSanPham, &item.SoTon,
			&item.NgaySinceLastOut, &item.LastOutboundDate); err != nil {
			return nil, fmt.Errorf("scan restock alert: %w", err)
		}
		items = append(items, item)
	}

	if items == nil {
		items = []domain.RestockAlertItem{}
	}
	return items, nil
}

// GetThresholdsByMaHang returns active threshold + history for a SKU
func (r *PostgresRepo) GetThresholdsByMaHang(ctx context.Context, maHang string) ([]domain.InventoryThreshold, error) {
	query := `
		SELECT id, ma_hang, min_qty, optimal_qty, max_age_days,
		       source, model_version, effective_from, effective_to, created_at
		FROM inventory_thresholds
		WHERE ma_hang = $1
		ORDER BY created_at DESC
	`

	rows, err := r.Pool.Query(ctx, query, maHang)
	if err != nil {
		return nil, fmt.Errorf("get thresholds: %w", err)
	}
	defer rows.Close()

	var thresholds []domain.InventoryThreshold
	for rows.Next() {
		var t domain.InventoryThreshold
		err := rows.Scan(&t.ID, &t.MaHang, &t.MinQty, &t.OptimalQty, &t.MaxAgeDays,
			&t.Source, &t.ModelVersion, &t.EffectiveFrom, &t.EffectiveTo, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan threshold: %w", err)
		}
		thresholds = append(thresholds, t)
	}

	if thresholds == nil {
		thresholds = []domain.InventoryThreshold{}
	}

	return thresholds, nil
}

// SaveThreshold closes active threshold and inserts new one (within TX)
func (r *PostgresRepo) SaveThresholdEntry(ctx context.Context, req domain.ThresholdRequest) (*domain.InventoryThreshold, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()

	// Close active threshold for this ma_hang
	_, err = tx.Exec(ctx,
		`UPDATE inventory_thresholds
		 SET effective_to = $1
		 WHERE ma_hang = $2
		   AND effective_from <= $1
		   AND (effective_to IS NULL OR effective_to > $1)`,
		now, req.MaHang)
	if err != nil {
		return nil, fmt.Errorf("close active threshold: %w", err)
	}

	// Set effective_from
	effectiveFrom := now
	if req.EffectiveFrom != nil {
		effectiveFrom = *req.EffectiveFrom
	}

	source := req.Source
	if source == "" {
		source = "manual"
	}

	// Insert new threshold
	var t domain.InventoryThreshold
	err = tx.QueryRow(ctx,
		`INSERT INTO inventory_thresholds (ma_hang, min_qty, optimal_qty, max_age_days, source, model_version, effective_from, effective_to)
		 VALUES ($1, $2, $3, $4, $5, '', $6, $7)
		 RETURNING id, ma_hang, min_qty, optimal_qty, max_age_days, source, model_version, effective_from, effective_to, created_at`,
		req.MaHang, req.MinQty, req.OptimalQty, req.MaxAgeDays, source, effectiveFrom, req.EffectiveTo,
	).Scan(&t.ID, &t.MaHang, &t.MinQty, &t.OptimalQty, &t.MaxAgeDays,
		&t.Source, &t.ModelVersion, &t.EffectiveFrom, &t.EffectiveTo, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert threshold: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &t, nil
}
