package domain

import "time"

// --- Combo Master ---

type ComboMaster struct {
	MaCombo   string    `json:"ma_combo" db:"ma_combo"`
	TenCombo  string    `json:"ten_combo" db:"ten_combo"`
	MoTa      string    `json:"mo_ta" db:"mo_ta"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type ComboBOMSemi struct {
	ID      int64   `json:"id" db:"id"`
	MaCombo string  `json:"ma_combo" db:"ma_combo"`
	MaHang  string  `json:"ma_hang" db:"ma_hang"`
	SoLuong float64 `json:"so_luong" db:"so_luong"`
	// Joined fields for display
	TenSanPham string `json:"ten_san_pham,omitempty" db:"ten_san_pham"`
}

type ComboBOMAccessory struct {
	ID        int64   `json:"id" db:"id"`
	MaCombo   string  `json:"ma_combo" db:"ma_combo"`
	MaPhuKien string  `json:"ma_phu_kien" db:"ma_phu_kien"`
	SoLuong   float64 `json:"so_luong" db:"so_luong"`
	// Joined fields for display
	TenPhuKien string `json:"ten_phu_kien,omitempty" db:"ten_phu_kien"`
}

// ComboDetail is a combo master with its full BOM
type ComboDetail struct {
	ComboMaster
	BOMSemi      []ComboBOMSemi      `json:"bom_semi"`
	BOMAccessory []ComboBOMAccessory `json:"bom_accessory"`
}

// --- Combo Inventory ---

type ComboInventory struct {
	MaCombo     string    `json:"ma_combo" db:"ma_combo"`
	WarehouseID int64     `json:"warehouse_id" db:"warehouse_id"`
	SoTon       float64   `json:"so_ton" db:"so_ton"`
	SoNhap      float64   `json:"so_nhap" db:"so_nhap"`
	SoXuat      float64   `json:"so_xuat" db:"so_xuat"`
	SoTra       float64   `json:"so_tra" db:"so_tra"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	// Joined for display
	TenCombo string `json:"ten_combo,omitempty" db:"ten_combo"`
}

// --- Combo Transactions ---

type ComboTransaction struct {
	ID              int64     `json:"id" db:"id"`
	MaCombo         string    `json:"ma_combo" db:"ma_combo"`
	WarehouseID     int64     `json:"warehouse_id" db:"warehouse_id"`
	TransactionType string    `json:"transaction_type" db:"transaction_type"` // CREATE, CANCEL, OUT, RETURN
	SoLuong         float64   `json:"so_luong" db:"so_luong"`
	Note            string    `json:"note" db:"note"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	// Joined for display
	TenCombo string `json:"ten_combo,omitempty" db:"ten_combo"`
}

type ComboComponentMovement struct {
	ID                 int64   `json:"id" db:"id"`
	ComboTransactionID int64   `json:"combo_transaction_id" db:"combo_transaction_id"`
	WarehouseID        int64   `json:"warehouse_id" db:"warehouse_id"`
	ComponentType      string  `json:"component_type" db:"component_type"` // SEMI, ACCESSORY
	MaComponent        string  `json:"ma_component" db:"ma_component"`
	SoLuong            float64 `json:"so_luong" db:"so_luong"`
	LotID              *int64  `json:"lot_id,omitempty" db:"lot_id"`
}

// --- Accessories ---

type Accessory struct {
	MaPhuKien  string    `json:"ma_phu_kien" db:"ma_phu_kien"`
	TenPhuKien string    `json:"ten_phu_kien" db:"ten_phu_kien"`
	DonViTinh  string    `json:"don_vi_tinh" db:"don_vi_tinh"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type AccessoryInventory struct {
	MaPhuKien   string    `json:"ma_phu_kien" db:"ma_phu_kien"`
	WarehouseID int64     `json:"warehouse_id" db:"warehouse_id"`
	SoTon       float64   `json:"so_ton" db:"so_ton"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	// Joined for display
	TenPhuKien string `json:"ten_phu_kien,omitempty" db:"ten_phu_kien"`
	DonViTinh  string `json:"don_vi_tinh,omitempty" db:"don_vi_tinh"`
}

type AccessoryMovement struct {
	ID           int64     `json:"id" db:"id"`
	MaPhuKien    string    `json:"ma_phu_kien" db:"ma_phu_kien"`
	WarehouseID  int64     `json:"warehouse_id" db:"warehouse_id"`
	MovementType string    `json:"movement_type" db:"movement_type"` // IN, OUT, RETURN
	SoLuong      float64   `json:"so_luong" db:"so_luong"`
	Note         string    `json:"note" db:"note"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// --- Request types ---

type CreateComboRequest struct {
	MaCombo     string  `json:"ma_combo"`
	WarehouseID int64   `json:"warehouse_id" binding:"required"`
	SoLuong     float64 `json:"so_luong"`
	Note        string  `json:"note"`
}

type CancelComboRequest struct {
	MaCombo     string  `json:"ma_combo"`
	WarehouseID int64   `json:"warehouse_id" binding:"required"`
	SoLuong     float64 `json:"so_luong"`
	Note        string  `json:"note"`
}

type ComboOutRequest struct {
	MaCombo     string  `json:"ma_combo"`
	WarehouseID int64   `json:"warehouse_id" binding:"required"`
	SoLuong     float64 `json:"so_luong"`
	Note        string  `json:"note"`
}

type ComboReturnRequest struct {
	MaCombo     string  `json:"ma_combo"`
	WarehouseID int64   `json:"warehouse_id" binding:"required"`
	SoLuong     float64 `json:"so_luong"`
	Note        string  `json:"note"`
}

type SaveComboMasterRequest struct {
	MaCombo      string           `json:"ma_combo"`
	WarehouseID  int64            `json:"warehouse_id"`
	TenCombo     string           `json:"ten_combo"`
	MoTa         string           `json:"mo_ta"`
	BOMSemi      []BOMItemRequest `json:"bom_semi"`
	BOMAccessory []BOMItemRequest `json:"bom_accessory"`
}

type BOMItemRequest struct {
	MaComponent string  `json:"ma_component"`
	SoLuong     float64 `json:"so_luong"`
}

type AccessoryStockInRequest struct {
	MaPhuKien   string  `json:"ma_phu_kien"`
	WarehouseID int64   `json:"warehouse_id" binding:"required"`
	SoLuong     float64 `json:"so_luong"`
	Note        string  `json:"note"`
}

type CreateAccessoryRequest struct {
	MaPhuKien  string `json:"ma_phu_kien"`
	TenPhuKien string `json:"ten_phu_kien"`
	DonViTinh  string `json:"don_vi_tinh"`
}
