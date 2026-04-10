package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"wms-v1/internal/domain"
)

// --- Accessory CRUD ---

// ListAccessories returns all accessories
func (r *PostgresRepo) ListAccessories(ctx context.Context) ([]domain.Accessory, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT ma_phu_kien, ten_phu_kien, don_vi_tinh, created_at
		FROM accessories
		ORDER BY ten_phu_kien`)
	if err != nil {
		return nil, fmt.Errorf("list accessories: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.Accessory])
}

// CreateAccessory inserts a new accessory + inventory row
func (r *PostgresRepo) CreateAccessory(ctx context.Context, tx pgx.Tx, acc *domain.Accessory) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO accessories (ma_phu_kien, ten_phu_kien, don_vi_tinh)
		VALUES ($1, $2, $3)`,
		acc.MaPhuKien, acc.TenPhuKien, acc.DonViTinh)
	if err != nil {
		return fmt.Errorf("insert accessory: %w", err)
	}

	// Create inventory rows for all active warehouses
	_, err = tx.Exec(ctx, `
		INSERT INTO accessory_inventory (ma_phu_kien, warehouse_id, so_ton)
		SELECT $1, id, 0 FROM warehouses WHERE is_active = TRUE
		ON CONFLICT (ma_phu_kien, warehouse_id) DO NOTHING`,
		acc.MaPhuKien)
	if err != nil {
		return fmt.Errorf("insert accessory inventory: %w", err)
	}

	return nil
}

// --- Accessory Inventory ---

// GetAccessoryInventory returns all accessory inventory with names
func (r *PostgresRepo) GetAccessoryInventory(ctx context.Context, warehouseID int64) ([]domain.AccessoryInventory, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT ai.ma_phu_kien, ai.warehouse_id, ai.so_ton, ai.updated_at,
		       a.ten_phu_kien, a.don_vi_tinh
		FROM accessory_inventory ai
		JOIN accessories a ON a.ma_phu_kien = ai.ma_phu_kien
		WHERE ai.warehouse_id = $1
		ORDER BY a.ten_phu_kien`, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("list accessory inventory: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.AccessoryInventory])
}

// UpdateAccessoryStock adjusts accessory stock (positive = add, negative = subtract)
func (r *PostgresRepo) UpdateAccessoryStock(ctx context.Context, tx pgx.Tx, maPhuKien string, warehouseID int64, delta float64) error {
	tag, err := tx.Exec(ctx, `
		UPDATE accessory_inventory
		SET so_ton = so_ton + $3, updated_at = NOW()
		WHERE ma_phu_kien = $1 AND warehouse_id = $2`,
		maPhuKien, warehouseID, delta)
	if err != nil {
		return fmt.Errorf("update accessory stock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("accessory %s not found", maPhuKien)
	}
	return nil
}

// GetAccessoryStockForUpdate gets current stock with lock
func (r *PostgresRepo) GetAccessoryStockForUpdate(ctx context.Context, tx pgx.Tx, maPhuKien string, warehouseID int64) (float64, error) {
	var soTon float64
	err := tx.QueryRow(ctx, `
		SELECT so_ton FROM accessory_inventory
		WHERE ma_phu_kien = $1 AND warehouse_id = $2
		FOR UPDATE`, maPhuKien, warehouseID).Scan(&soTon)
	if err != nil {
		return 0, fmt.Errorf("get accessory stock: %w", err)
	}
	return soTon, nil
}

// InsertAccessoryMovement records an accessory movement
func (r *PostgresRepo) InsertAccessoryMovement(ctx context.Context, tx pgx.Tx, maPhuKien, movementType string, soLuong float64, note string, warehouseID int64) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO accessory_movements (ma_phu_kien, movement_type, so_luong, note, warehouse_id)
		VALUES ($1, $2, $3, $4, $5)`,
		maPhuKien, movementType, soLuong, note, warehouseID)
	if err != nil {
		return fmt.Errorf("insert accessory movement: %w", err)
	}
	return nil
}
