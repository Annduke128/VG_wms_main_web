package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"wms-v1/internal/domain"
	"wms-v1/internal/grid"
)

// QueryInventoryGrid executes the grid query and returns results.
// Wraps the builder query to COALESCE nullable columns (luong_ban_binh_quan_ngay, so_ngay_ton_ban).
func (r *PostgresRepo) QueryInventoryGrid(ctx context.Context, req domain.GridRequest) (*domain.GridResponse, error) {
	rawDataQuery, countQuery, args := grid.BuildQuery("inventory_grid", req)

	// Wrap to COALESCE nullable columns so pgx can scan into float64
	dataQuery := fmt.Sprintf(
		`SELECT ma_hang, ten_san_pham, so_ton, so_nhap, so_xuat,
		        tien_ton, tien_nhap, tien_xuat, so_ngay_ton,
		        COALESCE(luong_ban_binh_quan_ngay, 0) AS luong_ban_binh_quan_ngay,
		        COALESCE(so_ngay_ton_ban, 0) AS so_ngay_ton_ban,
		        don_gia, ma_bu, ma_nhom_hang
		 FROM (%s) _sub`, rawDataQuery)

	// Get total count
	var total int64
	if err := r.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count inventory: %w", err)
	}

	// Get rows
	rows, err := r.Pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query inventory: %w", err)
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.InventoryMain])
	if err != nil {
		return nil, fmt.Errorf("scan inventory rows: %w", err)
	}

	return &domain.GridResponse{
		RowsData:      items,
		TotalRowCount: total,
	}, nil
}

// UpdateInventoryItem updates a single inventory row
func (r *PostgresRepo) UpdateInventoryItem(ctx context.Context, maHang string, fields map[string]interface{}) error {
	// Build dynamic UPDATE
	setClauses := []string{}
	args := []interface{}{}
	idx := 1

	for col, val := range fields {
		if !grid.AllowedUpdateColumns[col] {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, idx))
		args = append(args, val)
		idx++
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no valid fields to update")
	}

	args = append(args, maHang)
	query := fmt.Sprintf("UPDATE inventory_main SET %s WHERE ma_hang = $%d",
		joinStrings(setClauses, ", "), idx)

	_, err := r.Pool.Exec(ctx, query, args...)
	return err
}

func joinStrings(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}
