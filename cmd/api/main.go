package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"wms-v1/internal/queue"
	"wms-v1/internal/repo"
	"wms-v1/internal/service"
	"wms-v1/internal/web"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Init Postgres
	pg, err := repo.NewPostgresRepo(ctx)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pg.Close()

	// Init Redis
	rq, err := queue.NewRedisQueue()
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rq.Close()

	// Init services
	invService := service.NewInventoryService(pg, rq)
	kanService := service.NewKanbanService(pg)
	impService := service.NewImportService(pg, rq)
	ordService := service.NewOrderService(pg)
	dashService := service.NewDashboardService(pg)
	comboService := service.NewComboService(pg)

	// Init handlers
	handlers := web.NewHandlers(invService, kanService, impService, ordService, dashService, comboService, rq)

	// Setup Gin router
	router := web.SetupRoutes(handlers)

	// Setup HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("WMS API server started on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
}
