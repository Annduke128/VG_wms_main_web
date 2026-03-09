package domain

// GridRequest is the POST body for server-side grid operations
type GridRequest struct {
	StartRow    int                   `json:"startRow"`
	EndRow      int                   `json:"endRow"`
	SortModel   []SortItem            `json:"sortModel"`
	FilterModel map[string]FilterItem `json:"filterModel"`
}

type SortItem struct {
	ColID string `json:"colId"`
	Sort  string `json:"sort"` // "asc" or "desc"
}

type FilterItem struct {
	FilterType string      `json:"filterType"` // "text", "number", "date"
	Type       string      `json:"type"`       // "contains", "equals", "startsWith", "endsWith", "inRange", "set"
	Filter     interface{} `json:"filter"`
	FilterTo   interface{} `json:"filterTo"` // for inRange
	Values     []string    `json:"values"`   // for set filter
}

// GridResponse is the standard response for grid queries
type GridResponse struct {
	RowsData      interface{} `json:"rowsData"`
	TotalRowCount int64       `json:"totalRowCount"`
}

// BulkUpdateRequest for async bulk operations
type BulkUpdateRequest struct {
	Updates []BulkUpdateItem `json:"updates"`
}

type BulkUpdateItem struct {
	MaHang string                 `json:"ma_hang"`
	Fields map[string]interface{} `json:"fields"`
}

// BulkUpdateResponse returns job ID for polling
type BulkUpdateResponse struct {
	JobID string `json:"job_id"`
}
