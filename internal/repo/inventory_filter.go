package repo

import "context"

// FilterOptions holds distinct values for dropdown filters
type FilterOptions struct {
	MaBu       []string `json:"ma_bu"`
	MaNhomHang []string `json:"ma_nhom_hang"`
}

// GetInventoryFilterOptions returns distinct BU and nhom hang values for filter dropdowns
func (r *PostgresRepo) GetInventoryFilterOptions(ctx context.Context, warehouseID int64) (*FilterOptions, error) {
	opts := &FilterOptions{}

	buRows, err := r.Pool.Query(ctx,
		`SELECT DISTINCT p.ma_bu
		 FROM products p
		 JOIN inventory_main im ON im.ma_hang = p.ma_hang
		 WHERE p.ma_bu != '' AND im.warehouse_id = $1
		 ORDER BY p.ma_bu`, warehouseID)
	if err != nil {
		return nil, err
	}
	defer buRows.Close()
	for buRows.Next() {
		var v string
		if err := buRows.Scan(&v); err != nil {
			return nil, err
		}
		opts.MaBu = append(opts.MaBu, v)
	}

	nhomRows, err := r.Pool.Query(ctx,
		`SELECT DISTINCT p.ma_nhom_hang
		 FROM products p
		 JOIN inventory_main im ON im.ma_hang = p.ma_hang
		 WHERE p.ma_nhom_hang != '' AND im.warehouse_id = $1
		 ORDER BY p.ma_nhom_hang`, warehouseID)
	if err != nil {
		return nil, err
	}
	defer nhomRows.Close()
	for nhomRows.Next() {
		var v string
		if err := nhomRows.Scan(&v); err != nil {
			return nil, err
		}
		opts.MaNhomHang = append(opts.MaNhomHang, v)
	}

	return opts, nil
}
