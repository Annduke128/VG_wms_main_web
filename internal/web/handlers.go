package web

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"wms-v1/internal/domain"
	"wms-v1/internal/service"
)

type Handlers struct {
	Inventory *service.InventoryService
	Kanban    *service.KanbanService
	Import    *service.ImportService
}

func NewHandlers(inv *service.InventoryService, kan *service.KanbanService, imp *service.ImportService) *Handlers {
	return &Handlers{Inventory: inv, Kanban: kan, Import: imp}
}

// --- Inventory Grid ---

func (h *Handlers) InventoryGrid(c *fiber.Ctx) error {
	var req domain.GridRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	resp, err := h.Inventory.GridQuery(c.Context(), req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(resp)
}

func (h *Handlers) UpdateInventoryItem(c *fiber.Ctx) error {
	maHang := c.Params("ma_hang")
	if maHang == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ma_hang is required"})
	}

	var fields map[string]interface{}
	if err := c.BodyParser(&fields); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.Inventory.UpdateItem(c.Context(), maHang, fields); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *Handlers) BulkUpdateInventory(c *fiber.Ctx) error {
	var req domain.BulkUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	jobID, err := h.Inventory.BulkUpdate(c.Context(), req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(202).JSON(domain.BulkUpdateResponse{JobID: jobID})
}

func (h *Handlers) GetJob(c *fiber.Ctx) error {
	jobID := c.Params("id")
	job, err := h.Inventory.GetJob(c.Context(), jobID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "job not found"})
	}
	return c.JSON(job)
}

// --- Kanban Inbound ---

func (h *Handlers) ListKanbanInbound(c *fiber.Ctx) error {
	items, err := h.Kanban.ListInbound(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *Handlers) CreateKanbanInbound(c *fiber.Ctx) error {
	var req domain.CreateKanbanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.MaHang == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ma_hang is required"})
	}

	item, err := h.Kanban.CreateInbound(c.Context(), req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(item)
}

func (h *Handlers) MoveKanbanInbound(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}

	var req domain.MoveKanbanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.Kanban.MoveInbound(c.Context(), id, req.ToStage, req.UserID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// --- Kanban Outbound ---

func (h *Handlers) ListKanbanOutbound(c *fiber.Ctx) error {
	items, err := h.Kanban.ListOutbound(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *Handlers) CreateKanbanOutbound(c *fiber.Ctx) error {
	var req domain.CreateKanbanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.MaHang == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ma_hang is required"})
	}

	item, err := h.Kanban.CreateOutbound(c.Context(), req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(item)
}

func (h *Handlers) MoveKanbanOutbound(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}

	var req domain.MoveKanbanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	negativeAlert, err := h.Kanban.MoveOutbound(c.Context(), id, req.ToStage, req.UserID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	resp := fiber.Map{"status": "ok"}
	if negativeAlert {
		resp["warning"] = "NEGATIVE_STOCK"
		resp["message"] = "Stock went negative after this outbound"
	}

	return c.JSON(resp)
}

// --- Import ---

func (h *Handlers) ImportFile(fileType string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "file is required"})
		}

		// Save to temp
		tempPath := "/tmp/wms_import_" + file.Filename
		if err := c.SaveFile(file, tempPath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to save file"})
		}

		jobID, err := h.Import.EnqueueImport(c.Context(), fileType, tempPath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(202).JSON(fiber.Map{"job_id": jobID})
	}
}
