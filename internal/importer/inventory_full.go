package importer

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"

	"wms-v1/internal/domain"
)

// dateFmtDDMMYYYY kept for reference; actual parsing uses parseDateFlexible.
const dateFmtDDMMYYYY = "02/01/2006" // dd/mm/yyyy

// InventoryFullRow holds parsed data from a single row of the 15-column file.
type InventoryFullRow struct {
	Product   domain.Product
	Inventory domain.InventoryMain
	Inbound   domain.InboundItem
}

// ParseInventoryFull reads the 15-column inventory file.
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
//	7  Quy cách        → products.quy_cach (NUMERIC)
//	8  Đơn giá         → products.don_gia
//	9  VAT             → products.vat
//	10 Ngày cập nhật   → products.ngay_cap_nhat (if empty → ngày nhập)
//	11 Hoa hồng        → products.hoa_hong
//	12 Mã lô hàng      → inbound_items.batch_code (REQUIRED)
//	13 Ngày nhập        → inbound_items.ngay_nhan_hang (REQUIRED, dd/mm/yyyy)
//	14 Số tồn           → inventory_main.so_ton
//
// Số nhập/Số xuất removed — these are computed from inbound/outbound orders.
//
// Computed fields:
//
//	products.gia_nhap = don_gia * quy_cach * hoa_hong
//	products.gia_niv  = gia_nhap / (1 + VAT)
//	inventory_main.tien_ton  = don_gia * so_ton
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

	// Reject old 17-column files — new format requires exactly 15 columns
	headerCount := len(rows[0])
	if headerCount > 15 {
		return nil, nil, fmt.Errorf("file có %d cột (yêu cầu 15 cột). Vui lòng tải mẫu mới — Số nhập/Số xuất đã bị bỏ", headerCount)
	}

	var results []InventoryFullRow
	var parseErrors []string

	for i, row := range rows[1:] { // skip header
		rowNum := i + 2

		if len(row) < 15 {
			// Pad with empty strings
			for len(row) < 15 {
				row = append(row, "")
			}
		}

		maVach := row[0]
		if maVach == "" {
			parseErrors = append(parseErrors, fmt.Sprintf("row %d: mã vạch trống", rowNum))
			continue
		}

		// --- Ngày nhập (flexible parsing) ---
		ngayNhapStr := row[13]
		var ngayNhap time.Time
		if ngayNhapStr == "" {
			parseErrors = append(parseErrors, fmt.Sprintf("row %d: ngày nhập trống", rowNum))
			continue
		}
		parsedDate, _, dateErr := parseDateFlexible(ngayNhapStr)
		if dateErr != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("row %d: ngày nhập không hợp lệ (%s), dùng dd/mm/yyyy hoặc dd-mm-yyyy hoặc dd-mm-yy", rowNum, ngayNhapStr))
			continue
		}
		ngayNhap = parsedDate

		// --- Mã lô hàng: auto-generate if empty ---
		batchCode := row[12]
		if batchCode == "" {
			batchCode = fmt.Sprintf("LOT-%s-%s", maVach, ngayNhap.Format("20060102"))
		}

		tenSanPham := row[1]
		donGia := parseFloat(row[8])
		quyCach := parseFloat(row[7])
		vat := parseFloat(row[9])
		hoaHong := parseFloat(row[11])
		soTon := parseFloat(row[14])

		// Computed: gia_nhap = don_gia * quy_cach * hoa_hong
		giaNhap := donGia * quyCach * hoaHong
		// Computed: gia_niv = gia_nhap / (1 + VAT)
		giaNiv := 0.0
		if 1+vat > 0 {
			giaNiv = giaNhap / (1 + vat)
		}

		// Ngày cập nhật: if file has value → use it; if empty → use ngày nhập
		ngayCapNhat := ngayNhap
		if row[10] != "" {
			parsed, _, err := parseDateFlexible(row[10])
			if err == nil {
				ngayCapNhat = parsed
			}
		}

		product := domain.Product{
			MaHang:      maVach,
			TenSanPham:  tenSanPham,
			MaBu:        row[2],
			MaCat:       row[3],
			MaNhomHang:  row[4],
			NhomHang:    row[5],
			DonViTinh:   row[6],
			QuyCach:     quyCach,
			DonGia:      donGia,
			Vat:         vat,
			GiaNiv:      giaNiv,
			GiaNhap:     giaNhap,
			NgayCapNhat: ngayCapNhat,
			HoaHong:     hoaHong,
		}

		inventory := domain.InventoryMain{
			MaHang:     maVach,
			TenSanPham: tenSanPham,
			SoTon:      soTon,
			TienTon:    donGia * soTon,
		}

		inbound := domain.InboundItem{
			MaHang:       maVach,
			TenSanPham:   tenSanPham,
			DonViTinh:    row[6],
			QuyCach:      row[7],
			SoLuong:      soTon, // Use soTon as the inbound lot qty
			BatchCode:    batchCode,
			NgayNhanHang: ngayNhap,
		}

		results = append(results, InventoryFullRow{
			Product:   product,
			Inventory: inventory,
			Inbound:   inbound,
		})
	}

	return results, parseErrors, nil
}
