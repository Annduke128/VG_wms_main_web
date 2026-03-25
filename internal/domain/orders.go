package domain

import "time"

// CreateInboundRequest — tạo đơn nhập kho
type CreateInboundRequest struct {
	MaHang       string  `json:"ma_hang"`
	TenSanPham   string  `json:"ten_san_pham"`
	DonViTinh    string  `json:"don_vi_tinh"`
	QuyCach      string  `json:"quy_cach"`
	SoLuong      float64 `json:"so_luong"`
	BatchCode    string  `json:"batch_code"`
	DoanhSo      float64 `json:"doanh_so"`
	ChietKhau    float64 `json:"chiet_khau"`
	DoanhThu     float64 `json:"doanh_thu"`
	Von          float64 `json:"von"`
	NgayNhanHang string  `json:"ngay_nhan_hang"` // RFC3339 or empty for NOW()
}

// CreateOutboundRequest — tạo đơn xuất kho (hệ thống tự FIFO)
type CreateOutboundRequest struct {
	MaHang     string  `json:"ma_hang"`
	TenSanPham string  `json:"ten_san_pham"`
	DonViTinh  string  `json:"don_vi_tinh"`
	QuyCach    string  `json:"quy_cach"`
	SoLuong    float64 `json:"so_luong"` // tổng số lượng cần xuất
	DoanhSo    float64 `json:"doanh_so"`
	ChietKhau  float64 `json:"chiet_khau"`
	DoanhThu   float64 `json:"doanh_thu"`
	Von        float64 `json:"von"`
}

// OutboundAllocation — kết quả FIFO allocation cho 1 lot
type OutboundAllocation struct {
	BatchCode    string  `json:"batch_code"`
	AllocatedQty float64 `json:"allocated_qty"`
	LotID        int64   `json:"lot_id"`
}

// InboundResult — kết quả sau khi tạo đơn nhập
type InboundResult struct {
	InboundItem InboundItem  `json:"inbound_item"`
	Lot         InventoryLot `json:"lot"`
}

// OutboundResult — kết quả sau khi tạo đơn xuất (có thể nhiều rows)
type OutboundResult struct {
	OutboundItems  []OutboundItem       `json:"outbound_items"`
	Allocations    []OutboundAllocation `json:"allocations"`
	TotalAllocated float64              `json:"total_allocated"`
}

// OrderFilter — filter params cho danh sách đơn hàng
type OrderFilter struct {
	OrderType  string    // "in", "out", or "" for all
	DateFrom   time.Time // zero = no lower bound
	DateTo     time.Time // zero = no upper bound
	MaBu       string    // filter by business unit (JOIN products)
	MaNhomHang string    // filter by product group (JOIN products)
	Limit      int
	Offset     int
}

// OrderListItem — row trong danh sách đơn hàng (UNION inbound + outbound)
type OrderListItem struct {
	ID           int64     `json:"id" db:"id"`
	OrderType    string    `json:"order_type" db:"order_type"` // "IN" or "OUT"
	MaHang       string    `json:"ma_hang" db:"ma_hang"`
	TenSanPham   string    `json:"ten_san_pham" db:"ten_san_pham"`
	DonViTinh    string    `json:"don_vi_tinh" db:"don_vi_tinh"`
	SoLuong      float64   `json:"so_luong" db:"so_luong"`
	BatchCode    string    `json:"batch_code" db:"batch_code"`
	DoanhSo      float64   `json:"doanh_so" db:"doanh_so"`
	DoanhThu     float64   `json:"doanh_thu" db:"doanh_thu"`
	NgayNhanHang time.Time `json:"ngay_nhan_hang" db:"ngay_nhan_hang"`
}
