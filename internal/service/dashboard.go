package service

import (
	"context"

	"wms-v1/internal/domain"
	"wms-v1/internal/repo"
)

type DashboardService struct {
	Repo *repo.PostgresRepo
}

func NewDashboardService(r *repo.PostgresRepo) *DashboardService {
	return &DashboardService{Repo: r}
}

func (s *DashboardService) GetSummary(ctx context.Context) (*domain.DashboardSummary, error) {
	return s.Repo.GetDashboardSummary(ctx)
}

func (s *DashboardService) GetCharts(ctx context.Context, weeks int) (*domain.DashboardCharts, error) {
	if weeks <= 0 {
		weeks = 4
	}

	invVsOpt, err := s.Repo.GetInventoryVsOptimal(ctx, 20) // top 20 SKUs
	if err != nil {
		return nil, err
	}

	inOut, err := s.Repo.GetInboundOutboundByWeek(ctx, weeks)
	if err != nil {
		return nil, err
	}

	return &domain.DashboardCharts{
		InventoryVsOptimal: invVsOpt,
		InboundOutbound:    inOut,
	}, nil
}

func (s *DashboardService) GetAlerts(ctx context.Context) ([]domain.AlertItem, error) {
	return s.Repo.GetAlerts(ctx)
}

func (s *DashboardService) GetZeroSales(ctx context.Context) ([]domain.ZeroSalesItem, error) {
	return s.Repo.GetZeroSalesSKUs(ctx)
}

func (s *DashboardService) GetRestockAlerts(ctx context.Context) ([]domain.RestockAlertItem, error) {
	return s.Repo.GetRestockAlerts(ctx)
}

func (s *DashboardService) GetThresholds(ctx context.Context, maHang string) ([]domain.InventoryThreshold, error) {
	return s.Repo.GetThresholdsByMaHang(ctx, maHang)
}

func (s *DashboardService) SaveThreshold(ctx context.Context, req domain.ThresholdRequest) (*domain.InventoryThreshold, error) {
	return s.Repo.SaveThresholdEntry(ctx, req)
}
