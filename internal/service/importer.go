package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"wms-v1/internal/importer"
	"wms-v1/internal/queue"
	"wms-v1/internal/repo"
)

type ImportService struct {
	Repo  *repo.PostgresRepo
	Queue *queue.RedisQueue
}

func NewImportService(r *repo.PostgresRepo, q *queue.RedisQueue) *ImportService {
	return &ImportService{Repo: r, Queue: q}
}

type ImportPayload struct {
	FilePath string `json:"file_path"`
	FileType string `json:"file_type"`
	BatchID  int64  `json:"batch_id"`
}

// EnqueueImport saves the uploaded file and enqueues it for processing
func (s *ImportService) EnqueueImport(ctx context.Context, fileType string, filePath string) (string, error) {
	jobID := uuid.New().String()

	// Count rows (quick open for estimate)
	var totalRows int
	switch fileType {
	case "products":
		items, _, _ := importer.ParseProducts(filePath)
		totalRows = len(items)
	case "inventory":
		items, _, _ := importer.ParseInventoryFull(filePath)
		totalRows = len(items)
	case "inbound":
		items, _, _ := importer.ParseInbound(filePath)
		totalRows = len(items)
	case "outbound":
		items, _, _ := importer.ParseOutbound(filePath)
		totalRows = len(items)
	}

	batchID, err := s.Repo.CreateImportBatch(ctx, fileType, filepath.Base(filePath), totalRows)
	if err != nil {
		return "", fmt.Errorf("create import batch: %w", err)
	}

	payload := ImportPayload{
		FilePath: filePath,
		FileType: fileType,
		BatchID:  batchID,
	}
	payloadJSON, _ := json.Marshal(payload)

	if err := s.Repo.CreateAsyncJob(ctx, jobID, "import_"+fileType, string(payloadJSON)); err != nil {
		return "", err
	}

	job := queue.Job{
		ID:      jobID,
		Type:    "import",
		Payload: payloadJSON,
	}
	return jobID, s.Queue.Enqueue(ctx, queue.QueueImport, job)
}

// ProcessImport is called by the worker to actually do the import
func (s *ImportService) ProcessImport(ctx context.Context, payload ImportPayload) error {
	defer os.Remove(payload.FilePath) // cleanup temp file

	var success int
	var parseErrors []string
	var err error

	switch payload.FileType {
	case "products":
		var items []interface{}
		products, errs, parseErr := importer.ParseProducts(payload.FilePath)
		if parseErr != nil {
			return s.failBatch(ctx, payload.BatchID, parseErr)
		}
		parseErrors = errs
		_ = items
		success, err = s.Repo.UpsertProducts(ctx, products)

	case "inventory":
		rows, errs, parseErr := importer.ParseInventoryFull(payload.FilePath)
		if parseErr != nil {
			return s.failBatch(ctx, payload.BatchID, parseErr)
		}
		parseErrors = errs
		success, err = s.Repo.ImportInventoryFull(ctx, rows)

		// Recalculate metrics for affected SKUs
		if err == nil && len(rows) > 0 {
			seen := make(map[string]bool)
			var maHangs []string
			for _, row := range rows {
				if !seen[row.Product.MaHang] {
					seen[row.Product.MaHang] = true
					maHangs = append(maHangs, row.Product.MaHang)
				}
			}
			_ = s.Repo.RecalcMetricsForSKUs(ctx, maHangs)
		}

	case "inbound":
		items, errs, parseErr := importer.ParseInbound(payload.FilePath)
		if parseErr != nil {
			return s.failBatch(ctx, payload.BatchID, parseErr)
		}
		parseErrors = errs
		success, err = s.Repo.InsertInboundItems(ctx, items)

	case "outbound":
		items, errs, parseErr := importer.ParseOutbound(payload.FilePath)
		if parseErr != nil {
			return s.failBatch(ctx, payload.BatchID, parseErr)
		}
		parseErrors = errs
		success, err = s.Repo.InsertOutboundItems(ctx, items)

	default:
		return fmt.Errorf("unknown file type: %s", payload.FileType)
	}

	if err != nil {
		return s.failBatch(ctx, payload.BatchID, err)
	}

	// Append warning if no rows succeeded
	if success == 0 && len(parseErrors) > 0 {
		parseErrors = append(parseErrors, "⚠ Không có dòng nào import thành công. Kiểm tra định dạng file.")
	}

	errJSON, _ := json.Marshal(parseErrors)
	return s.Repo.UpdateImportBatch(ctx, payload.BatchID, success, len(parseErrors), "completed", string(errJSON))
}

func (s *ImportService) failBatch(ctx context.Context, batchID int64, err error) error {
	errJSON, _ := json.Marshal([]string{err.Error()})
	return s.Repo.UpdateImportBatch(ctx, batchID, 0, 0, "failed", string(errJSON))
}
