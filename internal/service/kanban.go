package service

import (
	"context"
	"fmt"

	"wms-v1/internal/domain"
	"wms-v1/internal/repo"
)

type KanbanService struct {
	Repo *repo.PostgresRepo
}

func NewKanbanService(r *repo.PostgresRepo) *KanbanService {
	return &KanbanService{Repo: r}
}

// --- Inbound ---

func (s *KanbanService) ListInbound(ctx context.Context) ([]domain.KanbanInbound, error) {
	return s.Repo.ListKanbanInbound(ctx)
}

func (s *KanbanService) CreateInbound(ctx context.Context, req domain.CreateKanbanRequest) (*domain.KanbanInbound, error) {
	return s.Repo.CreateKanbanInbound(ctx, req)
}

func (s *KanbanService) MoveInbound(ctx context.Context, id int64, toStage, userID string) error {
	maHang, err := s.Repo.MoveKanbanInbound(ctx, id, toStage, userID)
	if err != nil {
		return err
	}

	// Recalc metrics after inbound reaches "da_ve_hang" (inventory changed)
	if toStage == domain.InboundStageDaVeHang && maHang != "" {
		if recalcErr := s.Repo.RecalcMetricsForSKU(ctx, maHang); recalcErr != nil {
			fmt.Printf("WARN: recalc after kanban inbound for %s: %v\n", maHang, recalcErr)
		}
	}

	return nil
}

// --- Outbound ---

func (s *KanbanService) ListOutbound(ctx context.Context) ([]domain.KanbanOutbound, error) {
	return s.Repo.ListKanbanOutbound(ctx)
}

func (s *KanbanService) CreateOutbound(ctx context.Context, req domain.CreateKanbanRequest) (*domain.KanbanOutbound, error) {
	return s.Repo.CreateKanbanOutbound(ctx, req)
}

// MoveOutbound returns negativeAlert=true if stock went negative
func (s *KanbanService) MoveOutbound(ctx context.Context, id int64, toStage, userID string) (bool, error) {
	maHang, negativeAlert, err := s.Repo.MoveKanbanOutbound(ctx, id, toStage, userID)
	if err != nil {
		return false, err
	}

	// Recalc metrics after outbound reaches "da_giao" (inventory changed)
	if toStage == domain.OutboundStageDaGiao && maHang != "" {
		if recalcErr := s.Repo.RecalcMetricsForSKU(ctx, maHang); recalcErr != nil {
			fmt.Printf("WARN: recalc after kanban outbound for %s: %v\n", maHang, recalcErr)
		}
	}

	return negativeAlert, nil
}
