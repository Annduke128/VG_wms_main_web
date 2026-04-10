package repo

import (
	"context"
	"fmt"

	"wms-v1/internal/domain"
)

func (r *PostgresRepo) ListWarehouses(ctx context.Context) ([]domain.Warehouse, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT id, code, name, address, is_active, created_at, updated_at
		FROM warehouses ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list warehouses: %w", err)
	}
	defer rows.Close()

	var warehouses []domain.Warehouse
	for rows.Next() {
		var w domain.Warehouse
		if err := rows.Scan(&w.ID, &w.Code, &w.Name, &w.Address, &w.IsActive, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan warehouse: %w", err)
		}
		warehouses = append(warehouses, w)
	}
	return warehouses, rows.Err()
}

func (r *PostgresRepo) GetWarehouse(ctx context.Context, id int64) (*domain.Warehouse, error) {
	var w domain.Warehouse
	err := r.Pool.QueryRow(ctx, `
		SELECT id, code, name, address, is_active, created_at, updated_at
		FROM warehouses WHERE id = $1`, id).
		Scan(&w.ID, &w.Code, &w.Name, &w.Address, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get warehouse %d: %w", id, err)
	}
	return &w, nil
}

func (r *PostgresRepo) CreateWarehouse(ctx context.Context, req domain.CreateWarehouseRequest) (*domain.Warehouse, error) {
	var w domain.Warehouse
	err := r.Pool.QueryRow(ctx, `
		INSERT INTO warehouses (code, name, address)
		VALUES ('WH-' || LPAD(nextval('warehouse_code_seq')::TEXT, 4, '0'), $1, $2)
		RETURNING id, code, name, address, is_active, created_at, updated_at`,
		req.Name, req.Address).
		Scan(&w.ID, &w.Code, &w.Name, &w.Address, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create warehouse: %w", err)
	}
	return &w, nil
}

func (r *PostgresRepo) UpdateWarehouse(ctx context.Context, id int64, req domain.UpdateWarehouseRequest) (*domain.Warehouse, error) {
	var w domain.Warehouse
	err := r.Pool.QueryRow(ctx, `
		UPDATE warehouses SET
			name = COALESCE($2, name),
			address = COALESCE($3, address),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, code, name, address, is_active, created_at, updated_at`,
		id, req.Name, req.Address).
		Scan(&w.ID, &w.Code, &w.Name, &w.Address, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update warehouse %d: %w", id, err)
	}
	return &w, nil
}

// WarehouseExists checks if a warehouse exists and is active.
func (r *PostgresRepo) WarehouseExists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM warehouses WHERE id = $1 AND is_active = TRUE)`,
		id).Scan(&exists)
	return exists, err
}

// InitWarehouseInventory seeds zero-stock rows for all existing combos and
// accessories in a new warehouse. Call after CreateWarehouse.
func (r *PostgresRepo) InitWarehouseInventory(ctx context.Context, warehouseID int64) error {
	// Seed combo_inventory for all active combo masters
	_, err := r.Pool.Exec(ctx, `
		INSERT INTO combo_inventory (ma_combo, warehouse_id, so_ton, so_nhap, so_xuat, so_tra)
		SELECT ma_combo, $1, 0, 0, 0, 0 FROM combo_master WHERE active = TRUE
		ON CONFLICT (ma_combo, warehouse_id) DO NOTHING`, warehouseID)
	if err != nil {
		return fmt.Errorf("init combo inventory for warehouse %d: %w", warehouseID, err)
	}

	// Seed accessory_inventory for all accessories
	_, err = r.Pool.Exec(ctx, `
		INSERT INTO accessory_inventory (ma_phu_kien, warehouse_id, so_ton)
		SELECT ma_phu_kien, $1, 0 FROM accessories
		ON CONFLICT (ma_phu_kien, warehouse_id) DO NOTHING`, warehouseID)
	if err != nil {
		return fmt.Errorf("init accessory inventory for warehouse %d: %w", warehouseID, err)
	}

	return nil
}
