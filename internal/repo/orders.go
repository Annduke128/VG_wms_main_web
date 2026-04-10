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
			 lai_gop, ti_le_lai_gop, ngay_nhan_hang, batch_code, warehouse_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,0,$8,$9,0,0,$10,$11,$12)
		RETURNING id`

	return tx.QueryRow(ctx, query,
		item.MaHang, item.TenSanPham, item.DonViTinh, item.QuyCach, item.SoLuong,
		item.DoanhSo, item.ChietKhau, item.DoanhThu, item.Von,
		item.NgayNhanHang, item.BatchCode, item.WarehouseID,
	).Scan(&item.ID)
}

// UpsertInventoryLot inserts or updates a lot for the given ma_hang + batch_code
func (r *PostgresRepo) UpsertInventoryLot(ctx context.Context, tx pgx.Tx, lot *domain.InventoryLot) error {
	query := `
		INSERT INTO inventory_lots (ma_hang, batch_code, received_at, qty_in, qty_out, qty_remaining, warehouse_id)
		VALUES ($1, $2, $3, $4, 0, $4, $5)
		ON CONFLICT (ma_hang, batch_code, warehouse_id) DO UPDATE SET
			qty_in = inventory_lots.qty_in + EXCLUDED.qty_in,
			qty_remaining = inventory_lots.qty_remaining + EXCLUDED.qty_in
		RETURNING id, qty_in, qty_out, qty_remaining`

	return tx.QueryRow(ctx, query,
		lot.MaHang, lot.BatchCode, lot.ReceivedAt, lot.QtyIn, lot.WarehouseID,
	).Scan(&lot.ID, &lot.QtyIn, &lot.QtyOut, &lot.QtyRemaining)
}

// GetAvailableLotsFIFO returns lots with remaining qty, oldest first
func (r *PostgresRepo) GetAvailableLotsFIFO(ctx context.Context, tx pgx.Tx, maHang string, warehouseID int64) ([]domain.InventoryLot, error) {
	query := `
		SELECT id, ma_hang, batch_code, received_at, qty_in, qty_out, qty_remaining, created_at
		FROM inventory_lots
		WHERE ma_hang = $1 AND warehouse_id = $2 AND qty_remaining > 0
		ORDER BY received_at ASC, id ASC
		FOR UPDATE`

	rows, err := tx.Query(ctx, query, maHang, warehouseID)
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
			 lai_gop, ti_le_lai_gop, ngay_nhan_hang, batch_code, warehouse_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,0,$8,$9,0,0,NOW(),$10,$11)
		RETURNING id`

	return tx.QueryRow(ctx, query,
		item.MaHang, item.TenSanPham, item.DonViTinh, item.QuyCach, item.SoLuong,
		item.DoanhSo, item.ChietKhau, item.DoanhThu, item.Von,
		item.BatchCode, item.WarehouseID,
	).Scan(&item.ID)
}

// UpdateInventoryMainInbound adds qty to inventory_main after inbound
func (r *PostgresRepo) UpdateInventoryMainInbound(ctx context.Context, tx pgx.Tx, maHang string, qty float64, warehouseID int64) error {
	query := `
		UPDATE inventory_main
		SET so_ton = so_ton + $1, so_nhap = so_nhap + $1
		WHERE ma_hang = $2 AND warehouse_id = $3`

	tag, err := tx.Exec(ctx, query, qty, maHang, warehouseID)
	if err != nil {
		return fmt.Errorf("update inventory inbound: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Insert if not exists
		insertQuery := `
			INSERT INTO inventory_main (ma_hang, ten_san_pham, so_ton, so_nhap, so_xuat, warehouse_id)
			SELECT $1, ten_san_pham, $2, $2, 0, $3 FROM products WHERE ma_hang = $1`
		_, err = tx.Exec(ctx, insertQuery, maHang, qty, warehouseID)
		if err != nil {
			return fmt.Errorf("insert inventory_main: %w", err)
		}
	}
	return nil
}

// UpdateInventoryMainOutbound subtracts qty from inventory_main after outbound
func (r *PostgresRepo) UpdateInventoryMainOutbound(ctx context.Context, tx pgx.Tx, maHang string, qty float64, warehouseID int64) error {
	query := `
		UPDATE inventory_main
		SET so_ton = so_ton - $1, so_xuat = so_xuat + $1
		WHERE ma_hang = $2 AND warehouse_id = $3`

	_, err := tx.Exec(ctx, query, qty, maHang, warehouseID)
	if err != nil {
		return fmt.Errorf("update inventory outbound: %w", err)
	}
	return nil
}

// InsertInventoryMovement records a movement event
func (r *PostgresRepo) InsertInventoryMovement(ctx context.Context, tx pgx.Tx, maHang string, qty float64, movementType string, warehouseID int64) error {
	query := `
		INSERT INTO inventory_movements (ma_hang, qty, movement_type, warehouse_id)
		VALUES ($1, $2, $3, $4)`

	_, err := tx.Exec(ctx, query, maHang, qty, movementType, warehouseID)
	return err
}

// ListOrders returns UNION of inbound + outbound with optional filters, newest first.
// Supports filtering by date range, ma_bu, and ma_nhom_hang (via JOIN products).
func (r *PostgresRepo) ListOrders(ctx context.Context, f domain.OrderFilter) ([]domain.OrderListItem, int64, error) {
	// Build WHERE conditions + params (shared between data and count queries)
	needJoin := f.MaBu != "" || f.MaNhomHang != ""
	var conditions []string
	var args []interface{}
	paramIdx := 1

	if !f.DateFrom.IsZero() {
		conditions = append(conditions, fmt.Sprintf("t.ngay_nhan_hang >= $%d", paramIdx))
		args = append(args, f.DateFrom)
		paramIdx++
	}
	if !f.DateTo.IsZero() {
		conditions = append(conditions, fmt.Sprintf("t.ngay_nhan_hang <= $%d", paramIdx))
		args = append(args, f.DateTo)
		paramIdx++
	}
	if f.MaBu != "" {
		conditions = append(conditions, fmt.Sprintf("p.ma_bu = $%d", paramIdx))
		args = append(args, f.MaBu)
		paramIdx++
	}
	if f.MaNhomHang != "" {
		conditions = append(conditions, fmt.Sprintf("p.ma_nhom_hang = $%d", paramIdx))
		args = append(args, f.MaNhomHang)
		paramIdx++
	}
	if f.WarehouseID > 0 {
		conditions = append(conditions, fmt.Sprintf("t.warehouse_id = $%d", paramIdx))
		args = append(args, f.WarehouseID)
		paramIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + joinStrings(conditions, " AND ")
	}

	joinClause := ""
	if needJoin {
		joinClause = " JOIN products p ON t.ma_hang = p.ma_hang"
	}

	// Build per-table SELECT
	selectCols := `t.id, %s as order_type, t.ma_hang, t.ten_san_pham, t.don_vi_tinh,
		t.so_luong, t.batch_code, t.doanh_so, t.doanh_thu, t.ngay_nhan_hang, t.warehouse_id`

	buildSingle := func(table, typeLabel string) (string, string) {
		sel := fmt.Sprintf(selectCols, fmt.Sprintf("'%s'", typeLabel))
		from := fmt.Sprintf(" FROM %s t%s%s", table, joinClause, whereClause)
		data := "SELECT " + sel + from
		cnt := "SELECT COUNT(*)" + from
		return data, cnt
	}

	var dataQuery, countQuery string

	switch f.OrderType {
	case "in":
		dq, cq := buildSingle("inbound_items", "IN")
		dataQuery = dq + fmt.Sprintf(" ORDER BY t.ngay_nhan_hang DESC LIMIT $%d OFFSET $%d", paramIdx, paramIdx+1)
		countQuery = cq
	case "out":
		dq, cq := buildSingle("outbound_items", "OUT")
		dataQuery = dq + fmt.Sprintf(" ORDER BY t.ngay_nhan_hang DESC LIMIT $%d OFFSET $%d", paramIdx, paramIdx+1)
		countQuery = cq
	default:
		dqIn, cqIn := buildSingle("inbound_items", "IN")
		dqOut, cqOut := buildSingle("outbound_items", "OUT")
		dataQuery = fmt.Sprintf("(%s) UNION ALL (%s) ORDER BY ngay_nhan_hang DESC LIMIT $%d OFFSET $%d",
			dqIn, dqOut, paramIdx, paramIdx+1)
		countQuery = fmt.Sprintf("SELECT (%s) + (%s)", cqIn, cqOut)
	}

	// Count
	var total int64
	if err := r.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	// Data
	dataArgs := append(args, f.Limit, f.Offset)
	rows, err := r.Pool.Query(ctx, dataQuery, dataArgs...)
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

// joinStrings is defined in inventory.go — reused here

// GetLotsByMaHang returns all lots for a product
func (r *PostgresRepo) GetLotsByMaHang(ctx context.Context, maHang string, warehouseID int64) ([]domain.InventoryLot, error) {
	query := `
		SELECT id, ma_hang, batch_code, received_at, qty_in, qty_out, qty_remaining, created_at
		FROM inventory_lots
		WHERE ma_hang = $1 AND warehouse_id = $2
		ORDER BY received_at ASC`

	rows, err := r.Pool.Query(ctx, query, maHang, warehouseID)
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
		WHERE ma_hang = $1 AND warehouse_id = $2 AND effective_to IS NULL`
	_, _ = tx.Exec(ctx, closeQuery, t.MaHang, t.WarehouseID)

	// Insert new threshold
	insertQuery := `
		INSERT INTO inventory_thresholds
			(ma_hang, min_qty, optimal_qty, max_age_days, source, model_version, effective_from, warehouse_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	return tx.QueryRow(ctx, insertQuery,
		t.MaHang, t.MinQty, t.OptimalQty, t.MaxAgeDays,
		t.Source, t.ModelVersion, t.EffectiveFrom, t.WarehouseID,
	).Scan(&t.ID)
}

// GetActiveThresholds returns current active thresholds
func (r *PostgresRepo) GetActiveThresholds(ctx context.Context, maHang string, warehouseID int64) ([]domain.InventoryThreshold, error) {
	query := `
		SELECT id, ma_hang, min_qty, optimal_qty, max_age_days,
		       source, model_version, effective_from, effective_to, created_at
		FROM inventory_thresholds
		WHERE ($1 = '' OR ma_hang = $1) AND warehouse_id = $2 AND effective_to IS NULL
		ORDER BY ma_hang`

	rows, err := r.Pool.Query(ctx, query, maHang, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("query thresholds: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.InventoryThreshold])
}

// GetThresholdHistory returns all thresholds for a product (including expired)
func (r *PostgresRepo) GetThresholdHistory(ctx context.Context, maHang string, warehouseID int64) ([]domain.InventoryThreshold, error) {
	query := `
		SELECT id, ma_hang, min_qty, optimal_qty, max_age_days,
		       source, model_version, effective_from, effective_to, created_at
		FROM inventory_thresholds
		WHERE ma_hang = $1 AND warehouse_id = $2
		ORDER BY effective_from DESC`

	rows, err := r.Pool.Query(ctx, query, maHang, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("query threshold history: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.InventoryThreshold])
}

// Needed by time import
var _ = time.Now
