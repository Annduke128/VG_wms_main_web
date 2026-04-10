package domain

import "time"

type Product struct {
	MaHang      string    `json:"ma_hang" db:"ma_hang"`
	TenSanPham  string    `json:"ten_san_pham" db:"ten_san_pham"`
	MaBu        string    `json:"ma_bu" db:"ma_bu"`
	MaCat       string    `json:"ma_cat" db:"ma_cat"`
	MaNhomHang  string    `json:"ma_nhom_hang" db:"ma_nhom_hang"`
	NhomHang    string    `json:"nhom_hang" db:"nhom_hang"`
	DonViTinh   string    `json:"don_vi_tinh" db:"don_vi_tinh"`
	QuyCach     float64   `json:"quy_cach" db:"quy_cach"`
	DonGia      float64   `json:"don_gia" db:"don_gia"`
	Vat         float64   `json:"vat" db:"vat"`
	GiaNiv      float64   `json:"gia_niv" db:"gia_niv"`
	GiaNhap     float64   `json:"gia_nhap" db:"gia_nhap"`
	NgayCapNhat time.Time `json:"ngay_cap_nhat" db:"ngay_cap_nhat"`
	HoaHong     float64   `json:"hoa_hong" db:"hoa_hong"`
}

type InventoryMain struct {
	MaHang               string  `json:"ma_hang" db:"ma_hang"`
	TenSanPham           string  `json:"ten_san_pham" db:"ten_san_pham"`
	SoTon                float64 `json:"so_ton" db:"so_ton"`
	SoNhap               float64 `json:"so_nhap" db:"so_nhap"`
	SoXuat               float64 `json:"so_xuat" db:"so_xuat"`
	TienTon              float64 `json:"tien_ton" db:"tien_ton"`
	TienNhap             float64 `json:"tien_nhap" db:"tien_nhap"`
	TienXuat             float64 `json:"tien_xuat" db:"tien_xuat"`
	SoNgayTon            float64 `json:"so_ngay_ton" db:"so_ngay_ton"`
	LuongBanBinhQuanNgay float64 `json:"luong_ban_binh_quan_ngay" db:"luong_ban_binh_quan_ngay"`
	SoNgayTonBan         float64 `json:"so_ngay_ton_ban" db:"so_ngay_ton_ban"`
	WarehouseID          int64   `json:"warehouse_id" db:"warehouse_id"`
	// Joined from products table via inventory_grid view
	DonGia     float64 `json:"don_gia" db:"don_gia"`
	MaBu       string  `json:"ma_bu" db:"ma_bu"`
	MaNhomHang string  `json:"ma_nhom_hang" db:"ma_nhom_hang"`
}

type InboundItem struct {
	ID            int64     `json:"id" db:"id"`
	MaHang        string    `json:"ma_hang" db:"ma_hang"`
	TenSanPham    string    `json:"ten_san_pham" db:"ten_san_pham"`
	WarehouseID   int64     `json:"warehouse_id" db:"warehouse_id"`
	DonViTinh     string    `json:"don_vi_tinh" db:"don_vi_tinh"`
	QuyCach       string    `json:"quy_cach" db:"quy_cach"`
	SoLuong       float64   `json:"so_luong" db:"so_luong"`
	DoanhSo       float64   `json:"doanh_so" db:"doanh_so"`
	ChietKhau     float64   `json:"chiet_khau" db:"chiet_khau"`
	SoLuongTraLai float64   `json:"so_luong_tra_lai" db:"so_luong_tra_lai"`
	DoanhThu      float64   `json:"doanh_thu" db:"doanh_thu"`
	Von           float64   `json:"von" db:"von"`
	LaiGop        float64   `json:"lai_gop" db:"lai_gop"`
	TiLeLaiGop    float64   `json:"ti_le_lai_gop" db:"ti_le_lai_gop"`
	NgayNhanHang  time.Time `json:"ngay_nhan_hang" db:"ngay_nhan_hang"`
	BatchCode     string    `json:"batch_code" db:"batch_code"`
}

type OutboundItem struct {
	ID            int64     `json:"id" db:"id"`
	MaHang        string    `json:"ma_hang" db:"ma_hang"`
	TenSanPham    string    `json:"ten_san_pham" db:"ten_san_pham"`
	WarehouseID   int64     `json:"warehouse_id" db:"warehouse_id"`
	DonViTinh     string    `json:"don_vi_tinh" db:"don_vi_tinh"`
	QuyCach       string    `json:"quy_cach" db:"quy_cach"`
	SoLuong       float64   `json:"so_luong" db:"so_luong"`
	DoanhSo       float64   `json:"doanh_so" db:"doanh_so"`
	ChietKhau     float64   `json:"chiet_khau" db:"chiet_khau"`
	SoLuongTraLai float64   `json:"so_luong_tra_lai" db:"so_luong_tra_lai"`
	DoanhThu      float64   `json:"doanh_thu" db:"doanh_thu"`
	Von           float64   `json:"von" db:"von"`
	LaiGop        float64   `json:"lai_gop" db:"lai_gop"`
	TiLeLaiGop    float64   `json:"ti_le_lai_gop" db:"ti_le_lai_gop"`
	NgayNhanHang  time.Time `json:"ngay_nhan_hang" db:"ngay_nhan_hang"`
	BatchCode     string    `json:"batch_code" db:"batch_code"`
}

// InventoryLot tracks FIFO batch quantities
type InventoryLot struct {
	ID           int64     `json:"id" db:"id"`
	MaHang       string    `json:"ma_hang" db:"ma_hang"`
	BatchCode    string    `json:"batch_code" db:"batch_code"`
	WarehouseID  int64     `json:"warehouse_id" db:"warehouse_id"`
	ReceivedAt   time.Time `json:"received_at" db:"received_at"`
	QtyIn        float64   `json:"qty_in" db:"qty_in"`
	QtyOut       float64   `json:"qty_out" db:"qty_out"`
	QtyRemaining float64   `json:"qty_remaining" db:"qty_remaining"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// InventoryThreshold stores threshold config with version history
type InventoryThreshold struct {
	ID            int64      `json:"id" db:"id"`
	MaHang        string     `json:"ma_hang" db:"ma_hang"`
	WarehouseID   int64      `json:"warehouse_id" db:"warehouse_id"`
	MinQty        float64    `json:"min_qty" db:"min_qty"`
	OptimalQty    float64    `json:"optimal_qty" db:"optimal_qty"`
	MaxAgeDays    int        `json:"max_age_days" db:"max_age_days"`
	Source        string     `json:"source" db:"source"`
	ModelVersion  string     `json:"model_version" db:"model_version"`
	EffectiveFrom time.Time  `json:"effective_from" db:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to" db:"effective_to"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

type InventoryMovement struct {
	MovementID   int64     `json:"movement_id" db:"movement_id"`
	MaHang       string    `json:"ma_hang" db:"ma_hang"`
	WarehouseID  int64     `json:"warehouse_id" db:"warehouse_id"`
	Qty          float64   `json:"qty" db:"qty"`
	MovementType string    `json:"movement_type" db:"movement_type"` // IN or OUT
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type AsyncJob struct {
	JobID     string    `json:"job_id" db:"job_id"`
	JobType   string    `json:"job_type" db:"job_type"`
	Status    string    `json:"status" db:"status"`
	Payload   string    `json:"payload" db:"payload"`
	Result    string    `json:"result" db:"result"`
	Error     string    `json:"error" db:"error"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Warehouse represents a physical warehouse location
type Warehouse struct {
	ID        int64     `json:"id" db:"id"`
	Code      string    `json:"code" db:"code"`
	Name      string    `json:"name" db:"name"`
	Address   string    `json:"address" db:"address"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateWarehouseRequest is the request body for creating a warehouse
type CreateWarehouseRequest struct {
	Name    string `json:"name" binding:"required"`
	Address string `json:"address"`
}

// UpdateWarehouseRequest is the request body for updating a warehouse
type UpdateWarehouseRequest struct {
	Name    *string `json:"name"`
	Address *string `json:"address"`
}

type ImportBatch struct {
	BatchID     int64      `json:"batch_id" db:"batch_id"`
	FileType    string     `json:"file_type" db:"file_type"`
	FileName    string     `json:"file_name" db:"file_name"`
	TotalRows   int        `json:"total_rows" db:"total_rows"`
	SuccessRows int        `json:"success_rows" db:"success_rows"`
	ErrorRows   int        `json:"error_rows" db:"error_rows"`
	Status      string     `json:"status" db:"status"`
	Errors      string     `json:"errors" db:"errors"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
}
