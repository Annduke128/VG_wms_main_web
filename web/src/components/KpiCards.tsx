import type { DashboardSummary } from "../types/dashboard";

interface KpiCardsProps {
	data: DashboardSummary | null;
	loading: boolean;
}

const KPI_CONFIG = [
	{
		key: "sku_count" as const,
		label: "Tổng SKU",
		color: "#2196F3",
		format: (v: number) => v.toLocaleString("vi-VN"),
	},
	{
		key: "sku_ton_lau" as const,
		label: "SKU tồn lâu",
		color: "#FF9800",
		format: (v: number) => v.toLocaleString("vi-VN"),
	},
	{
		key: "sku_thieu_hang" as const,
		label: "SKU thiếu hàng",
		color: "#f44336",
		format: (v: number) => v.toLocaleString("vi-VN"),
	},
	{
		key: "tong_tien_hang" as const,
		label: "Tổng tiền hàng",
		color: "#4CAF50",
		format: (v: number) =>
			new Intl.NumberFormat("vi-VN", {
				style: "currency",
				currency: "VND",
			}).format(v),
	},
];

export function KpiCards({ data, loading }: KpiCardsProps) {
	return (
		<div
			style={{
				display: "grid",
				gridTemplateColumns: "repeat(4, 1fr)",
				gap: 16,
			}}
		>
			{KPI_CONFIG.map((kpi) => (
				<div
					key={kpi.key}
					style={{
						background: "#fff",
						borderRadius: 10,
						padding: "20px 24px",
						boxShadow: "0 1px 4px rgba(0,0,0,0.08)",
						borderTop: `3px solid ${kpi.color}`,
					}}
				>
					<div style={{ fontSize: 13, color: "#888", marginBottom: 8 }}>
						{kpi.label}
					</div>
					<div style={{ fontSize: 28, fontWeight: 700, color: "#222" }}>
						{loading || !data ? "—" : kpi.format(data[kpi.key])}
					</div>
				</div>
			))}
		</div>
	);
}
