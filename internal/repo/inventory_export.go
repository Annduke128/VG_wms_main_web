package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"wms-v1/internal/domain"
	"wms-v1/internal/grid"
)

// QueryExportRows returns inventory rows for export.
// If maHangs is non-empty, exports those specific rows.
// If maHangs is empty, exports all rows matching filterModel (max 10000).
func (r *PostgresRepo) QueryExportRows(ctx context.Context, maHangs []string, filterModel map[string]domain.FilterItem) ([]domain.InventoryMain, error) {
	var query string
	var args []interface{}

	if len(maHangs) > 0 {
		placeholders := make([]string, len(maHangs))
		args = make([]interface{}, len(maHangs))
		for i, mh := range maHangs {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = mh
		}
		query = fmt.Sprintf(`
			SELECT ma_hang, ten_san_pham, so_ton, so_nhap, so_xuat,
			       tien_ton, tien_nhap, tien_xuat, so_ngay_ton,
			       COALESCE(luong_ban_binh_quan_ngay, 0) AS luong_ban_binh_quan_ngay,
			       COALESCE(so_ngay_ton_ban, 0) AS so_ngay_ton_ban,
			       don_gia, ma_bu, ma_nhom_hang
			FROM inventory_grid
			WHERE ma_hang IN (%s)
			ORDER BY ma_hang`, strings.Join(placeholders, ","))
	} else {
		req := domain.GridRequest{
			StartRow:    0,
			EndRow:      10000,
			FilterModel: filterModel,
		}
		rawDataQuery, _, rawArgs := grid.BuildQuery("inventory_grid", req)
		query = fmt.Sprintf(`
			SELECT ma_hang, ten_san_pham, so_ton, so_nhap, so_xuat,
			       tien_ton, tien_nhap, tien_xuat, so_ngay_ton,
			       COALESCE(luong_ban_binh_quan_ngay, 0) AS luong_ban_binh_quan_ngay,
			       COALESCE(so_ngay_ton_ban, 0) AS so_ngay_ton_ban,
			       don_gia, ma_bu, ma_nhom_hang
			FROM (%s) _sub`, rawDataQuery)
		args = rawArgs
	}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query export rows: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.InventoryMain])
}
