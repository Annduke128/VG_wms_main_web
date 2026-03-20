package importer

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"

	"wms-v1/internal/domain"
)

const dateFmtDDMMYYYY = "02/01/2006" // dd/mm/yyyy

// InventoryFullRow holds parsed data from a single row of the 22-column file.
type InventoryFullRow struct {
	Product   domain.Product
	Inventory domain.InventoryMain
	Inbound   domain.InboundItem
}

// ParseInventoryFull reads the 22-column inventory file.
// Returns parsed rows + parse errors (best-effort: each row independent).
//
// Column order (0-indexed):
//
//	0  Mã vạch        → products.ma_hang
//	1  Tên sản phẩm   → products.ten_san_pham
//	2  BU              → products.ma_bu
//	3  Mã cat          → products.ma_cat
//	4  Mã nhóm hàng   → products.ma_nhom_hang
//	5  Nhóm hàng       → products.nhom_hang
//	6  ĐVT             → products.don_vi_tinh
//	7  Quy cách        → products.quy_cach
//	8  Đơn giá bán     → products.don_gia
//	9  VAT             → products.vat
//	10 Giá NIV         → products.gia_niv
//	11 Đơn giá nhập    → products.gia_nhap
//	12 Ngày cập nhật   → IGNORED (computed as max ngày nhập per mã vạch)
//	13 Hoa hồng        → products.hoa_hong
//	14 Mã lô hàng      → inbound_items.batch_code (REQUIRED)
//	15 Ngày nhập        → inbound_items.ngay_nhan_hang (REQUIRED, dd/mm/yyyy)
//	16 Số tồn           → inventory_main.so_ton
//	17 Số nhập          → inventory_main.so_nhap
//	18 Số xuất          → inventory_main.so_xuat
//	19 Tiền tồn         → inventory_main.tien_ton
//	20 Tiền nhập        → inventory_main.tien_nhap
//	21 Tiền xuất        → inventory_main.tien_xuat
func ParseInventoryFull(filePath string) ([]InventoryFullRow, []string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, nil, fmt.Errorf("get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, nil, fmt.Errorf("no data rows found")
	}

	var results []InventoryFullRow
	var parseErrors []string

	// Track max ngày nhập per mã vạch for ngay_cap_nhat
	maxDates := make(map[string]time.Time)

	for i, row := range rows[1:] { // skip header
		rowNum := i + 2

		if len(row) < 22 {
			// Pad with empty strings
			for len(row) < 22 {
				row = append(row, "")
			}
		}

		maVach := row[0]
		if maVach == "" {
			parseErrors = append(parseErrors, fmt.Sprintf("row %d: mã vạch trống", rowNum))
			continue
		}

		batchCode := row[14]
		if batchCode == "" {
			parseErrors = append(parseErrors, fmt.Sprintf("row %d: mã lô hàng trống", rowNum))
			continue
		}

		ngayNhapStr := row[15]
		if ngayNhapStr == "" {
			parseErrors = append(parseErrors, fmt.Sprintf("row %d: ngày nhập trống", rowNum))
			continue
		}

		ngayNhap, err := time.Parse(dateFmtDDMMYYYY, ngayNhapStr)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("row %d: ngày nhập không hợp lệ (%s), dùng dd/mm/yyyy", rowNum, ngayNhapStr))
			continue
		}

		tenSanPham := row[1]

		// Track max date per mã vạch
		if cur, ok := maxDates[maVach]; !ok || ngayNhap.After(cur) {
			maxDates[maVach] = ngayNhap
		}

		product := domain.Product{
			MaHang:     maVach,
			TenSanPham: tenSanPham,
			MaBu:       row[2],
			MaCat:      row[3],
			MaNhomHang: row[4],
			NhomHang:   row[5],
			DonViTinh:  row[6],
			QuyCach:    row[7],
			DonGia:     parseFloat(row[8]),
			Vat:        parseFloat(row[9]),
			GiaNiv:     parseFloat(row[10]),
			GiaNhap:    parseFloat(row[11]),
			// NgayCapNhat will be set after all rows are parsed
			HoaHong: parseFloat(row[13]),
		}

		inventory := domain.InventoryMain{
			MaHang:     maVach,
			TenSanPham: tenSanPham,
			SoTon:      parseFloat(row[16]),
			SoNhap:     parseFloat(row[17]),
			SoXuat:     parseFloat(row[18]),
			TienTon:    parseFloat(row[19]),
			TienNhap:   parseFloat(row[20]),
			TienXuat:   parseFloat(row[21]),
		}

		inbound := domain.InboundItem{
			MaHang:       maVach,
			TenSanPham:   tenSanPham,
			DonViTinh:    row[6],
			QuyCach:      row[7],
			SoLuong:      parseFloat(row[17]), // Số nhập = lot qty
			BatchCode:    batchCode,
			NgayNhanHang: ngayNhap,
		}

		results = append(results, InventoryFullRow{
			Product:   product,
			Inventory: inventory,
			Inbound:   inbound,
		})
	}

	// Set NgayCapNhat = max(ngày nhập) per mã vạch
	for i := range results {
		maVach := results[i].Product.MaHang
		if maxDate, ok := maxDates[maVach]; ok {
			results[i].Product.NgayCapNhat = maxDate
		}
	}

	return results, parseErrors, nil
}
