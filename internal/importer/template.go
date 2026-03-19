package importer

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// inventoryHeaders maps column index → Vietnamese header name.
// Order MUST match ParseInventory in xlsx.go (row[0]–row[9]).
var inventoryHeaders = []string{
	"Mã hàng",
	"Tên sản phẩm",
	"Số tồn",
	"Số nhập",
	"Số xuất",
	"Tiền tồn",
	"Tiền nhập",
	"Tiền xuất",
	"Số ngày tồn",
	"Lượng bán bình quân ngày",
}

// BuildInventoryTemplate generates an empty .xlsx with header row only.
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

	// Auto-width: set reasonable column widths
	widths := []float64{14, 30, 10, 10, 10, 12, 12, 12, 12, 24}
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
