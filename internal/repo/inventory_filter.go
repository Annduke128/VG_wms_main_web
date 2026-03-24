package repo

import "context"

// FilterOptions holds distinct values for dropdown filters
type FilterOptions struct {
	MaBu       []string `json:"ma_bu"`
	MaNhomHang []string `json:"ma_nhom_hang"`
}

// GetInventoryFilterOptions returns distinct BU and nhom hang values for filter dropdowns
func (r *PostgresRepo) GetInventoryFilterOptions(ctx context.Context) (*FilterOptions, error) {
	opts := &FilterOptions{}

	buRows, err := r.Pool.Query(ctx,
		"SELECT DISTINCT ma_bu FROM products WHERE ma_bu != '' ORDER BY ma_bu")
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
		"SELECT DISTINCT ma_nhom_hang FROM products WHERE ma_nhom_hang != '' ORDER BY ma_nhom_hang")
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
