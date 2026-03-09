package grid

import (
	"fmt"
	"strings"

	"wms-v1/internal/domain"
)

// AllowedColumns is the whitelist of columns that can be filtered/sorted
var AllowedColumns = map[string]bool{
	"ma_hang": true, "ten_san_pham": true, "ma_bu": true, "ma_cat": true,
	"ma_nhom_hang": true, "nhom_hang": true, "don_vi_tinh": true, "quy_cach": true,
	"don_gia": true, "vat": true, "gia_niv": true, "gia_nhap": true,
	"ngay_cap_nhat": true, "hoa_hong": true,
	"so_ton": true, "so_nhap": true, "so_xuat": true,
	"tien_ton": true, "tien_nhap": true, "tien_xuat": true,
	"so_ngay_ton": true, "luong_ban_binh_quan_ngay": true,
}

// BuildQuery builds a paginated, filtered, sorted SQL query from GridRequest.
// Returns: dataQuery, countQuery, args
func BuildQuery(table string, req domain.GridRequest) (string, string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	// Build WHERE from filterModel
	for col, filter := range req.FilterModel {
		if !AllowedColumns[col] {
			continue
		}
		cond, newArgs, newIdx := buildFilterCondition(col, filter, argIdx)
		if cond != "" {
			conditions = append(conditions, cond)
			args = append(args, newArgs...)
			argIdx = newIdx
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Build ORDER BY from sortModel
	orderClause := ""
	if len(req.SortModel) > 0 {
		var orderParts []string
		for _, s := range req.SortModel {
			if !AllowedColumns[s.ColID] {
				continue
			}
			dir := "ASC"
			if strings.ToLower(s.Sort) == "desc" {
				dir = "DESC"
			}
			orderParts = append(orderParts, fmt.Sprintf("%s %s", s.ColID, dir))
		}
		if len(orderParts) > 0 {
			orderClause = "ORDER BY " + strings.Join(orderParts, ", ")
		}
	}

	// Build LIMIT/OFFSET
	limit := req.EndRow - req.StartRow
	if limit <= 0 {
		limit = 100
	}
	if limit > 5000 {
		limit = 5000
	}
	offset := req.StartRow
	if offset < 0 {
		offset = 0
	}

	dataQuery := fmt.Sprintf("SELECT * FROM %s %s %s LIMIT %d OFFSET %d",
		table, whereClause, orderClause, limit, offset)

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", table, whereClause)

	return dataQuery, countQuery, args
}

func buildFilterCondition(col string, filter domain.FilterItem, argIdx int) (string, []interface{}, int) {
	switch filter.Type {
	case "contains":
		v := fmt.Sprintf("%%%v%%", filter.Filter)
		return fmt.Sprintf("%s ILIKE $%d", col, argIdx), []interface{}{v}, argIdx + 1

	case "equals":
		return fmt.Sprintf("%s = $%d", col, argIdx), []interface{}{filter.Filter}, argIdx + 1

	case "startsWith":
		v := fmt.Sprintf("%v%%", filter.Filter)
		return fmt.Sprintf("%s ILIKE $%d", col, argIdx), []interface{}{v}, argIdx + 1

	case "endsWith":
		v := fmt.Sprintf("%%%v", filter.Filter)
		return fmt.Sprintf("%s ILIKE $%d", col, argIdx), []interface{}{v}, argIdx + 1

	case "inRange":
		return fmt.Sprintf("%s BETWEEN $%d AND $%d", col, argIdx, argIdx+1),
			[]interface{}{filter.Filter, filter.FilterTo}, argIdx + 2

	case "set":
		if len(filter.Values) == 0 {
			return "", nil, argIdx
		}
		placeholders := make([]string, len(filter.Values))
		vals := make([]interface{}, len(filter.Values))
		for i, v := range filter.Values {
			placeholders[i] = fmt.Sprintf("$%d", argIdx+i)
			vals[i] = v
		}
		return fmt.Sprintf("%s IN (%s)", col, strings.Join(placeholders, ",")),
			vals, argIdx + len(filter.Values)

	default:
		return "", nil, argIdx
	}
}
