package service

import (
	"context"
	"fmt"
	"time"

	"wms-v1/internal/domain"
	"wms-v1/internal/repo"
)

type OrderService struct {
	Repo *repo.PostgresRepo
}

func NewOrderService(r *repo.PostgresRepo) *OrderService {
	return &OrderService{Repo: r}
}

// CreateInbound processes an inbound order:
// 1. Insert inbound_items row (with batch_code)
// 2. Upsert inventory_lots
// 3. Update inventory_main (so_ton, so_nhap)
// 4. Record movement
func (s *OrderService) CreateInbound(ctx context.Context, req domain.CreateInboundRequest) (*domain.InboundResult, error) {
	if req.MaHang == "" {
		return nil, fmt.Errorf("ma_hang is required")
	}
	if req.SoLuong <= 0 {
		return nil, fmt.Errorf("so_luong must be positive")
	}
	if req.BatchCode == "" {
		return nil, fmt.Errorf("batch_code is required")
	}

	tx, err := s.Repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Parse ngay_nhan_hang
	receivedAt := time.Now()
	if req.NgayNhanHang != "" {
		parsed, err := time.Parse(time.RFC3339, req.NgayNhanHang)
		if err != nil {
			return nil, fmt.Errorf("invalid ngay_nhan_hang format (use RFC3339): %w", err)
		}
		receivedAt = parsed
	}

	// 1. Insert inbound_items
	item := &domain.InboundItem{
		MaHang:       req.MaHang,
		TenSanPham:   req.TenSanPham,
		DonViTinh:    req.DonViTinh,
		QuyCach:      req.QuyCach,
		SoLuong:      req.SoLuong,
		BatchCode:    req.BatchCode,
		DoanhSo:      req.DoanhSo,
		ChietKhau:    req.ChietKhau,
		DoanhThu:     req.DoanhThu,
		Von:          req.Von,
		NgayNhanHang: receivedAt,
	}
	if err := s.Repo.InsertInboundItem(ctx, tx, item); err != nil {
		return nil, fmt.Errorf("insert inbound: %w", err)
	}

	// 2. Upsert lot
	lot := &domain.InventoryLot{
		MaHang:     req.MaHang,
		BatchCode:  req.BatchCode,
		ReceivedAt: receivedAt,
		QtyIn:      req.SoLuong,
	}
	if err := s.Repo.UpsertInventoryLot(ctx, tx, lot); err != nil {
		return nil, fmt.Errorf("upsert lot: %w", err)
	}

	// 3. Update inventory_main
	if err := s.Repo.UpdateInventoryMainInbound(ctx, tx, req.MaHang, req.SoLuong); err != nil {
		return nil, fmt.Errorf("update inventory: %w", err)
	}

	// 4. Movement record
	if err := s.Repo.InsertInventoryMovement(ctx, tx, req.MaHang, req.SoLuong, "IN"); err != nil {
		return nil, fmt.Errorf("insert movement: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// Recalculate metrics after inbound
	_ = s.Repo.RecalcMetricsForSKU(ctx, req.MaHang)

	return &domain.InboundResult{
		InboundItem: *item,
		Lot:         *lot,
	}, nil
}

// CreateOutbound processes an outbound order with FIFO allocation:
// 1. Get lots FIFO (oldest first)
// 2. Allocate from each lot until qty fulfilled
// 3. Insert outbound_items row(s) — one per lot used
// 4. Update inventory_main
// 5. Record movement
func (s *OrderService) CreateOutbound(ctx context.Context, req domain.CreateOutboundRequest) (*domain.OutboundResult, error) {
	if req.MaHang == "" {
		return nil, fmt.Errorf("ma_hang is required")
	}
	if req.SoLuong <= 0 {
		return nil, fmt.Errorf("so_luong must be positive")
	}

	tx, err := s.Repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Get available lots FIFO
	lots, err := s.Repo.GetAvailableLotsFIFO(ctx, tx, req.MaHang)
	if err != nil {
		return nil, fmt.Errorf("get FIFO lots: %w", err)
	}

	// Calculate total available
	var totalAvailable float64
	for _, l := range lots {
		totalAvailable += l.QtyRemaining
	}
	if totalAvailable < req.SoLuong {
		return nil, fmt.Errorf("insufficient stock: available=%.2f, requested=%.2f", totalAvailable, req.SoLuong)
	}

	// 2. Allocate FIFO
	remaining := req.SoLuong
	var allocations []domain.OutboundAllocation
	var outboundItems []domain.OutboundItem

	for _, lot := range lots {
		if remaining <= 0 {
			break
		}

		allocQty := lot.QtyRemaining
		if allocQty > remaining {
			allocQty = remaining
		}

		// Deduct from lot
		if err := s.Repo.DeductLot(ctx, tx, lot.ID, allocQty); err != nil {
			return nil, fmt.Errorf("deduct lot: %w", err)
		}

		// Insert outbound_items row for this allocation
		outItem := &domain.OutboundItem{
			MaHang:     req.MaHang,
			TenSanPham: req.TenSanPham,
			DonViTinh:  req.DonViTinh,
			QuyCach:    req.QuyCach,
			SoLuong:    allocQty,
			BatchCode:  lot.BatchCode,
			DoanhSo:    req.DoanhSo * (allocQty / req.SoLuong), // proportional
			ChietKhau:  req.ChietKhau * (allocQty / req.SoLuong),
			DoanhThu:   req.DoanhThu * (allocQty / req.SoLuong),
			Von:        req.Von * (allocQty / req.SoLuong),
		}
		if err := s.Repo.InsertOutboundItem(ctx, tx, outItem); err != nil {
			return nil, fmt.Errorf("insert outbound: %w", err)
		}

		allocations = append(allocations, domain.OutboundAllocation{
			BatchCode:    lot.BatchCode,
			AllocatedQty: allocQty,
			LotID:        lot.ID,
		})
		outboundItems = append(outboundItems, *outItem)

		remaining -= allocQty
	}

	// 3. Update inventory_main
	totalAllocated := req.SoLuong - remaining
	if err := s.Repo.UpdateInventoryMainOutbound(ctx, tx, req.MaHang, totalAllocated); err != nil {
		return nil, fmt.Errorf("update inventory: %w", err)
	}

	// 4. Movement record
	if err := s.Repo.InsertInventoryMovement(ctx, tx, req.MaHang, totalAllocated, "OUT"); err != nil {
		return nil, fmt.Errorf("insert movement: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// Recalculate metrics after outbound
	_ = s.Repo.RecalcMetricsForSKU(ctx, req.MaHang)

	return &domain.OutboundResult{
		OutboundItems:  outboundItems,
		Allocations:    allocations,
		TotalAllocated: totalAllocated,
	}, nil
}

// ListOrders returns paginated list of orders (UNION inbound + outbound) with filters
func (s *OrderService) ListOrders(ctx context.Context, f domain.OrderFilter) ([]domain.OrderListItem, int64, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	return s.Repo.ListOrders(ctx, f)
}

// GetLots returns lot details for a product
func (s *OrderService) GetLots(ctx context.Context, maHang string) ([]domain.InventoryLot, error) {
	return s.Repo.GetLotsByMaHang(ctx, maHang)
}
