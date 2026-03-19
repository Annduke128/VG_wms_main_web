package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"wms-v1/internal/domain"
)

// InsertInboundItem inserts a new inbound item row
func (r *PostgresRepo) InsertInboundItem(ctx context.Context, tx pgx.Tx, item *domain.InboundItem) error {
	query := `
		INSERT INTO inbound_items
			(ma_hang, ten_san_pham, don_vi_tinh, quy_cach, so_luong,
			 doanh_so, chiet_khau, so_luong_tra_lai, doanh_thu, von,
			 lai_gop, ti_le_lai_gop, ngay_nhan_hang, batch_code)
		VALUES ($1,$2,$3,$4,$5,$6,$7,0,$8,$9,0,0,$10,$11)
		RETURNING id`

	return tx.QueryRow(ctx, query,
		item.MaHang, item.TenSanPham, item.DonViTinh, item.QuyCach, item.SoLuong,
		item.DoanhSo, item.ChietKhau, item.DoanhThu, item.Von,
		item.NgayNhanHang, item.BatchCode,
	).Scan(&item.ID)
}

// UpsertInventoryLot inserts or updates a lot for the given ma_hang + batch_code
func (r *PostgresRepo) UpsertInventoryLot(ctx context.Context, tx pgx.Tx, lot *domain.InventoryLot) error {
	query := `
		INSERT INTO inventory_lots (ma_hang, batch_code, received_at, qty_in, qty_out, qty_remaining)
		VALUES ($1, $2, $3, $4, 0, $4)
		ON CONFLICT (ma_hang, batch_code) DO UPDATE SET
			qty_in = inventory_lots.qty_in + EXCLUDED.qty_in,
			qty_remaining = inventory_lots.qty_remaining + EXCLUDED.qty_in
		RETURNING id, qty_in, qty_out, qty_remaining`

	return tx.QueryRow(ctx, query,
		lot.MaHang, lot.BatchCode, lot.ReceivedAt, lot.QtyIn,
	).Scan(&lot.ID, &lot.QtyIn, &lot.QtyOut, &lot.QtyRemaining)
}

// GetAvailableLotsFIFO returns lots with remaining qty, oldest first
func (r *PostgresRepo) GetAvailableLotsFIFO(ctx context.Context, tx pgx.Tx, maHang string) ([]domain.InventoryLot, error) {
	query := `
		SELECT id, ma_hang, batch_code, received_at, qty_in, qty_out, qty_remaining, created_at
		FROM inventory_lots
		WHERE ma_hang = $1 AND qty_remaining > 0
		ORDER BY received_at ASC, id ASC
		FOR UPDATE`

	rows, err := tx.Query(ctx, query, maHang)
	if err != nil {
		return nil, fmt.Errorf("query FIFO lots: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.InventoryLot])
}

// DeductLot reduces qty_remaining on a lot after outbound allocation
func (r *PostgresRepo) DeductLot(ctx context.Context, tx pgx.Tx, lotID int64, qty float64) error {
	query := `
		UPDATE inventory_lots
		SET qty_out = qty_out + $1, qty_remaining = qty_remaining - $1
		WHERE id = $2 AND qty_remaining >= $1`

	tag, err := tx.Exec(ctx, query, qty, lotID)
	if err != nil {
		return fmt.Errorf("deduct lot %d: %w", lotID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("lot %d: insufficient remaining qty", lotID)
	}
	return nil
}

// InsertOutboundItem inserts a single outbound row (one per FIFO allocation)
func (r *PostgresRepo) InsertOutboundItem(ctx context.Context, tx pgx.Tx, item *domain.OutboundItem) error {
	query := `
		INSERT INTO outbound_items
			(ma_hang, ten_san_pham, don_vi_tinh, quy_cach, so_luong,
			 doanh_so, chiet_khau, so_luong_tra_lai, doanh_thu, von,
			 lai_gop, ti_le_lai_gop, ngay_nhan_hang, batch_code)
		VALUES ($1,$2,$3,$4,$5,$6,$7,0,$8,$9,0,0,NOW(),$10)
		RETURNING id`

	return tx.QueryRow(ctx, query,
		item.MaHang, item.TenSanPham, item.DonViTinh, item.QuyCach, item.SoLuong,
		item.DoanhSo, item.ChietKhau, item.DoanhThu, item.Von,
		item.BatchCode,
	).Scan(&item.ID)
}

// UpdateInventoryMainInbound adds qty to inventory_main after inbound
func (r *PostgresRepo) UpdateInventoryMainInbound(ctx context.Context, tx pgx.Tx, maHang string, qty float64) error {
	query := `
		UPDATE inventory_main
		SET so_ton = so_ton + $1, so_nhap = so_nhap + $1
		WHERE ma_hang = $2`

	tag, err := tx.Exec(ctx, query, qty, maHang)
	if err != nil {
		return fmt.Errorf("update inventory inbound: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Insert if not exists
		insertQuery := `
			INSERT INTO inventory_main (ma_hang, ten_san_pham, so_ton, so_nhap, so_xuat)
			SELECT $1, ten_san_pham, $2, $2, 0 FROM products WHERE ma_hang = $1`
		_, err = tx.Exec(ctx, insertQuery, maHang, qty)
		if err != nil {
			return fmt.Errorf("insert inventory_main: %w", err)
		}
	}
	return nil
}

// UpdateInventoryMainOutbound subtracts qty from inventory_main after outbound
func (r *PostgresRepo) UpdateInventoryMainOutbound(ctx context.Context, tx pgx.Tx, maHang string, qty float64) error {
	query := `
		UPDATE inventory_main
		SET so_ton = so_ton - $1, so_xuat = so_xuat + $1
		WHERE ma_hang = $2`

	_, err := tx.Exec(ctx, query, qty, maHang)
	if err != nil {
		return fmt.Errorf("update inventory outbound: %w", err)
	}
	return nil
}

// InsertInventoryMovement records a movement event
func (r *PostgresRepo) InsertInventoryMovement(ctx context.Context, tx pgx.Tx, maHang string, qty float64, movementType string) error {
	query := `
		INSERT INTO inventory_movements (ma_hang, qty, movement_type)
		VALUES ($1, $2, $3)`

	_, err := tx.Exec(ctx, query, maHang, qty, movementType)
	return err
}

// ListOrders returns UNION of inbound + outbound, newest first
func (r *PostgresRepo) ListOrders(ctx context.Context, orderType string, limit, offset int) ([]domain.OrderListItem, int64, error) {
	var dataQuery, countQuery string

	switch orderType {
	case "in":
		dataQuery = `
			SELECT id, 'IN' as order_type, ma_hang, ten_san_pham, don_vi_tinh,
				   so_luong, batch_code, doanh_so, doanh_thu, ngay_nhan_hang
			FROM inbound_items
			ORDER BY ngay_nhan_hang DESC
			LIMIT $1 OFFSET $2`
		countQuery = `SELECT COUNT(*) FROM inbound_items`
	case "out":
		dataQuery = `
			SELECT id, 'OUT' as order_type, ma_hang, ten_san_pham, don_vi_tinh,
				   so_luong, batch_code, doanh_so, doanh_thu, ngay_nhan_hang
			FROM outbound_items
			ORDER BY ngay_nhan_hang DESC
			LIMIT $1 OFFSET $2`
		countQuery = `SELECT COUNT(*) FROM outbound_items`
	default:
		dataQuery = `
			(SELECT id, 'IN' as order_type, ma_hang, ten_san_pham, don_vi_tinh,
				    so_luong, batch_code, doanh_so, doanh_thu, ngay_nhan_hang
			 FROM inbound_items)
			UNION ALL
			(SELECT id, 'OUT' as order_type, ma_hang, ten_san_pham, don_vi_tinh,
				    so_luong, batch_code, doanh_so, doanh_thu, ngay_nhan_hang
			 FROM outbound_items)
			ORDER BY ngay_nhan_hang DESC
			LIMIT $1 OFFSET $2`
		countQuery = `SELECT (SELECT COUNT(*) FROM inbound_items) + (SELECT COUNT(*) FROM outbound_items)`
	}

	var total int64
	if err := r.Pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	rows, err := r.Pool.Query(ctx, dataQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.OrderListItem])
	if err != nil {
		return nil, 0, fmt.Errorf("scan orders: %w", err)
	}

	return items, total, nil
}

// GetLotsByMaHang returns all lots for a product
func (r *PostgresRepo) GetLotsByMaHang(ctx context.Context, maHang string) ([]domain.InventoryLot, error) {
	query := `
		SELECT id, ma_hang, batch_code, received_at, qty_in, qty_out, qty_remaining, created_at
		FROM inventory_lots
		WHERE ma_hang = $1
		ORDER BY received_at ASC`

	rows, err := r.Pool.Query(ctx, query, maHang)
	if err != nil {
		return nil, fmt.Errorf("query lots: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.InventoryLot])
}

// BeginTx starts a new transaction
func (r *PostgresRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.Pool.Begin(ctx)
}

// --- Thresholds ---

// SaveThreshold inserts a new threshold (closes current active one)
func (r *PostgresRepo) SaveThreshold(ctx context.Context, tx pgx.Tx, t *domain.InventoryThreshold) error {
	// Close any active threshold for this ma_hang
	closeQuery := `
		UPDATE inventory_thresholds
		SET effective_to = NOW()
		WHERE ma_hang = $1 AND effective_to IS NULL`
	_, _ = tx.Exec(ctx, closeQuery, t.MaHang)

	// Insert new threshold
	insertQuery := `
		INSERT INTO inventory_thresholds
			(ma_hang, min_qty, optimal_qty, max_age_days, source, model_version, effective_from)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	return tx.QueryRow(ctx, insertQuery,
		t.MaHang, t.MinQty, t.OptimalQty, t.MaxAgeDays,
		t.Source, t.ModelVersion, t.EffectiveFrom,
	).Scan(&t.ID)
}

// GetActiveThresholds returns current active thresholds
func (r *PostgresRepo) GetActiveThresholds(ctx context.Context, maHang string) ([]domain.InventoryThreshold, error) {
	query := `
		SELECT id, ma_hang, min_qty, optimal_qty, max_age_days,
		       source, model_version, effective_from, effective_to, created_at
		FROM inventory_thresholds
		WHERE ($1 = '' OR ma_hang = $1) AND effective_to IS NULL
		ORDER BY ma_hang`

	rows, err := r.Pool.Query(ctx, query, maHang)
	if err != nil {
		return nil, fmt.Errorf("query thresholds: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.InventoryThreshold])
}

// GetThresholdHistory returns all thresholds for a product (including expired)
func (r *PostgresRepo) GetThresholdHistory(ctx context.Context, maHang string) ([]domain.InventoryThreshold, error) {
	query := `
		SELECT id, ma_hang, min_qty, optimal_qty, max_age_days,
		       source, model_version, effective_from, effective_to, created_at
		FROM inventory_thresholds
		WHERE ma_hang = $1
		ORDER BY effective_from DESC`

	rows, err := r.Pool.Query(ctx, query, maHang)
	if err != nil {
		return nil, fmt.Errorf("query threshold history: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.InventoryThreshold])
}

// Needed by time import
var _ = time.Now
