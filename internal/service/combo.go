package service

import (
	"context"
	"fmt"

	"wms-v1/internal/domain"
	"wms-v1/internal/repo"
)

type ComboService struct {
	Repo *repo.PostgresRepo
}

func NewComboService(r *repo.PostgresRepo) *ComboService {
	return &ComboService{Repo: r}
}

// --- Combo Master ---

func (s *ComboService) ListComboMasters(ctx context.Context) ([]domain.ComboMaster, error) {
	return s.Repo.ListComboMasters(ctx, true)
}

func (s *ComboService) GetComboDetail(ctx context.Context, maCombo string) (*domain.ComboDetail, error) {
	return s.Repo.GetComboDetail(ctx, maCombo)
}

func (s *ComboService) SaveComboMaster(ctx context.Context, req domain.SaveComboMasterRequest) error {
	if req.MaCombo == "" || req.TenCombo == "" {
		return fmt.Errorf("ma_combo and ten_combo are required")
	}

	tx, err := s.Repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.Repo.SaveComboMaster(ctx, tx, req); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *ComboService) DeleteComboMaster(ctx context.Context, maCombo string) error {
	return s.Repo.DeleteComboMaster(ctx, maCombo)
}

// --- Combo Inventory ---

func (s *ComboService) GetComboInventory(ctx context.Context, warehouseID int64) ([]domain.ComboInventory, error) {
	return s.Repo.GetComboInventory(ctx, warehouseID)
}

// --- Combo Transactions ---

func (s *ComboService) ListComboTransactions(ctx context.Context, maCombo string, page, limit int, warehouseID int64) ([]domain.ComboTransaction, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}
	return s.Repo.ListComboTransactions(ctx, maCombo, limit, offset, warehouseID)
}

// CreateCombo: tạo combo từ NVL
// 1. Load BOM
// 2. Check NVL đủ (kho chính FIFO + phụ kiện)
// 3. Trừ kho chính (FIFO lots) + tăng so_xuat
// 4. Trừ phụ kiện
// 5. Cộng combo inventory
// 6. Ghi transaction + component movements
// 7. Recalc metrics cho mỗi ma_hang bị trừ
func (s *ComboService) CreateCombo(ctx context.Context, req domain.CreateComboRequest) (*domain.ComboTransaction, error) {
	if req.MaCombo == "" || req.SoLuong <= 0 {
		return nil, fmt.Errorf("ma_combo and so_luong > 0 required")
	}

	tx, err := s.Repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Load BOM
	bomSemi, bomAcc, err := s.Repo.GetBOMForCombo(ctx, tx, req.MaCombo)
	if err != nil {
		return nil, err
	}
	if len(bomSemi) == 0 && len(bomAcc) == 0 {
		return nil, fmt.Errorf("combo %s has empty BOM", req.MaCombo)
	}

	// 2. Insert transaction first to get ID
	txnID, err := s.Repo.InsertComboTransaction(ctx, tx, req.MaCombo, "CREATE", req.SoLuong, req.Note, req.WarehouseID)
	if err != nil {
		return nil, err
	}

	// 3. Deduct semi-finished products (FIFO)
	var affectedSKUs []string
	for _, bom := range bomSemi {
		needed := bom.SoLuong * req.SoLuong

		// Get FIFO lots
		lots, err := s.Repo.GetAvailableLotsFIFO(ctx, tx, bom.MaHang, req.WarehouseID)
		if err != nil {
			return nil, fmt.Errorf("get FIFO lots for %s: %w", bom.MaHang, err)
		}

		var totalAvailable float64
		for _, l := range lots {
			totalAvailable += l.QtyRemaining
		}
		if totalAvailable < needed {
			return nil, fmt.Errorf("thiếu NVL %s: cần %.2f, còn %.2f", bom.MaHang, needed, totalAvailable)
		}

		// Allocate FIFO
		remaining := needed
		for _, lot := range lots {
			if remaining <= 0 {
				break
			}
			allocQty := lot.QtyRemaining
			if allocQty > remaining {
				allocQty = remaining
			}

			if err := s.Repo.DeductLot(ctx, tx, lot.ID, allocQty); err != nil {
				return nil, fmt.Errorf("deduct lot for %s: %w", bom.MaHang, err)
			}

			// Record component movement with lot reference
			lotID := lot.ID
			if err := s.Repo.InsertComboComponentMovement(ctx, tx, txnID, "SEMI", bom.MaHang, allocQty, &lotID, req.WarehouseID); err != nil {
				return nil, err
			}

			remaining -= allocQty
		}

		// Update inventory_main: trừ so_ton, tăng so_xuat (KHÔNG ghi outbound_items → LBBQ không tính)
		_, err = tx.Exec(ctx, `
			UPDATE inventory_main
			SET so_ton = so_ton - $1, so_xuat = so_xuat + $1
			WHERE ma_hang = $2 AND warehouse_id = $3`, needed, bom.MaHang, req.WarehouseID)
		if err != nil {
			return nil, fmt.Errorf("update inventory_main for %s: %w", bom.MaHang, err)
		}

		affectedSKUs = append(affectedSKUs, bom.MaHang)
	}

	// 4. Deduct accessories
	for _, bom := range bomAcc {
		needed := bom.SoLuong * req.SoLuong

		// Check stock
		stock, err := s.Repo.GetAccessoryStockForUpdate(ctx, tx, bom.MaPhuKien, req.WarehouseID)
		if err != nil {
			return nil, fmt.Errorf("get accessory stock %s: %w", bom.MaPhuKien, err)
		}
		if stock < needed {
			return nil, fmt.Errorf("thiếu phụ kiện %s: cần %.2f, còn %.2f", bom.MaPhuKien, needed, stock)
		}

		if err := s.Repo.UpdateAccessoryStock(ctx, tx, bom.MaPhuKien, req.WarehouseID, -needed); err != nil {
			return nil, err
		}

		// Record accessory movement
		if err := s.Repo.InsertAccessoryMovement(ctx, tx, bom.MaPhuKien, "OUT", needed, fmt.Sprintf("Tạo combo %s x%.0f", req.MaCombo, req.SoLuong), req.WarehouseID); err != nil {
			return nil, err
		}

		// Record component movement
		if err := s.Repo.InsertComboComponentMovement(ctx, tx, txnID, "ACCESSORY", bom.MaPhuKien, needed, nil, req.WarehouseID); err != nil {
			return nil, err
		}
	}

	// 5. Update combo inventory: +so_ton, +so_nhap
	if err := s.Repo.UpdateComboInventory(ctx, tx, req.MaCombo, req.WarehouseID, req.SoLuong, req.SoLuong, 0, 0); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// 6. Recalc metrics (outside tx)
	for _, sku := range affectedSKUs {
		_ = s.Repo.RecalcMetricsForSKU(ctx, sku, req.WarehouseID)
	}

	txn := &domain.ComboTransaction{
		ID:              txnID,
		MaCombo:         req.MaCombo,
		TransactionType: "CREATE",
		SoLuong:         req.SoLuong,
		Note:            req.Note,
	}
	return txn, nil
}

// CancelCombo: hủy combo → hoàn nguyên NVL
// 1. Check combo_inventory đủ
// 2. Trừ combo inventory
// 3. Cộng lại kho chính (so_ton, giảm so_xuat) — cộng lại vào lot gần nhất (simplified)
// 4. Cộng lại phụ kiện
// 5. Ghi transaction
// 6. Recalc metrics
func (s *ComboService) CancelCombo(ctx context.Context, req domain.CancelComboRequest) (*domain.ComboTransaction, error) {
	if req.MaCombo == "" || req.SoLuong <= 0 {
		return nil, fmt.Errorf("ma_combo and so_luong > 0 required")
	}

	tx, err := s.Repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check combo stock
	ci, err := s.Repo.GetComboInventoryForUpdate(ctx, tx, req.MaCombo, req.WarehouseID)
	if err != nil {
		return nil, err
	}
	if ci.SoTon < req.SoLuong {
		return nil, fmt.Errorf("tồn combo không đủ: còn %.2f, cần hủy %.2f", ci.SoTon, req.SoLuong)
	}

	// Load BOM
	bomSemi, bomAcc, err := s.Repo.GetBOMForCombo(ctx, tx, req.MaCombo)
	if err != nil {
		return nil, err
	}

	// Insert transaction
	txnID, err := s.Repo.InsertComboTransaction(ctx, tx, req.MaCombo, "CANCEL", req.SoLuong, req.Note, req.WarehouseID)
	if err != nil {
		return nil, err
	}

	// Restore semi-finished products
	var affectedSKUs []string
	for _, bom := range bomSemi {
		restoreQty := bom.SoLuong * req.SoLuong

		// Cộng lại inventory_main: tăng so_ton, giảm so_xuat
		_, err = tx.Exec(ctx, `
			UPDATE inventory_main
			SET so_ton = so_ton + $1, so_xuat = so_xuat - $1
			WHERE ma_hang = $2 AND warehouse_id = $3`, restoreQty, bom.MaHang, req.WarehouseID)
		if err != nil {
			return nil, fmt.Errorf("restore inventory_main for %s: %w", bom.MaHang, err)
		}

		// Cộng lại vào lot mới nhất (simplified — tạo lot mới nếu cần)
		_, err = tx.Exec(ctx, `
			UPDATE inventory_lots SET
				qty_out = GREATEST(qty_out - $2, 0),
				qty_remaining = qty_remaining + $2
			WHERE id = (
				SELECT id FROM inventory_lots
				WHERE ma_hang = $1 AND warehouse_id = $3
				ORDER BY received_at DESC, id DESC
				LIMIT 1
			)`, bom.MaHang, restoreQty, req.WarehouseID)
		if err != nil {
			return nil, fmt.Errorf("restore lot for %s: %w", bom.MaHang, err)
		}

		if err := s.Repo.InsertComboComponentMovement(ctx, tx, txnID, "SEMI", bom.MaHang, restoreQty, nil, req.WarehouseID); err != nil {
			return nil, err
		}

		affectedSKUs = append(affectedSKUs, bom.MaHang)
	}

	// Restore accessories
	for _, bom := range bomAcc {
		restoreQty := bom.SoLuong * req.SoLuong

		if err := s.Repo.UpdateAccessoryStock(ctx, tx, bom.MaPhuKien, req.WarehouseID, restoreQty); err != nil {
			return nil, err
		}

		if err := s.Repo.InsertAccessoryMovement(ctx, tx, bom.MaPhuKien, "RETURN", restoreQty, fmt.Sprintf("Hủy combo %s x%.0f", req.MaCombo, req.SoLuong), req.WarehouseID); err != nil {
			return nil, err
		}

		if err := s.Repo.InsertComboComponentMovement(ctx, tx, txnID, "ACCESSORY", bom.MaPhuKien, restoreQty, nil, req.WarehouseID); err != nil {
			return nil, err
		}
	}

	// Trừ combo inventory
	if err := s.Repo.UpdateComboInventory(ctx, tx, req.MaCombo, req.WarehouseID, -req.SoLuong, 0, 0, 0); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	for _, sku := range affectedSKUs {
		_ = s.Repo.RecalcMetricsForSKU(ctx, sku, req.WarehouseID)
	}

	return &domain.ComboTransaction{
		ID:              txnID,
		MaCombo:         req.MaCombo,
		TransactionType: "CANCEL",
		SoLuong:         req.SoLuong,
		Note:            req.Note,
	}, nil
}

// ComboOut: xuất combo
// Trừ combo_inventory (so_ton, tăng so_xuat)
func (s *ComboService) ComboOut(ctx context.Context, req domain.ComboOutRequest) (*domain.ComboTransaction, error) {
	if req.MaCombo == "" || req.SoLuong <= 0 {
		return nil, fmt.Errorf("ma_combo and so_luong > 0 required")
	}

	tx, err := s.Repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	ci, err := s.Repo.GetComboInventoryForUpdate(ctx, tx, req.MaCombo, req.WarehouseID)
	if err != nil {
		return nil, err
	}
	if ci.SoTon < req.SoLuong {
		return nil, fmt.Errorf("tồn combo không đủ: còn %.2f, cần xuất %.2f", ci.SoTon, req.SoLuong)
	}

	txnID, err := s.Repo.InsertComboTransaction(ctx, tx, req.MaCombo, "OUT", req.SoLuong, req.Note, req.WarehouseID)
	if err != nil {
		return nil, err
	}

	// -so_ton, +so_xuat
	if err := s.Repo.UpdateComboInventory(ctx, tx, req.MaCombo, req.WarehouseID, -req.SoLuong, 0, req.SoLuong, 0); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &domain.ComboTransaction{
		ID:              txnID,
		MaCombo:         req.MaCombo,
		TransactionType: "OUT",
		SoLuong:         req.SoLuong,
		Note:            req.Note,
	}, nil
}

// ComboReturn: trả hàng combo → tăng tồn combo (KHÔNG hoàn nguyên NVL)
func (s *ComboService) ComboReturn(ctx context.Context, req domain.ComboReturnRequest) (*domain.ComboTransaction, error) {
	if req.MaCombo == "" || req.SoLuong <= 0 {
		return nil, fmt.Errorf("ma_combo and so_luong > 0 required")
	}

	tx, err := s.Repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	txnID, err := s.Repo.InsertComboTransaction(ctx, tx, req.MaCombo, "RETURN", req.SoLuong, req.Note, req.WarehouseID)
	if err != nil {
		return nil, err
	}

	// +so_ton, +so_tra
	if err := s.Repo.UpdateComboInventory(ctx, tx, req.MaCombo, req.WarehouseID, req.SoLuong, 0, 0, req.SoLuong); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &domain.ComboTransaction{
		ID:              txnID,
		MaCombo:         req.MaCombo,
		TransactionType: "RETURN",
		SoLuong:         req.SoLuong,
		Note:            req.Note,
	}, nil
}

// --- Accessories ---

func (s *ComboService) ListAccessories(ctx context.Context) ([]domain.Accessory, error) {
	return s.Repo.ListAccessories(ctx)
}

func (s *ComboService) CreateAccessory(ctx context.Context, req domain.CreateAccessoryRequest) error {
	if req.MaPhuKien == "" || req.TenPhuKien == "" {
		return fmt.Errorf("ma_phu_kien and ten_phu_kien are required")
	}

	tx, err := s.Repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	acc := &domain.Accessory{
		MaPhuKien:  req.MaPhuKien,
		TenPhuKien: req.TenPhuKien,
		DonViTinh:  req.DonViTinh,
	}
	if err := s.Repo.CreateAccessory(ctx, tx, acc); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *ComboService) GetAccessoryInventory(ctx context.Context, warehouseID int64) ([]domain.AccessoryInventory, error) {
	return s.Repo.GetAccessoryInventory(ctx, warehouseID)
}

// AccessoryStockIn: nhập phụ kiện
func (s *ComboService) AccessoryStockIn(ctx context.Context, req domain.AccessoryStockInRequest) error {
	if req.MaPhuKien == "" || req.SoLuong <= 0 {
		return fmt.Errorf("ma_phu_kien and so_luong > 0 required")
	}

	tx, err := s.Repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.Repo.UpdateAccessoryStock(ctx, tx, req.MaPhuKien, req.WarehouseID, req.SoLuong); err != nil {
		return err
	}

	if err := s.Repo.InsertAccessoryMovement(ctx, tx, req.MaPhuKien, "IN", req.SoLuong, req.Note, req.WarehouseID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
