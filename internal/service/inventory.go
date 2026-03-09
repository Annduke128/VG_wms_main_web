package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"wms-v1/internal/domain"
	"wms-v1/internal/queue"
	"wms-v1/internal/repo"
)

type InventoryService struct {
	Repo  *repo.PostgresRepo
	Queue *queue.RedisQueue
}

func NewInventoryService(r *repo.PostgresRepo, q *queue.RedisQueue) *InventoryService {
	return &InventoryService{Repo: r, Queue: q}
}

func (s *InventoryService) GridQuery(ctx context.Context, req domain.GridRequest) (*domain.GridResponse, error) {
	return s.Repo.QueryInventoryGrid(ctx, req)
}

func (s *InventoryService) UpdateItem(ctx context.Context, maHang string, fields map[string]interface{}) error {
	return s.Repo.UpdateInventoryItem(ctx, maHang, fields)
}

func (s *InventoryService) BulkUpdate(ctx context.Context, req domain.BulkUpdateRequest) (string, error) {
	jobID := uuid.New().String()

	payload, _ := json.Marshal(req)

	// Create job record in DB
	if err := s.Repo.CreateAsyncJob(ctx, jobID, "bulk_update", string(payload)); err != nil {
		return "", err
	}

	// Enqueue for worker
	job := queue.Job{
		ID:      jobID,
		Type:    "bulk_update",
		Payload: payload,
	}
	if err := s.Queue.Enqueue(ctx, queue.QueueBulkUpdate, job); err != nil {
		return "", err
	}

	return jobID, nil
}

func (s *InventoryService) GetJob(ctx context.Context, jobID string) (*domain.AsyncJob, error) {
	return s.Repo.GetAsyncJob(ctx, jobID)
}
