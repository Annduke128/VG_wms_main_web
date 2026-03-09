package web

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupRoutes(app *fiber.App, h *Handlers) {
	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Content-Type, Authorization",
		AllowMethods: "GET, POST, PATCH, DELETE, OPTIONS",
	}))

	api := app.Group("/api")

	// Inventory
	api.Post("/inventory/grid", h.InventoryGrid)
	api.Patch("/inventory/:ma_hang", h.UpdateInventoryItem)
	api.Post("/inventory/bulk-update", h.BulkUpdateInventory)

	// Jobs
	api.Get("/jobs/:id", h.GetJob)

	// Kanban Inbound
	api.Get("/kanban/inbound", h.ListKanbanInbound)
	api.Post("/kanban/inbound", h.CreateKanbanInbound)
	api.Post("/kanban/inbound/:id/move", h.MoveKanbanInbound)

	// Kanban Outbound
	api.Get("/kanban/outbound", h.ListKanbanOutbound)
	api.Post("/kanban/outbound", h.CreateKanbanOutbound)
	api.Post("/kanban/outbound/:id/move", h.MoveKanbanOutbound)

	// Import
	api.Post("/import/products", h.ImportFile("products"))
	api.Post("/import/inventory", h.ImportFile("inventory"))
	api.Post("/import/inbound", h.ImportFile("inbound"))
	api.Post("/import/outbound", h.ImportFile("outbound"))
}
