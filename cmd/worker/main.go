package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"wms-v1/internal/domain"
	"wms-v1/internal/queue"
	"wms-v1/internal/repo"
	"wms-v1/internal/service"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pg, err := repo.NewPostgresRepo(ctx)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pg.Close()

	rq, err := queue.NewRedisQueue()
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rq.Close()

	impService := service.NewImportService(pg, rq)

	log.Println("Worker started, waiting for jobs...")

	// Handle shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Worker shutting down...")
		cancel()
	}()

	// Process import queue
	go processQueue(ctx, rq, pg, queue.QueueImport, func(job queue.Job) error {
		var payload service.ImportPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return err
		}
		return impService.ProcessImport(ctx, payload)
	})

	// Process bulk update queue (with recalc after each item)
	go processQueue(ctx, rq, pg, queue.QueueBulkUpdate, func(job queue.Job) error {
		var req domain.BulkUpdateRequest
		if err := json.Unmarshal(job.Payload, &req); err != nil {
			return err
		}
		for _, item := range req.Updates {
			if err := pg.UpdateInventoryItem(ctx, item.MaHang, req.WarehouseID, item.Fields); err != nil {
				log.Printf("bulk update error for %s: %v", item.MaHang, err)
				continue
			}
			// Recalc metrics immediately after each update
			if err := pg.RecalcMetricsForSKU(ctx, item.MaHang, req.WarehouseID); err != nil {
				log.Printf("recalc error for %s: %v", item.MaHang, err)
			}
		}
		return nil
	})

	// Process recalc-all queue
	go processQueue(ctx, rq, pg, queue.QueueRecalc, func(job queue.Job) error {
		log.Println("Starting recalc-all metrics...")
		skuWarehouses, err := pg.GetAllSKUsAllWarehouses(ctx)
		if err != nil {
			return err
		}
		log.Printf("Recalculating metrics for %d SKU/warehouse pairs...", len(skuWarehouses))
		for _, sw := range skuWarehouses {
			if err := pg.RecalcMetricsForSKU(ctx, sw.MaHang, sw.WarehouseID); err != nil {
				log.Printf("recalc error for %s (warehouse %d): %v", sw.MaHang, sw.WarehouseID, err)
			}
		}
		log.Printf("Recalc-all completed for %d SKU/warehouse pairs", len(skuWarehouses))
		return nil
	})

	<-ctx.Done()
	log.Println("Worker stopped")
}

func processQueue(ctx context.Context, rq *queue.RedisQueue, pg *repo.PostgresRepo, queueName string, handler func(queue.Job) error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			job, err := rq.Dequeue(ctx, queueName)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("dequeue error from %s: %v", queueName, err)
				continue
			}

			log.Printf("Processing job %s from %s", job.ID, queueName)
			if err := handler(*job); err != nil {
				log.Printf("job %s failed: %v", job.ID, err)
				pg.UpdateAsyncJob(ctx, job.ID, "failed", "{}", err.Error())
			} else {
				log.Printf("job %s completed", job.ID)
				pg.UpdateAsyncJob(ctx, job.ID, "completed", "{}", "")
			}
		}
	}
}
