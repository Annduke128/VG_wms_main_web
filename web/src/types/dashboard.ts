export interface DashboardSummary {
	sku_count: number;
	sku_ton_lau: number;
	sku_thieu_hang: number;
	tong_tien_hang: number;
}

export interface InOutWeekData {
	week_start: string;
	total_in: number;
	total_out: number;
}

export interface InventoryVsOptimalItem {
	ma_hang: string;
	ten_san_pham: string;
	so_ton: number;
	optimal_qty: number;
}

export interface DashboardCharts {
	in_out_by_week: InOutWeekData[];
	inventory_vs_optimal: InventoryVsOptimalItem[];
}

export interface AlertItem {
	ma_hang: string;
	ten_san_pham: string;
	alert_type: string;
	so_ton: number;
	threshold_value: number;
	message: string;
}

export interface InventoryLot {
	id: number;
	ma_hang: string;
	batch_code: string;
	received_at: string;
	qty_in: number;
	qty_out: number;
	qty_remaining: number;
	created_at: string;
}

export interface OrderListItem {
	id: number;
	type: string;
	ma_hang: string;
	ten_san_pham: string;
	batch_code: string;
	so_luong: number;
	ngay_nhan_hang: string;
}

export interface ZeroSalesItem {
	ma_hang: string;
	ten_san_pham: string;
	so_ton: number;
	luong_ban_binh_quan_ngay: number;
	latest_outbound_month: string;
}

export interface RestockAlertItem {
	ma_hang: string;
	ten_san_pham: string;
	so_ton: number;
	days_since_last_out: number;
	last_outbound_date: string;
}

export interface ThresholdEntry {
	id: number;
	ma_hang: string;
	min_qty: number;
	optimal_qty: number;
	max_age_days: number;
	source: string;
	model_version: string;
	effective_from: string;
	effective_to: string | null;
	created_at: string;
}
