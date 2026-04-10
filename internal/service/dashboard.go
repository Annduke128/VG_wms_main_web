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

func (s *DashboardService) GetSummary(ctx context.Context, warehouseID int64) (*domain.DashboardSummary, error) {
	return s.Repo.GetDashboardSummary(ctx, warehouseID)
}

func (s *DashboardService) GetCharts(ctx context.Context, warehouseID int64, weeks int) (*domain.DashboardCharts, error) {
	if weeks <= 0 {
		weeks = 4
	}

	invVsOpt, err := s.Repo.GetInventoryVsOptimal(ctx, warehouseID, 20) // top 20 SKUs
	if err != nil {
		return nil, err
	}

	inOut, err := s.Repo.GetInboundOutboundByWeek(ctx, warehouseID, weeks)
	if err != nil {
		return nil, err
	}

	return &domain.DashboardCharts{
		InventoryVsOptimal: invVsOpt,
		InboundOutbound:    inOut,
	}, nil
}

func (s *DashboardService) GetAlerts(ctx context.Context, warehouseID int64) ([]domain.AlertItem, error) {
	return s.Repo.GetAlerts(ctx, warehouseID)
}

func (s *DashboardService) GetZeroSales(ctx context.Context, warehouseID int64) ([]domain.ZeroSalesItem, error) {
	return s.Repo.GetZeroSalesSKUs(ctx, warehouseID)
}

func (s *DashboardService) GetRestockAlerts(ctx context.Context, warehouseID int64) ([]domain.RestockAlertItem, error) {
	return s.Repo.GetRestockAlerts(ctx, warehouseID)
}

func (s *DashboardService) GetThresholds(ctx context.Context, maHang string, warehouseID int64) ([]domain.InventoryThreshold, error) {
	return s.Repo.GetThresholdsByMaHang(ctx, maHang, warehouseID)
}

func (s *DashboardService) SaveThreshold(ctx context.Context, req domain.ThresholdRequest) (*domain.InventoryThreshold, error) {
	return s.Repo.SaveThresholdEntry(ctx, req)
}
