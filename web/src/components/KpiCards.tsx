import type { DashboardSummary } from "../types/dashboard";

interface KpiCardsProps {
	data: DashboardSummary | null;
	loading: boolean;
}

const KPI_CONFIG = [
	{
		key: "sku_count" as const,
		label: "Tổng SKU",
		color: "#6b7efa",
		format: (v: number) => v.toLocaleString("vi-VN"),
	},
	{
		key: "sku_ton_lau" as const,
		label: "SKU tồn lâu",
		color: "#e5a04b",
		format: (v: number) => v.toLocaleString("vi-VN"),
	},
	{
		key: "sku_thieu_hang" as const,
		label: "SKU thiếu hàng",
		color: "#e06363",
		format: (v: number) => v.toLocaleString("vi-VN"),
	},
	{
		key: "tong_tien_hang" as const,
		label: "Tổng tiền hàng",
		color: "#5bb98c",
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
						borderRadius: 8,
						padding: "18px 20px",
						border: "1px solid #e8eaed",
						borderTop: `2px solid ${kpi.color}`,
					}}
				>
					<div style={{ fontSize: 12, color: "#7a7f8e", marginBottom: 6 }}>
						{kpi.label}
					</div>
					<div style={{ fontSize: 24, fontWeight: 700, color: "#1e2330" }}>
						{loading || !data ? "—" : kpi.format(data[kpi.key])}
					</div>
				</div>
			))}
		</div>
	);
}
