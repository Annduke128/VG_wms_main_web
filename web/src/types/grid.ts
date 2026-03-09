export interface SortItem {
	colId: string;
	sort: "asc" | "desc";
}

export interface FilterItem {
	filterType: "text" | "number" | "date";
	type: "contains" | "equals" | "startsWith" | "endsWith" | "inRange" | "set";
	filter?: string | number;
	filterTo?: string | number;
	values?: string[];
}

export interface GridRequest {
	startRow: number;
	endRow: number;
	sortModel: SortItem[];
	filterModel: Record<string, FilterItem>;
}

export interface GridResponse<T = Record<string, unknown>> {
	rowsData: T[];
	totalRowCount: number;
}

export interface BulkUpdateItem {
	ma_hang: string;
	fields: Record<string, unknown>;
}

export interface BulkUpdateRequest {
	updates: BulkUpdateItem[];
}

export interface BulkUpdateResponse {
	job_id: string;
}

export interface AsyncJob {
	job_id: string;
	job_type: string;
	status: "pending" | "running" | "completed" | "failed";
	payload: string;
	result: string;
	error: string;
	created_at: string;
	updated_at: string;
}
