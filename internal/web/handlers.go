package web

import (
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"

	"wms-v1/internal/domain"
	"wms-v1/internal/importer"
	"wms-v1/internal/service"
)

type Handlers struct {
	Inventory *service.InventoryService
	Kanban    *service.KanbanService
	Import    *service.ImportService
	Orders    *service.OrderService
	Dashboard *service.DashboardService
}

func NewHandlers(inv *service.InventoryService, kan *service.KanbanService, imp *service.ImportService, ord *service.OrderService, dash *service.DashboardService) *Handlers {
	return &Handlers{Inventory: inv, Kanban: kan, Import: imp, Orders: ord, Dashboard: dash}
}

// --- Inventory Grid ---

func (h *Handlers) InventoryGrid(c *gin.Context) {
	var req domain.GridRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.Inventory.GridQuery(c.Request.Context(), req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, resp)
}

func (h *Handlers) UpdateInventoryItem(c *gin.Context) {
	maHang := c.Param("ma_hang")
	if maHang == "" {
		c.JSON(400, gin.H{"error": "ma_hang is required"})
		return
	}

	var fields map[string]interface{}
	if err := c.ShouldBindJSON(&fields); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.Inventory.UpdateItem(c.Request.Context(), maHang, fields); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "ok"})
}

func (h *Handlers) BulkUpdateInventory(c *gin.Context) {
	var req domain.BulkUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	jobID, err := h.Inventory.BulkUpdate(c.Request.Context(), req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(202, domain.BulkUpdateResponse{JobID: jobID})
}

func (h *Handlers) GetJob(c *gin.Context) {
	jobID := c.Param("id")
	job, err := h.Inventory.GetJob(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(404, gin.H{"error": "job not found"})
		return
	}
	c.JSON(200, job)
}

// --- Kanban Inbound ---

func (h *Handlers) ListKanbanInbound(c *gin.Context) {
	items, err := h.Kanban.ListInbound(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, items)
}

func (h *Handlers) CreateKanbanInbound(c *gin.Context) {
	var req domain.CreateKanbanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	if req.MaHang == "" {
		c.JSON(400, gin.H{"error": "ma_hang is required"})
		return
	}

	item, err := h.Kanban.CreateInbound(c.Request.Context(), req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, item)
}

func (h *Handlers) MoveKanbanInbound(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	var req domain.MoveKanbanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.Kanban.MoveInbound(c.Request.Context(), id, req.ToStage, req.UserID); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "ok"})
}

// --- Kanban Outbound ---

func (h *Handlers) ListKanbanOutbound(c *gin.Context) {
	items, err := h.Kanban.ListOutbound(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, items)
}

func (h *Handlers) CreateKanbanOutbound(c *gin.Context) {
	var req domain.CreateKanbanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	if req.MaHang == "" {
		c.JSON(400, gin.H{"error": "ma_hang is required"})
		return
	}

	item, err := h.Kanban.CreateOutbound(c.Request.Context(), req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, item)
}

func (h *Handlers) MoveKanbanOutbound(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	var req domain.MoveKanbanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	negativeAlert, err := h.Kanban.MoveOutbound(c.Request.Context(), id, req.ToStage, req.UserID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp := gin.H{"status": "ok"}
	if negativeAlert {
		resp["warning"] = "NEGATIVE_STOCK"
		resp["message"] = "Stock went negative after this outbound"
	}

	c.JSON(200, resp)
}

// --- Import ---

func (h *Handlers) ImportFile(fileType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(400, gin.H{"error": "file is required"})
			return
		}

		// Save to temp file (avoid Gin's SaveUploadedFile which tries to chmod /tmp)
		src, err := file.Open()
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to open uploaded file: " + err.Error()})
			return
		}
		defer src.Close()

		ext := filepath.Ext(file.Filename)
		tmpFile, err := os.CreateTemp("", "wms_import_*"+ext)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to create temp file: " + err.Error()})
			return
		}

		if _, err := io.Copy(tmpFile, src); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			c.JSON(500, gin.H{"error": "failed to write temp file: " + err.Error()})
			return
		}
		tmpFile.Close()
		tempPath := tmpFile.Name()

		jobID, err := h.Import.EnqueueImport(c.Request.Context(), fileType, tempPath)
		if err != nil {
			os.Remove(tempPath)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(202, gin.H{"job_id": jobID})
	}
}

// --- Dashboard ---

func (h *Handlers) DashboardSummary(c *gin.Context) {
	summary, err := h.Dashboard.GetSummary(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, summary)
}

func (h *Handlers) DashboardCharts(c *gin.Context) {
	weeks := 4
	if w := c.Query("weeks"); w != "" {
		if parsed, err := strconv.Atoi(w); err == nil && parsed > 0 {
			weeks = parsed
		}
	}

	charts, err := h.Dashboard.GetCharts(c.Request.Context(), weeks)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, charts)
}

func (h *Handlers) InventoryAlerts(c *gin.Context) {
	alerts, err := h.Dashboard.GetAlerts(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, alerts)
}

func (h *Handlers) ZeroSales(c *gin.Context) {
	items, err := h.Dashboard.GetZeroSales(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, items)
}

func (h *Handlers) RestockAlerts(c *gin.Context) {
	items, err := h.Dashboard.GetRestockAlerts(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, items)
}

func (h *Handlers) InventoryLots(c *gin.Context) {
	maHang := c.Query("ma_hang")
	if maHang == "" {
		c.JSON(400, gin.H{"error": "ma_hang query parameter is required"})
		return
	}

	lots, err := h.Orders.GetLots(c.Request.Context(), maHang)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, lots)
}

// --- Orders ---

func (h *Handlers) ListOrders(c *gin.Context) {
	orderType := c.Query("type") // "in", "out", or "" for all
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}
	offset := (page - 1) * limit

	orders, total, err := h.Orders.ListOrders(c.Request.Context(), orderType, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": orders, "total": total, "page": page, "limit": limit})
}

func (h *Handlers) CreateOrder(c *gin.Context) {
	// Determine order type from body
	var raw map[string]interface{}
	if err := c.ShouldBindJSON(&raw); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	orderType, _ := raw["type"].(string)

	switch orderType {
	case "in", "inbound":
		var req domain.CreateInboundRequest
		// Re-bind from raw data
		c.Request.Body = nil // already consumed, use raw
		req.MaHang, _ = raw["ma_hang"].(string)
		req.BatchCode, _ = raw["batch_code"].(string)
		if qty, ok := raw["so_luong"].(float64); ok {
			req.SoLuong = qty
		}

		if req.MaHang == "" || req.BatchCode == "" || req.SoLuong <= 0 {
			c.JSON(400, gin.H{"error": "ma_hang, batch_code, and so_luong > 0 are required"})
			return
		}

		result, err := h.Orders.CreateInbound(c.Request.Context(), req)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(201, result)

	case "out", "outbound":
		var req domain.CreateOutboundRequest
		req.MaHang, _ = raw["ma_hang"].(string)
		if qty, ok := raw["so_luong"].(float64); ok {
			req.SoLuong = qty
		}

		if req.MaHang == "" || req.SoLuong <= 0 {
			c.JSON(400, gin.H{"error": "ma_hang and so_luong > 0 are required"})
			return
		}

		result, err := h.Orders.CreateOutbound(c.Request.Context(), req)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(201, result)

	default:
		c.JSON(400, gin.H{"error": "type must be 'in' or 'out'"})
	}
}

// --- Thresholds ---

func (h *Handlers) GetThresholds(c *gin.Context) {
	maHang := c.Query("ma_hang")
	if maHang == "" {
		c.JSON(400, gin.H{"error": "ma_hang query parameter is required"})
		return
	}

	thresholds, err := h.Dashboard.GetThresholds(c.Request.Context(), maHang)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, thresholds)
}

func (h *Handlers) GetImportBatch(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		// No ID → get latest
		batch, err := h.Import.Repo.GetLatestImportBatch(c.Request.Context())
		if err != nil {
			c.JSON(404, gin.H{"error": "no import batches found"})
			return
		}
		c.JSON(200, batch)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid batch id"})
		return
	}

	batch, err := h.Import.Repo.GetImportBatch(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "batch not found"})
		return
	}
	c.JSON(200, batch)
}

// --- Template Download ---

func (h *Handlers) DownloadInventoryTemplate(c *gin.Context) {
	buf, err := importer.BuildInventoryTemplate()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate template: " + err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=MauTonKho.xlsx")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf)
}

func (h *Handlers) SaveThreshold(c *gin.Context) {
	var req domain.ThresholdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	threshold, err := h.Dashboard.SaveThreshold(c.Request.Context(), req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, threshold)
}
