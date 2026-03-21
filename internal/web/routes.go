package web

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(h *Handlers) *gin.Engine {
	r := gin.Default() // includes Logger + Recovery middleware

	// CORS
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowHeaders:    []string{"Content-Type", "Authorization"},
		AllowMethods:    []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")

	// Inventory
	api.POST("/inventory/grid", h.InventoryGrid)
	api.PATCH("/inventory/:ma_hang", h.UpdateInventoryItem)
	api.POST("/inventory/bulk-update", h.BulkUpdateInventory)

	// Jobs
	api.GET("/jobs/:id", h.GetJob)

	// Kanban Inbound
	api.GET("/kanban/inbound", h.ListKanbanInbound)
	api.POST("/kanban/inbound", h.CreateKanbanInbound)
	api.POST("/kanban/inbound/:id/move", h.MoveKanbanInbound)

	// Kanban Outbound
	api.GET("/kanban/outbound", h.ListKanbanOutbound)
	api.POST("/kanban/outbound", h.CreateKanbanOutbound)
	api.POST("/kanban/outbound/:id/move", h.MoveKanbanOutbound)

	// Import
	api.POST("/import/products", h.ImportFile("products"))
	api.POST("/import/inventory", h.ImportFile("inventory"))
	api.POST("/import/inbound", h.ImportFile("inbound"))
	api.POST("/import/outbound", h.ImportFile("outbound"))
	api.GET("/import/inventory/template", h.DownloadInventoryTemplate)
	api.GET("/import/batches/latest", h.GetImportBatch)
	api.GET("/import/batches/:id", h.GetImportBatch)

	// Dashboard
	api.GET("/dashboard/summary", h.DashboardSummary)
	api.GET("/dashboard/charts", h.DashboardCharts)
	api.GET("/dashboard/zero-sales", h.ZeroSales)
	api.GET("/dashboard/restock-alerts", h.RestockAlerts)

	// Inventory extras
	api.GET("/inventory/lots", h.InventoryLots)
	api.GET("/inventory/alerts", h.InventoryAlerts)

	// Orders
	api.GET("/orders", h.ListOrders)
	api.POST("/orders", h.CreateOrder)

	// Thresholds
	api.GET("/thresholds", h.GetThresholds)
	api.POST("/thresholds", h.SaveThreshold)

	// Serve static frontend (production)
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "web/dist"
	}
	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
		r.StaticFS("/assets", http.Dir(filepath.Join(staticDir, "assets")))
		r.NoRoute(func(c *gin.Context) {
			c.File(filepath.Join(staticDir, "index.html"))
		})
	}

	return r
}
