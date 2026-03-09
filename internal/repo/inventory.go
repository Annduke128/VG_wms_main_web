package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"wms-v1/internal/domain"
	"wms-v1/internal/grid"
)

// QueryInventoryGrid executes the grid query and returns results
func (r *PostgresRepo) QueryInventoryGrid(ctx context.Context, req domain.GridRequest) (*domain.GridResponse, error) {
	dataQuery, countQuery, args := grid.BuildQuery("inventory_main", req)

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
		if !grid.AllowedColumns[col] {
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
