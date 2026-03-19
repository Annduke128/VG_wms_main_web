package domain

import "time"

// DashboardSummary contains the 4 KPI values
type DashboardSummary struct {
	SKUCount     int64   `json:"sku_count"`
	SKUTonLau    int64   `json:"sku_ton_lau"`
	SKUThieuHang int64   `json:"sku_thieu_hang"`
	TongTienHang float64 `json:"tong_tien_hang"`
}

// ChartDataPoint represents one point in a time-series chart
type ChartDataPoint struct {
	Week      string  `json:"week"`       // e.g. "2026-W12"
	WeekStart string  `json:"week_start"` // ISO date
	Value     float64 `json:"value"`
}

// InventoryVsOptimalItem represents one SKU's current vs optimal stock
type InventoryVsOptimalItem struct {
	MaHang     string  `json:"ma_hang"`
	TenSanPham string  `json:"ten_san_pham"`
	SoTon      float64 `json:"so_ton"`
	OptimalQty float64 `json:"optimal_qty"`
}

// DashboardCharts contains both chart datasets
type DashboardCharts struct {
	InventoryVsOptimal []InventoryVsOptimalItem `json:"inventory_vs_optimal"`
	InboundOutbound    []InOutWeekData          `json:"inbound_outbound"`
}

// InOutWeekData represents inbound/outbound totals for one week
type InOutWeekData struct {
	Week      string  `json:"week"`
	WeekStart string  `json:"week_start"`
	Inbound   float64 `json:"inbound"`
	Outbound  float64 `json:"outbound"`
}

// AlertItem represents one inventory alert row
type AlertItem struct {
	MaHang     string  `json:"ma_hang"`
	TenSanPham string  `json:"ten_san_pham"`
	SoTon      float64 `json:"so_ton"`
	AlertType  string  `json:"alert_type"` // "ton_lau" or "thieu_hang"
	// For tồn lâu
	SoNgayTon  float64 `json:"so_ngay_ton,omitempty"`
	MaxAgeDays float64 `json:"max_age_days,omitempty"`
	OptimalQty float64 `json:"optimal_qty,omitempty"`
	// For thiếu hàng
	MinQty float64 `json:"min_qty,omitempty"`
}

// ThresholdRequest is the POST body for saving a threshold
type ThresholdRequest struct {
	MaHang        string     `json:"ma_hang" binding:"required"`
	MinQty        float64    `json:"min_qty"`
	OptimalQty    float64    `json:"optimal_qty"`
	MaxAgeDays    float64    `json:"max_age_days"`
	Source        string     `json:"source"` // "manual"
	EffectiveFrom *time.Time `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to"`
}
