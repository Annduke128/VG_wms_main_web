package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"wms-v1/internal/domain"
)

// --- Combo Master CRUD ---

// ListComboMasters returns all combo masters (active only by default)
func (r *PostgresRepo) ListComboMasters(ctx context.Context, activeOnly bool) ([]domain.ComboMaster, error) {
	query := `SELECT ma_combo, ten_combo, mo_ta, active, created_at, updated_at FROM combo_master`
	if activeOnly {
		query += ` WHERE active = TRUE`
	}
	query += ` ORDER BY ten_combo`

	rows, err := r.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list combo masters: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.ComboMaster])
}

// GetComboDetail returns a combo with its full BOM
func (r *PostgresRepo) GetComboDetail(ctx context.Context, maCombo string) (*domain.ComboDetail, error) {
	// Master
	var master domain.ComboMaster
	err := r.Pool.QueryRow(ctx,
		`SELECT ma_combo, ten_combo, mo_ta, active, created_at, updated_at
		 FROM combo_master WHERE ma_combo = $1`, maCombo,
	).Scan(&master.MaCombo, &master.TenCombo, &master.MoTa, &master.Active, &master.CreatedAt, &master.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get combo master: %w", err)
	}

	// BOM Semi
	semiRows, err := r.Pool.Query(ctx,
		`SELECT b.id, b.ma_combo, b.ma_hang, b.so_luong, COALESCE(p.ten_san_pham,'') as ten_san_pham
		 FROM combo_bom_semi b
		 LEFT JOIN products p ON p.ma_hang = b.ma_hang
		 WHERE b.ma_combo = $1
		 ORDER BY b.id`, maCombo)
	if err != nil {
		return nil, fmt.Errorf("get combo bom semi: %w", err)
	}
	defer semiRows.Close()
	semi, err := pgx.CollectRows(semiRows, pgx.RowToStructByName[domain.ComboBOMSemi])
	if err != nil {
		return nil, fmt.Errorf("scan combo bom semi: %w", err)
	}

	// BOM Accessory
	accRows, err := r.Pool.Query(ctx,
		`SELECT b.id, b.ma_combo, b.ma_phu_kien, b.so_luong, COALESCE(a.ten_phu_kien,'') as ten_phu_kien
		 FROM combo_bom_accessory b
		 LEFT JOIN accessories a ON a.ma_phu_kien = b.ma_phu_kien
		 WHERE b.ma_combo = $1
		 ORDER BY b.id`, maCombo)
	if err != nil {
		return nil, fmt.Errorf("get combo bom accessory: %w", err)
	}
	defer accRows.Close()
	acc, err := pgx.CollectRows(accRows, pgx.RowToStructByName[domain.ComboBOMAccessory])
	if err != nil {
		return nil, fmt.Errorf("scan combo bom accessory: %w", err)
	}

	return &domain.ComboDetail{
		ComboMaster:  master,
		BOMSemi:      semi,
		BOMAccessory: acc,
	}, nil
}

// SaveComboMaster upserts a combo master and replaces its BOM
func (r *PostgresRepo) SaveComboMaster(ctx context.Context, tx pgx.Tx, req domain.SaveComboMasterRequest) error {
	// Upsert master
	_, err := tx.Exec(ctx, `
		INSERT INTO combo_master (ma_combo, ten_combo, mo_ta)
		VALUES ($1, $2, $3)
		ON CONFLICT (ma_combo) DO UPDATE SET
			ten_combo = EXCLUDED.ten_combo,
			mo_ta = EXCLUDED.mo_ta,
			updated_at = NOW()`,
		req.MaCombo, req.TenCombo, req.MoTa)
	if err != nil {
		return fmt.Errorf("upsert combo master: %w", err)
	}

	// Ensure combo_inventory rows exist for all active warehouses
	_, err = tx.Exec(ctx, `
		INSERT INTO combo_inventory (ma_combo, warehouse_id, so_ton, so_nhap, so_xuat, so_tra)
		SELECT $1, id, 0, 0, 0, 0 FROM warehouses WHERE is_active = TRUE
		ON CONFLICT (ma_combo, warehouse_id) DO NOTHING`, req.MaCombo)
	if err != nil {
		return fmt.Errorf("ensure combo inventory: %w", err)
	}

	// Replace BOM semi
	_, err = tx.Exec(ctx, `DELETE FROM combo_bom_semi WHERE ma_combo = $1`, req.MaCombo)
	if err != nil {
		return fmt.Errorf("delete bom semi: %w", err)
	}
	for _, item := range req.BOMSemi {
		_, err = tx.Exec(ctx, `
			INSERT INTO combo_bom_semi (ma_combo, ma_hang, so_luong)
			VALUES ($1, $2, $3)`,
			req.MaCombo, item.MaComponent, item.SoLuong)
		if err != nil {
			return fmt.Errorf("insert bom semi %s: %w", item.MaComponent, err)
		}
	}

	// Replace BOM accessory
	_, err = tx.Exec(ctx, `DELETE FROM combo_bom_accessory WHERE ma_combo = $1`, req.MaCombo)
	if err != nil {
		return fmt.Errorf("delete bom accessory: %w", err)
	}
	for _, item := range req.BOMAccessory {
		_, err = tx.Exec(ctx, `
			INSERT INTO combo_bom_accessory (ma_combo, ma_phu_kien, so_luong)
			VALUES ($1, $2, $3)`,
			req.MaCombo, item.MaComponent, item.SoLuong)
		if err != nil {
			return fmt.Errorf("insert bom accessory %s: %w", item.MaComponent, err)
		}
	}

	return nil
}

// DeleteComboMaster deactivates a combo (soft delete)
func (r *PostgresRepo) DeleteComboMaster(ctx context.Context, maCombo string) error {
	_, err := r.Pool.Exec(ctx,
		`UPDATE combo_master SET active = FALSE, updated_at = NOW() WHERE ma_combo = $1`, maCombo)
	if err != nil {
		return fmt.Errorf("deactivate combo: %w", err)
	}
	return nil
}

// --- Combo Inventory ---

// GetComboInventory returns all combo inventory with names
func (r *PostgresRepo) GetComboInventory(ctx context.Context, warehouseID int64) ([]domain.ComboInventory, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT ci.ma_combo, ci.warehouse_id, ci.so_ton, ci.so_nhap, ci.so_xuat, ci.so_tra, ci.updated_at,
		       cm.ten_combo
		FROM combo_inventory ci
		JOIN combo_master cm ON cm.ma_combo = ci.ma_combo
		WHERE cm.active = TRUE AND ci.warehouse_id = $1
		ORDER BY cm.ten_combo`, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("list combo inventory: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.ComboInventory])
}

// UpdateComboInventory atomically adjusts combo inventory
func (r *PostgresRepo) UpdateComboInventory(ctx context.Context, tx pgx.Tx, maCombo string, warehouseID int64, soTonDelta, soNhapDelta, soXuatDelta, soTraDelta float64) error {
	_, err := tx.Exec(ctx, `
		UPDATE combo_inventory SET
			so_ton = so_ton + $3,
			so_nhap = so_nhap + $4,
			so_xuat = so_xuat + $5,
			so_tra = so_tra + $6,
			updated_at = NOW()
		WHERE ma_combo = $1 AND warehouse_id = $2`,
		maCombo, warehouseID, soTonDelta, soNhapDelta, soXuatDelta, soTraDelta)
	if err != nil {
		return fmt.Errorf("update combo inventory: %w", err)
	}
	return nil
}

// GetComboInventoryForUpdate gets current combo_inventory with FOR UPDATE lock
func (r *PostgresRepo) GetComboInventoryForUpdate(ctx context.Context, tx pgx.Tx, maCombo string, warehouseID int64) (*domain.ComboInventory, error) {
	var ci domain.ComboInventory
	err := tx.QueryRow(ctx, `
		SELECT ma_combo, warehouse_id, so_ton, so_nhap, so_xuat, so_tra, updated_at
		FROM combo_inventory
		WHERE ma_combo = $1 AND warehouse_id = $2
		FOR UPDATE`, maCombo, warehouseID).Scan(&ci.MaCombo, &ci.WarehouseID, &ci.SoTon, &ci.SoNhap, &ci.SoXuat, &ci.SoTra, &ci.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get combo inventory for update: %w", err)
	}
	return &ci, nil
}

// --- Combo Transactions ---

// InsertComboTransaction inserts a combo transaction and returns the ID
func (r *PostgresRepo) InsertComboTransaction(ctx context.Context, tx pgx.Tx, maCombo, txnType string, soLuong float64, note string, warehouseID int64) (int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `
		INSERT INTO combo_transactions (ma_combo, transaction_type, so_luong, note, warehouse_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		maCombo, txnType, soLuong, note, warehouseID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert combo transaction: %w", err)
	}
	return id, nil
}

// InsertComboComponentMovement records NVL consumption for a combo transaction
func (r *PostgresRepo) InsertComboComponentMovement(ctx context.Context, tx pgx.Tx, txnID int64, componentType, maComponent string, soLuong float64, lotID *int64, warehouseID int64) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO combo_component_movements (combo_transaction_id, component_type, ma_component, so_luong, lot_id, warehouse_id)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		txnID, componentType, maComponent, soLuong, lotID, warehouseID)
	if err != nil {
		return fmt.Errorf("insert component movement: %w", err)
	}
	return nil
}

// ListComboTransactions returns combo transactions with filters
func (r *PostgresRepo) ListComboTransactions(ctx context.Context, maCombo string, limit, offset int, warehouseID int64) ([]domain.ComboTransaction, int64, error) {
	whereClause := " WHERE ct.warehouse_id = $1"
	var args []interface{}
	args = append(args, warehouseID)
	paramIdx := 2

	if maCombo != "" {
		whereClause += fmt.Sprintf(" AND ct.ma_combo = $%d", paramIdx)
		args = append(args, maCombo)
		paramIdx++
	}

	// Count
	var total int64
	countQuery := `SELECT COUNT(*) FROM combo_transactions ct` + whereClause
	if err := r.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count combo transactions: %w", err)
	}

	// Data
	dataQuery := fmt.Sprintf(`
		SELECT ct.id, ct.ma_combo, ct.warehouse_id, ct.transaction_type, ct.so_luong, ct.note, ct.created_at,
		       cm.ten_combo
		FROM combo_transactions ct
		JOIN combo_master cm ON cm.ma_combo = ct.ma_combo
		%s
		ORDER BY ct.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, paramIdx, paramIdx+1)

	dataArgs := append(args, limit, offset)
	rows, err := r.Pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query combo transactions: %w", err)
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.ComboTransaction])
	if err != nil {
		return nil, 0, fmt.Errorf("scan combo transactions: %w", err)
	}

	return items, total, nil
}

// GetBOMForCombo returns the full BOM for a combo (used by business logic)
func (r *PostgresRepo) GetBOMForCombo(ctx context.Context, tx pgx.Tx, maCombo string) ([]domain.ComboBOMSemi, []domain.ComboBOMAccessory, error) {
	// Semi
	semiRows, err := tx.Query(ctx, `
		SELECT id, ma_combo, ma_hang, so_luong
		FROM combo_bom_semi
		WHERE ma_combo = $1`, maCombo)
	if err != nil {
		return nil, nil, fmt.Errorf("get bom semi: %w", err)
	}
	defer semiRows.Close()
	semi, err := pgx.CollectRows(semiRows, pgx.RowToStructByName[domain.ComboBOMSemi])
	if err != nil {
		return nil, nil, fmt.Errorf("scan bom semi: %w", err)
	}

	// Accessory
	accRows, err := tx.Query(ctx, `
		SELECT id, ma_combo, ma_phu_kien, so_luong
		FROM combo_bom_accessory
		WHERE ma_combo = $1`, maCombo)
	if err != nil {
		return nil, nil, fmt.Errorf("get bom accessory: %w", err)
	}
	defer accRows.Close()
	acc, err := pgx.CollectRows(accRows, pgx.RowToStructByName[domain.ComboBOMAccessory])
	if err != nil {
		return nil, nil, fmt.Errorf("scan bom accessory: %w", err)
	}

	return semi, acc, nil
}
