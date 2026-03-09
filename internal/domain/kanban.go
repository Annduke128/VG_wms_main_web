package domain

// Inbound stages (order matters for FIFO)
const (
	InboundStageCanNhap  = "can_nhap"
	InboundStageDaLenDon = "da_len_don"
	InboundStageDaDuyet  = "da_duyet"
	InboundStageDaVeHang = "da_ve_hang"
)

// Outbound stages
const (
	OutboundStageCanDay    = "can_day"
	OutboundStageDaChotDon = "da_chot_don"
	OutboundStageDaGiao    = "da_giao"
)

// Valid stage transitions for inbound
var InboundTransitions = map[string]string{
	InboundStageCanNhap:  InboundStageDaLenDon,
	InboundStageDaLenDon: InboundStageDaDuyet,
	InboundStageDaDuyet:  InboundStageDaVeHang,
}

// Valid stage transitions for outbound
var OutboundTransitions = map[string]string{
	OutboundStageCanDay:    OutboundStageDaChotDon,
	OutboundStageDaChotDon: OutboundStageDaGiao,
}

// ValidateInboundTransition checks if the transition is valid
func ValidateInboundTransition(from, to string) bool {
	expected, ok := InboundTransitions[from]
	return ok && expected == to
}

// ValidateOutboundTransition checks if the transition is valid
func ValidateOutboundTransition(from, to string) bool {
	expected, ok := OutboundTransitions[from]
	return ok && expected == to
}

// MoveKanbanRequest is the request body for moving a kanban card
type MoveKanbanRequest struct {
	ToStage string `json:"to_stage"`
	UserID  string `json:"user_id"`
}

// CreateKanbanRequest is the request body for creating a kanban card
type CreateKanbanRequest struct {
	MaHang     string  `json:"ma_hang"`
	TenSanPham string  `json:"ten_san_pham"`
	SoLuong    float64 `json:"so_luong"`
	Note       string  `json:"note"`
}
