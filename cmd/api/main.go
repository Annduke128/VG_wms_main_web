package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
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

	// Init handlers
	handlers := web.NewHandlers(invService, kanService, impService)

	// Setup Fiber
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024, // 50MB for file uploads
	})

	web.SetupRoutes(app, handlers)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	log.Printf("WMS API server started on :%s", port)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	app.Shutdown()
}
