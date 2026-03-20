package importer

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// inventoryHeaders maps column index → Vietnamese header name.
// 17 columns: products (12) + lot (2) + inventory (3).
// Order MUST match ParseInventoryFull in inventory_full.go.
var inventoryHeaders = []string{
	// Products (1–12)
	"Mã vạch",
	"Tên sản phẩm",
	"BU",
	"Mã cat",
	"Mã nhóm hàng",
	"Nhóm hàng",
	"ĐVT",
	"Quy cách",
	"Đơn giá",
	"VAT",
	"Ngày cập nhật",
	"Hoa hồng",
	// Lot (13–14)
	"Mã lô hàng",
	"Ngày nhập",
	// Inventory (15–17)
	"Số tồn",
	"Số nhập",
	"Số xuất",
}

// BuildInventoryTemplate generates an empty .xlsx with 17-column header row.
func BuildInventoryTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"

	// Write header row
	for i, h := range inventoryHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return nil, fmt.Errorf("set header %d: %w", i, err)
		}
	}

	// Column widths
	widths := []float64{
		16, 30, 10, 10, 14, 16, 8, 12, // products 1–8
		12, 8, 14, 10, // products 9–12
		14, 14, // lot
		10, 10, 10, // inventory
	}
	for i, w := range widths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		if err := f.SetColWidth(sheet, col, col, w); err != nil {
			return nil, fmt.Errorf("set col width %d: %w", i, err)
		}
	}

	// Bold header style
	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	if err == nil {
		startCell, _ := excelize.CoordinatesToCellName(1, 1)
		endCell, _ := excelize.CoordinatesToCellName(len(inventoryHeaders), 1)
		_ = f.SetCellStyle(sheet, startCell, endCell, style)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("write xlsx buffer: %w", err)
	}

	return buf.Bytes(), nil
}
