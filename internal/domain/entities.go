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
	QuyCach     string    `json:"quy_cach" db:"quy_cach"`
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
}

type InboundItem struct {
	ID            int64     `json:"id" db:"id"`
	MaHang        string    `json:"ma_hang" db:"ma_hang"`
	TenSanPham    string    `json:"ten_san_pham" db:"ten_san_pham"`
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
}

// OutboundItem has same structure as InboundItem
type OutboundItem = InboundItem

type InventoryMovement struct {
	MovementID   int64     `json:"movement_id" db:"movement_id"`
	MaHang       string    `json:"ma_hang" db:"ma_hang"`
	Qty          float64   `json:"qty" db:"qty"`
	MovementType string    `json:"movement_type" db:"movement_type"` // IN or OUT
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type KanbanEvent struct {
	EventID   int64     `json:"event_id" db:"event_id"`
	SKU       string    `json:"sku" db:"sku"`
	FromStage string    `json:"from_stage" db:"from_stage"`
	ToStage   string    `json:"to_stage" db:"to_stage"`
	UserID    string    `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type KanbanInbound struct {
	ID         int64     `json:"id" db:"id"`
	MaHang     string    `json:"ma_hang" db:"ma_hang"`
	TenSanPham string    `json:"ten_san_pham" db:"ten_san_pham"`
	SoLuong    float64   `json:"so_luong" db:"so_luong"`
	Stage      string    `json:"stage" db:"stage"`
	Note       string    `json:"note" db:"note"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type KanbanOutbound struct {
	ID         int64     `json:"id" db:"id"`
	MaHang     string    `json:"ma_hang" db:"ma_hang"`
	TenSanPham string    `json:"ten_san_pham" db:"ten_san_pham"`
	SoLuong    float64   `json:"so_luong" db:"so_luong"`
	Stage      string    `json:"stage" db:"stage"`
	Note       string    `json:"note" db:"note"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
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
