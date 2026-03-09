package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"wms-v1/internal/domain"
)

// --- Inbound Kanban ---

func (r *PostgresRepo) ListKanbanInbound(ctx context.Context) ([]domain.KanbanInbound, error) {
	rows, err := r.Pool.Query(ctx, "SELECT id, ma_hang, ten_san_pham, so_luong, stage, note, created_at, updated_at FROM kanban_inbound ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("list kanban inbound: %w", err)
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.KanbanInbound])
}

func (r *PostgresRepo) GetKanbanInbound(ctx context.Context, id int64) (*domain.KanbanInbound, error) {
	row := r.Pool.QueryRow(ctx, "SELECT id, ma_hang, ten_san_pham, so_luong, stage, note, created_at, updated_at FROM kanban_inbound WHERE id = $1", id)
	var item domain.KanbanInbound
	err := row.Scan(&item.ID, &item.MaHang, &item.TenSanPham, &item.SoLuong, &item.Stage, &item.Note, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get kanban inbound %d: %w", id, err)
	}
	return &item, nil
}

func (r *PostgresRepo) CreateKanbanInbound(ctx context.Context, req domain.CreateKanbanRequest) (*domain.KanbanInbound, error) {
	var item domain.KanbanInbound
	err := r.Pool.QueryRow(ctx,
		`INSERT INTO kanban_inbound (ma_hang, ten_san_pham, so_luong, stage, note)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, ma_hang, ten_san_pham, so_luong, stage, note, created_at, updated_at`,
		req.MaHang, req.TenSanPham, req.SoLuong, domain.InboundStageCanNhap, req.Note,
	).Scan(&item.ID, &item.MaHang, &item.TenSanPham, &item.SoLuong, &item.Stage, &item.Note, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create kanban inbound: %w", err)
	}
	return &item, nil
}

// MoveKanbanInbound transitions card and applies side-effects when reaching "da_ve_hang"
func (r *PostgresRepo) MoveKanbanInbound(ctx context.Context, id int64, toStage string, userID string) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current card
	var card domain.KanbanInbound
	err = tx.QueryRow(ctx,
		"SELECT id, ma_hang, ten_san_pham, so_luong, stage FROM kanban_inbound WHERE id = $1 FOR UPDATE", id,
	).Scan(&card.ID, &card.MaHang, &card.TenSanPham, &card.SoLuong, &card.Stage)
	if err != nil {
		return fmt.Errorf("get card for move: %w", err)
	}

	// Validate transition
	if !domain.ValidateInboundTransition(card.Stage, toStage) {
		return fmt.Errorf("invalid transition from %s to %s", card.Stage, toStage)
	}

	// Update card stage
	_, err = tx.Exec(ctx,
		"UPDATE kanban_inbound SET stage = $1, updated_at = $2 WHERE id = $3",
		toStage, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update card stage: %w", err)
	}

	// Log kanban event
	_, err = tx.Exec(ctx,
		"INSERT INTO kanban_events (sku, from_stage, to_stage, user_id) VALUES ($1, $2, $3, $4)",
		card.MaHang, card.Stage, toStage, userID)
	if err != nil {
		return fmt.Errorf("log kanban event: %w", err)
	}

	// Side-effects when card reaches "da_ve_hang"
	if toStage == domain.InboundStageDaVeHang {
		now := time.Now()
		// Insert inbound_items record
		_, err = tx.Exec(ctx,
			`INSERT INTO inbound_items (ma_hang, ten_san_pham, so_luong, ngay_nhan_hang)
			 VALUES ($1, $2, $3, $4)`
			,
			card.MaHang, card.TenSanPham, card.SoLuong, now)
		if err != nil {
			return fmt.Errorf("insert inbound item: %w", err)
		}

		// Update inventory_main: add stock
		_, err = tx.Exec(ctx,
			`INSERT INTO inventory_main (ma_hang, ten_san_pham, so_ton, so_nhap)
			 VALUES ($1, $2, $3, $3)
			 ON CONFLICT (ma_hang) DO UPDATE SET
			   so_ton = inventory_main.so_ton + EXCLUDED.so_ton,
			   so_nhap = inventory_main.so_nhap + EXCLUDED.so_nhap`,
			card.MaHang, card.TenSanPham, card.SoLuong)
		if err != nil {
			return fmt.Errorf("update inventory for inbound: %w", err)
		}

		// Log movement
		_, err = tx.Exec(ctx,
			"INSERT INTO inventory_movements (ma_hang, qty, movement_type) VALUES ($1, $2, 'IN')",
			card.MaHang, card.SoLuong)
		if err != nil {
			return fmt.Errorf("log inbound movement: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// --- Outbound Kanban ---

func (r *PostgresRepo) ListKanbanOutbound(ctx context.Context) ([]domain.KanbanOutbound, error) {
	rows, err := r.Pool.Query(ctx, "SELECT id, ma_hang, ten_san_pham, so_luong, stage, note, created_at, updated_at FROM kanban_outbound ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("list kanban outbound: %w", err)
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.KanbanOutbound])
}

func (r *PostgresRepo) GetKanbanOutbound(ctx context.Context, id int64) (*domain.KanbanOutbound, error) {
	row := r.Pool.QueryRow(ctx, "SELECT id, ma_hang, ten_san_pham, so_luong, stage, note, created_at, updated_at FROM kanban_outbound WHERE id = $1", id)
	var item domain.KanbanOutbound
	err := row.Scan(&item.ID, &item.MaHang, &item.TenSanPham, &item.SoLuong, &item.Stage, &item.Note, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get kanban outbound %d: %w", id, err)
	}
	return &item, nil
}

func (r *PostgresRepo) CreateKanbanOutbound(ctx context.Context, req domain.CreateKanbanRequest) (*domain.KanbanOutbound, error) {
	var item domain.KanbanOutbound
	err := r.Pool.QueryRow(ctx,
		`INSERT INTO kanban_outbound (ma_hang, ten_san_pham, so_luong, stage, note)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, ma_hang, ten_san_pham, so_luong, stage, note, created_at, updated_at`,
		req.MaHang, req.TenSanPham, req.SoLuong, domain.OutboundStageCanDay, req.Note,
	).Scan(&item.ID, &item.MaHang, &item.TenSanPham, &item.SoLuong, &item.Stage, &item.Note, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create kanban outbound: %w", err)
	}
	return &item, nil
}

// MoveKanbanOutbound transitions card and applies side-effects when reaching "da_giao"
// IMPORTANT: Negative stock is allowed but triggers an alert
func (r *PostgresRepo) MoveKanbanOutbound(ctx context.Context, id int64, toStage string, userID string) (negativeAlert bool, err error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current card
	var card domain.KanbanOutbound
	err = tx.QueryRow(ctx,
		"SELECT id, ma_hang, ten_san_pham, so_luong, stage FROM kanban_outbound WHERE id = $1 FOR UPDATE", id,
	).Scan(&card.ID, &card.MaHang, &card.TenSanPham, &card.SoLuong, &card.Stage)
	if err != nil {
		return false, fmt.Errorf("get card for move: %w", err)
	}

	// Validate transition
	if !domain.ValidateOutboundTransition(card.Stage, toStage) {
		return false, fmt.Errorf("invalid transition from %s to %s", card.Stage, toStage)
	}

	// Update card stage
	_, err = tx.Exec(ctx,
		"UPDATE kanban_outbound SET stage = $1, updated_at = $2 WHERE id = $3",
		toStage, time.Now(), id)
	if err != nil {
		return false, fmt.Errorf("update card stage: %w", err)
	}

	// Log kanban event
	_, err = tx.Exec(ctx,
		"INSERT INTO kanban_events (sku, from_stage, to_stage, user_id) VALUES ($1, $2, $3, $4)",
		card.MaHang, card.Stage, toStage, userID)
	if err != nil {
		return false, fmt.Errorf("log kanban event: %w", err)
	}

	// Side-effects when card reaches "da_giao"
	if toStage == domain.OutboundStageDaGiao {
		now := time.Now()
		// Insert outbound_items record
		_, err = tx.Exec(ctx,
			`INSERT INTO outbound_items (ma_hang, ten_san_pham, so_luong, ngay_nhan_hang)
			 VALUES ($1, $2, $3, $4)`
			,
			card.MaHang, card.TenSanPham, card.SoLuong, now)
		if err != nil {
			return false, fmt.Errorf("insert outbound item: %w", err)
		}

		// Update inventory_main: subtract stock (negative allowed)
		_, err = tx.Exec(ctx,
			`INSERT INTO inventory_main (ma_hang, ten_san_pham, so_ton, so_xuat)
			 VALUES ($1, $2, -$3, $3)
			 ON CONFLICT (ma_hang) DO UPDATE SET
			   so_ton = inventory_main.so_ton - EXCLUDED.so_xuat,
			   so_xuat = inventory_main.so_xuat + EXCLUDED.so_xuat`,
			card.MaHang, card.TenSanPham, card.SoLuong)
		if err != nil {
			return false, fmt.Errorf("update inventory for outbound: %w", err)
		}

		// Check for negative stock
		var currentStock float64
		err = tx.QueryRow(ctx, "SELECT so_ton FROM inventory_main WHERE ma_hang = $1", card.MaHang).Scan(&currentStock)
		if err != nil {
			return false, fmt.Errorf("check stock: %w", err)
		}
		if currentStock < 0 {
			negativeAlert = true
		}

		// Log movement
		_, err = tx.Exec(ctx,
			"INSERT INTO inventory_movements (ma_hang, qty, movement_type) VALUES ($1, $2, 'OUT')",
			card.MaHang, card.SoLuong)
		if err != nil {
			return false, fmt.Errorf("log outbound movement: %w", err)
		}
	}

	return negativeAlert, tx.Commit(ctx)
}
