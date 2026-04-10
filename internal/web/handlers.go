package web

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"wms-v1/internal/domain"
	"wms-v1/internal/importer"
	"wms-v1/internal/queue"
	"wms-v1/internal/repo"
	"wms-v1/internal/service"
)

type Handlers struct {
	Inventory *service.InventoryService
	Import    *service.ImportService
	Orders    *service.OrderService
	Dashboard *service.DashboardService
	Combo     *service.ComboService
	Queue     *queue.RedisQueue
	Repo      *repo.PostgresRepo
}

func NewHandlers(inv *service.InventoryService, imp *service.ImportService, ord *service.OrderService, dash *service.DashboardService, combo *service.ComboService, q *queue.RedisQueue, r *repo.PostgresRepo) *Handlers {
	return &Handlers{Inventory: inv, Import: imp, Orders: ord, Dashboard: dash, Combo: combo, Queue: q, Repo: r}
}

// getWarehouseID extracts warehouse_id from query params (required).
func getWarehouseID(c *gin.Context) (int64, bool) {
	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
		return 0, false
	}
	warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
	if err != nil || warehouseID <= 0 {
		c.JSON(400, gin.H{"error": "invalid warehouse_id"})
		return 0, false
	}
	return warehouseID, true
}

// --- Inventory Grid ---

func (h *Handlers) InventoryGrid(c *gin.Context) {
	var req domain.GridRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	if req.WarehouseID <= 0 {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
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
	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
		return
	}
	warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
	if err != nil || warehouseID <= 0 {
		c.JSON(400, gin.H{"error": "invalid warehouse_id"})
		return
	}

	var fields map[string]interface{}
	if err := c.ShouldBindJSON(&fields); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.Inventory.UpdateItem(c.Request.Context(), maHang, warehouseID, fields); err != nil {
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
	if req.WarehouseID <= 0 {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
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

// --- Import ---

func (h *Handlers) ImportFile(fileType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(400, gin.H{"error": "file is required"})
			return
		}

		// Save to shared upload dir (accessible by both API and Worker containers)
		uploadDir := os.Getenv("UPLOAD_DIR")
		if uploadDir == "" {
			uploadDir = "/app/uploads"
		}
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			c.JSON(500, gin.H{"error": "failed to create upload dir: " + err.Error()})
			return
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to open uploaded file: " + err.Error()})
			return
		}
		defer src.Close()

		ext := filepath.Ext(file.Filename)
		tmpFile, err := os.CreateTemp(uploadDir, "wms_import_*"+ext)
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
	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
		return
	}
	warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
	if err != nil || warehouseID <= 0 {
		c.JSON(400, gin.H{"error": "invalid warehouse_id"})
		return
	}
	summary, err := h.Dashboard.GetSummary(c.Request.Context(), warehouseID)
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
	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
		return
	}
	warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
	if err != nil || warehouseID <= 0 {
		c.JSON(400, gin.H{"error": "invalid warehouse_id"})
		return
	}

	charts, err := h.Dashboard.GetCharts(c.Request.Context(), warehouseID, weeks)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, charts)
}

func (h *Handlers) InventoryAlerts(c *gin.Context) {
	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
		return
	}
	warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
	if err != nil || warehouseID <= 0 {
		c.JSON(400, gin.H{"error": "invalid warehouse_id"})
		return
	}
	alerts, err := h.Dashboard.GetAlerts(c.Request.Context(), warehouseID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, alerts)
}

func (h *Handlers) ZeroSales(c *gin.Context) {
	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
		return
	}
	warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
	if err != nil || warehouseID <= 0 {
		c.JSON(400, gin.H{"error": "invalid warehouse_id"})
		return
	}
	items, err := h.Dashboard.GetZeroSales(c.Request.Context(), warehouseID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, items)
}

func (h *Handlers) RestockAlerts(c *gin.Context) {
	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
		return
	}
	warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
	if err != nil || warehouseID <= 0 {
		c.JSON(400, gin.H{"error": "invalid warehouse_id"})
		return
	}
	items, err := h.Dashboard.GetRestockAlerts(c.Request.Context(), warehouseID)
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
	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
		return
	}
	warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
	if err != nil || warehouseID <= 0 {
		c.JSON(400, gin.H{"error": "invalid warehouse_id"})
		return
	}

	lots, err := h.Orders.GetLots(c.Request.Context(), maHang, warehouseID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, lots)
}

// --- Orders ---

func (h *Handlers) ListOrders(c *gin.Context) {
	var f domain.OrderFilter
	f.OrderType = c.Query("type") // "in", "out", or "" for all
	if wid := c.Query("warehouse_id"); wid != "" {
		if parsed, err := strconv.ParseInt(wid, 10, 64); err == nil && parsed > 0 {
			f.WarehouseID = parsed
		}
	}

	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	f.Limit = 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			f.Limit = parsed
		}
	}
	f.Offset = (page - 1) * f.Limit

	// Date filters: dd/mm/yyyy or yyyy-mm-dd
	if df := c.Query("date_from"); df != "" {
		if t, err := parseDateParam(df); err == nil {
			f.DateFrom = t
		}
	}
	if dt := c.Query("date_to"); dt != "" {
		if t, err := parseDateParam(dt); err == nil {
			// End of day
			f.DateTo = t.Add(24*time.Hour - time.Nanosecond)
		}
	}
	// Month filter: mm/yyyy — overrides date_from/date_to
	if m := c.Query("month"); m != "" {
		if t, err := parseMonthParam(m); err == nil {
			f.DateFrom = t
			f.DateTo = t.AddDate(0, 1, 0).Add(-time.Nanosecond)
		}
	}

	f.MaBu = c.Query("ma_bu")
	f.MaNhomHang = c.Query("ma_nhom_hang")

	orders, total, err := h.Orders.ListOrders(c.Request.Context(), f)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": orders, "total": total, "page": page, "limit": f.Limit})
}

// parseDateParam parses dd/mm/yyyy or yyyy-mm-dd
func parseDateParam(s string) (time.Time, error) {
	layouts := []string{"02/01/2006", "2006-01-02", "02-01-2006"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date: %s", s)
}

// parseMonthParam parses mm/yyyy
func parseMonthParam(s string) (time.Time, error) {
	layouts := []string{"01/2006", "2006-01"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid month: %s", s)
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
		if wid, ok := raw["warehouse_id"].(float64); ok {
			req.WarehouseID = int64(wid)
		}
		if qty, ok := raw["so_luong"].(float64); ok {
			req.SoLuong = qty
		}

		if req.MaHang == "" || req.BatchCode == "" || req.SoLuong <= 0 || req.WarehouseID <= 0 {
			c.JSON(400, gin.H{"error": "ma_hang, batch_code, warehouse_id, and so_luong > 0 are required"})
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
		if wid, ok := raw["warehouse_id"].(float64); ok {
			req.WarehouseID = int64(wid)
		}
		if qty, ok := raw["so_luong"].(float64); ok {
			req.SoLuong = qty
		}

		if req.MaHang == "" || req.SoLuong <= 0 || req.WarehouseID <= 0 {
			c.JSON(400, gin.H{"error": "ma_hang, warehouse_id, and so_luong > 0 are required"})
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
	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
		return
	}
	warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
	if err != nil || warehouseID <= 0 {
		c.JSON(400, gin.H{"error": "invalid warehouse_id"})
		return
	}

	thresholds, err := h.Dashboard.GetThresholds(c.Request.Context(), maHang, warehouseID)
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

// --- Recalc All Metrics ---

func (h *Handlers) RecalcAllMetrics(c *gin.Context) {
	jobID := uuid.New().String()

	if err := h.Import.Repo.CreateAsyncJob(c.Request.Context(), jobID, "recalc_all_metrics", "{}"); err != nil {
		c.JSON(500, gin.H{"error": "failed to create job: " + err.Error()})
		return
	}

	job := queue.Job{
		ID:      jobID,
		Type:    "recalc_all",
		Payload: json.RawMessage(`{}`),
	}
	if err := h.Queue.Enqueue(c.Request.Context(), queue.QueueRecalc, job); err != nil {
		c.JSON(500, gin.H{"error": "failed to enqueue job: " + err.Error()})
		return
	}

	c.JSON(202, gin.H{"job_id": jobID})
}

// --- Reset All Data ---

func (h *Handlers) ResetAllData(c *gin.Context) {
	var req struct {
		ConfirmText string `json:"confirm_text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	if req.ConfirmText != "RESET ALL" {
		c.JSON(400, gin.H{"error": "Vui lòng nhập đúng cụm từ xác nhận: RESET ALL"})
		return
	}

	if err := h.Import.Repo.ResetAllData(c.Request.Context()); err != nil {
		c.JSON(500, gin.H{"error": "reset failed: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "ok", "message": "Đã xóa toàn bộ dữ liệu thành công"})
}

// --- Inventory Filter Options ---

func (h *Handlers) InventoryFilterOptions(c *gin.Context) {
	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
		return
	}
	warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
	if err != nil || warehouseID <= 0 {
		c.JSON(400, gin.H{"error": "invalid warehouse_id"})
		return
	}
	opts, err := h.Inventory.GetFilterOptions(c.Request.Context(), warehouseID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, opts)
}

// --- Export Inventory ---

func (h *Handlers) ExportInventory(c *gin.Context) {
	var req struct {
		MaHang      []string                     `json:"ma_hang"`
		Columns     []string                     `json:"columns"`
		FilterModel map[string]domain.FilterItem `json:"filter_model"`
		WarehouseID int64                        `json:"warehouse_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	if req.WarehouseID <= 0 {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
		return
	}

	// Filter to only valid columns
	var cols []string
	for _, col := range req.Columns {
		if importer.ValidExportColumns[col] {
			cols = append(cols, col)
		}
	}
	if len(cols) == 0 {
		c.JSON(400, gin.H{"error": "no valid columns specified"})
		return
	}

	// Query data
	rows, err := h.Inventory.ExportRows(c.Request.Context(), req.MaHang, req.FilterModel, req.WarehouseID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Build Excel
	data, err := importer.BuildExportExcel(rows, cols)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate Excel: " + err.Error()})
		return
	}

	filename := fmt.Sprintf("BaoCaoTonKho_%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// --- Combo Master ---

func (h *Handlers) ListComboMasters(c *gin.Context) {
	items, err := h.Combo.ListComboMasters(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, items)
}

func (h *Handlers) GetComboDetail(c *gin.Context) {
	maCombo := c.Param("ma_combo")
	detail, err := h.Combo.GetComboDetail(c.Request.Context(), maCombo)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, detail)
}

func (h *Handlers) SaveComboMaster(c *gin.Context) {
	var req domain.SaveComboMasterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.Combo.SaveComboMaster(c.Request.Context(), req); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

func (h *Handlers) DeleteComboMaster(c *gin.Context) {
	maCombo := c.Param("ma_combo")
	if err := h.Combo.DeleteComboMaster(c.Request.Context(), maCombo); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

// --- Combo Operations ---

func (h *Handlers) CreateCombo(c *gin.Context) {
	var req domain.CreateComboRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	txn, err := h.Combo.CreateCombo(c.Request.Context(), req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, txn)
}

func (h *Handlers) CancelCombo(c *gin.Context) {
	var req domain.CancelComboRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	txn, err := h.Combo.CancelCombo(c.Request.Context(), req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, txn)
}

func (h *Handlers) ComboOut(c *gin.Context) {
	var req domain.ComboOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	txn, err := h.Combo.ComboOut(c.Request.Context(), req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, txn)
}

func (h *Handlers) ComboReturn(c *gin.Context) {
	var req domain.ComboReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	txn, err := h.Combo.ComboReturn(c.Request.Context(), req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, txn)
}

// --- Combo Inventory ---

func (h *Handlers) GetComboInventory(c *gin.Context) {
	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
		return
	}
	warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
	if err != nil || warehouseID <= 0 {
		c.JSON(400, gin.H{"error": "invalid warehouse_id"})
		return
	}
	items, err := h.Combo.GetComboInventory(c.Request.Context(), warehouseID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, items)
}

// --- Combo Transactions ---

func (h *Handlers) ListComboTransactions(c *gin.Context) {
	maCombo := c.Query("ma_combo")
	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
		return
	}
	warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
	if err != nil || warehouseID <= 0 {
		c.JSON(400, gin.H{"error": "invalid warehouse_id"})
		return
	}
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

	items, total, err := h.Combo.ListComboTransactions(c.Request.Context(), maCombo, page, limit, warehouseID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": items, "total": total, "page": page, "limit": limit})
}

// --- Accessories ---

func (h *Handlers) ListAccessories(c *gin.Context) {
	items, err := h.Combo.ListAccessories(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, items)
}

func (h *Handlers) CreateAccessory(c *gin.Context) {
	var req domain.CreateAccessoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.Combo.CreateAccessory(c.Request.Context(), req); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"status": "ok"})
}

func (h *Handlers) GetAccessoryInventory(c *gin.Context) {
	warehouseIDStr := c.Query("warehouse_id")
	if warehouseIDStr == "" {
		c.JSON(400, gin.H{"error": "warehouse_id is required"})
		return
	}
	warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
	if err != nil || warehouseID <= 0 {
		c.JSON(400, gin.H{"error": "invalid warehouse_id"})
		return
	}
	items, err := h.Combo.GetAccessoryInventory(c.Request.Context(), warehouseID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, items)
}

func (h *Handlers) AccessoryStockIn(c *gin.Context) {
	var req domain.AccessoryStockInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.Combo.AccessoryStockIn(c.Request.Context(), req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

// ─── Warehouse CRUD ─────────────────────────────────────────────────────────

func (h *Handlers) ListWarehouses(c *gin.Context) {
	warehouses, err := h.Repo.ListWarehouses(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, warehouses)
}

func (h *Handlers) GetWarehouse(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid warehouse id"})
		return
	}
	w, err := h.Repo.GetWarehouse(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, w)
}

func (h *Handlers) CreateWarehouse(c *gin.Context) {
	var req domain.CreateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	w, err := h.Repo.CreateWarehouse(c.Request.Context(), req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	// Initialize inventory records for the new warehouse
	if err := h.Repo.InitWarehouseInventory(c.Request.Context(), w.ID); err != nil {
		c.JSON(500, gin.H{"error": "warehouse created but inventory init failed: " + err.Error()})
		return
	}
	c.JSON(201, w)
}

func (h *Handlers) UpdateWarehouse(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid warehouse id"})
		return
	}
	var req domain.UpdateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	w, err := h.Repo.UpdateWarehouse(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, w)
}
