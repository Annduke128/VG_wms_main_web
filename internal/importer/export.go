package importer

import (
	"fmt"

	"github.com/xuri/excelize/v2"

	"wms-v1/internal/domain"
)

// ExportColumnHeaders maps column ID to Vietnamese header
var ExportColumnHeaders = map[string]string{
	"ma_hang":                  "Mã hàng",
	"ten_san_pham":             "Tên sản phẩm",
	"ma_bu":                    "Mã BU",
	"ma_nhom_hang":             "Mã nhóm hàng",
	"don_gia":                  "Đơn giá",
	"so_ton":                   "Số tồn",
	"so_nhap":                  "Số nhập",
	"so_xuat":                  "Số xuất",
	"tien_ton":                 "Tiền tồn",
	"tien_nhap":                "Tiền nhập",
	"tien_xuat":                "Tiền xuất",
	"so_ngay_ton":              "Số ngày tồn",
	"luong_ban_binh_quan_ngay": "LBBQ/ngày",
	"so_ngay_ton_ban":          "Ngày tồn bán",
}

// ValidExportColumns is the whitelist of columns allowed for export
var ValidExportColumns = map[string]bool{
	"ma_hang": true, "ten_san_pham": true, "ma_bu": true, "ma_nhom_hang": true,
	"don_gia": true, "so_ton": true, "so_nhap": true, "so_xuat": true,
	"tien_ton": true, "tien_nhap": true, "tien_xuat": true,
	"so_ngay_ton": true, "luong_ban_binh_quan_ngay": true, "so_ngay_ton_ban": true,
}

// BuildExportExcel creates an Excel file with selected columns from inventory data
func BuildExportExcel(items []domain.InventoryMain, columns []string) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "BaoCaoTonKho"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, fmt.Errorf("set sheet name: %w", err)
	}

	// Write headers
	for i, col := range columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		header := ExportColumnHeaders[col]
		if header == "" {
			header = col
		}
		if err := f.SetCellValue(sheet, cell, header); err != nil {
			return nil, fmt.Errorf("set header: %w", err)
		}
	}

	// Write data rows
	for rowIdx, item := range items {
		for colIdx, col := range columns {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			val := getFieldValue(item, col)
			if err := f.SetCellValue(sheet, cell, val); err != nil {
				return nil, fmt.Errorf("set cell: %w", err)
			}
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("write excel: %w", err)
	}
	return buf.Bytes(), nil
}

func getFieldValue(item domain.InventoryMain, col string) interface{} {
	switch col {
	case "ma_hang":
		return item.MaHang
	case "ten_san_pham":
		return item.TenSanPham
	case "ma_bu":
		return item.MaBu
	case "ma_nhom_hang":
		return item.MaNhomHang
	case "don_gia":
		return item.DonGia
	case "so_ton":
		return item.SoTon
	case "so_nhap":
		return item.SoNhap
	case "so_xuat":
		return item.SoXuat
	case "tien_ton":
		return item.TienTon
	case "tien_nhap":
		return item.TienNhap
	case "tien_xuat":
		return item.TienXuat
	case "so_ngay_ton":
		return item.SoNgayTon
	case "luong_ban_binh_quan_ngay":
		return item.LuongBanBinhQuanNgay
	case "so_ngay_ton_ban":
		return item.SoNgayTonBan
	default:
		return ""
	}
}
