package importer

import (
	"fmt"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"

	"wms-v1/internal/domain"
)

const dateFormat = "01/02/2006" // MM/DD/YYYY

// ParseProducts reads products from xlsx
func ParseProducts(filePath string) ([]domain.Product, []string, error) {
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

	var products []domain.Product
	var errors []string

	for i, row := range rows[1:] { // skip header
		if len(row) < 14 {
			errors = append(errors, fmt.Sprintf("row %d: insufficient columns (%d)", i+2, len(row)))
			continue
		}
		p := domain.Product{
			MaHang:      row[0],
			TenSanPham:  row[1],
			MaBu:        row[2],
			MaCat:       row[3],
			MaNhomHang:  row[4],
			NhomHang:    row[5],
			DonViTinh:   row[6],
			QuyCach:     row[7],
			DonGia:      parseFloat(row[8]),
			Vat:         parseFloat(row[9]),
			GiaNiv:      parseFloat(row[10]),
			GiaNhap:     parseFloat(row[11]),
			NgayCapNhat: parseDate(row[12]),
			HoaHong:     parseFloat(row[13]),
		}
		if p.MaHang == "" {
			errors = append(errors, fmt.Sprintf("row %d: empty ma_hang", i+2))
			continue
		}
		products = append(products, p)
	}

	return products, errors, nil
}

// ParseInventory reads inventory from xlsx
func ParseInventory(filePath string) ([]domain.InventoryMain, []string, error) {
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

	var items []domain.InventoryMain
	var errors []string

	for i, row := range rows[1:] {
		if len(row) < 10 {
			errors = append(errors, fmt.Sprintf("row %d: insufficient columns (%d)", i+2, len(row)))
			continue
		}
		item := domain.InventoryMain{
			MaHang:               row[0],
			TenSanPham:           row[1],
			SoTon:                parseFloat(row[2]),
			SoNhap:               parseFloat(row[3]),
			SoXuat:               parseFloat(row[4]),
			TienTon:              parseFloat(row[5]),
			TienNhap:             parseFloat(row[6]),
			TienXuat:             parseFloat(row[7]),
			SoNgayTon:            parseFloat(row[8]),
			LuongBanBinhQuanNgay: parseFloat(row[9]),
		}
		if item.MaHang == "" {
			errors = append(errors, fmt.Sprintf("row %d: empty ma_hang", i+2))
			continue
		}
		items = append(items, item)
	}

	return items, errors, nil
}

// ParseInbound reads inbound items from xlsx
func ParseInbound(filePath string) ([]domain.InboundItem, []string, error) {
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

	var items []domain.InboundItem
	var errors []string

	for i, row := range rows[1:] {
		if len(row) < 13 {
			errors = append(errors, fmt.Sprintf("row %d: insufficient columns (%d)", i+2, len(row)))
			continue
		}
		item := domain.InboundItem{
			MaHang:        row[0],
			TenSanPham:    row[1],
			DonViTinh:     row[2],
			QuyCach:       row[3],
			SoLuong:       parseFloat(row[4]),
			DoanhSo:       parseFloat(row[5]),
			ChietKhau:     parseFloat(row[6]),
			SoLuongTraLai: parseFloat(row[7]),
			DoanhThu:      parseFloat(row[8]),
			Von:           parseFloat(row[9]),
			LaiGop:        parseFloat(row[10]),
			TiLeLaiGop:    parseFloat(row[11]),
			NgayNhanHang:  parseDate(row[12]),
		}
		if item.MaHang == "" {
			errors = append(errors, fmt.Sprintf("row %d: empty ma_hang", i+2))
			continue
		}
		items = append(items, item)
	}

	return items, errors, nil
}

// ParseOutbound reads outbound items from xlsx (same structure as inbound)
func ParseOutbound(filePath string) ([]domain.InboundItem, []string, error) {
	return ParseInbound(filePath) // same structure
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseDate(s string) time.Time {
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		return time.Now()
	}
	return t
}
