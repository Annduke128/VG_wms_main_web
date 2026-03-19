import {
	Bar,
	BarChart,
	CartesianGrid,
	Legend,
	Line,
	LineChart,
	ResponsiveContainer,
	Tooltip,
	XAxis,
	YAxis,
} from "recharts";
import type { DashboardCharts } from "../types/dashboard";

interface ChartsProps {
	data: DashboardCharts | null;
	loading: boolean;
}

export function DashboardChartsView({ data, loading }: ChartsProps) {
	if (loading || !data) {
		return (
			<div style={{ color: "#888", padding: 20 }}>Đang tải biểu đồ...</div>
		);
	}

	const weekData = (data.in_out_by_week || []).map((w) => ({
		...w,
		week_label: new Date(w.week_start).toLocaleDateString("vi-VN", {
			day: "2-digit",
			month: "2-digit",
		}),
	}));

	const optimalData = (data.inventory_vs_optimal || []).map((item) => ({
		name:
			item.ma_hang.length > 12 ? item.ma_hang.slice(0, 12) + "…" : item.ma_hang,
		"Tồn thực": item.so_ton,
		"Tối ưu": item.optimal_qty,
	}));

	return (
		<div
			style={{
				display: "grid",
				gridTemplateColumns: "1fr 1fr",
				gap: 24,
				marginTop: 24,
			}}
		>
			{/* Chart 1: Inbound/Outbound theo tuần */}
			<div
				style={{
					background: "#fff",
					borderRadius: 10,
					padding: 20,
					boxShadow: "0 1px 4px rgba(0,0,0,0.08)",
				}}
			>
				<h3 style={{ margin: "0 0 16px", fontSize: 15, color: "#333" }}>
					Nhập / Xuất theo tuần (4 tuần)
				</h3>
				<ResponsiveContainer width="100%" height={280}>
					<LineChart data={weekData}>
						<CartesianGrid strokeDasharray="3 3" />
						<XAxis dataKey="week_label" fontSize={12} />
						<YAxis fontSize={12} />
						<Tooltip />
						<Legend />
						<Line
							type="monotone"
							dataKey="total_in"
							name="Nhập"
							stroke="#4CAF50"
							strokeWidth={2}
							dot={{ r: 4 }}
						/>
						<Line
							type="monotone"
							dataKey="total_out"
							name="Xuất"
							stroke="#f44336"
							strokeWidth={2}
							dot={{ r: 4 }}
						/>
					</LineChart>
				</ResponsiveContainer>
			</div>

			{/* Chart 2: Tồn thực vs Tối ưu */}
			<div
				style={{
					background: "#fff",
					borderRadius: 10,
					padding: 20,
					boxShadow: "0 1px 4px rgba(0,0,0,0.08)",
				}}
			>
				<h3 style={{ margin: "0 0 16px", fontSize: 15, color: "#333" }}>
					Tồn thực vs Tối ưu (Top SKU)
				</h3>
				<ResponsiveContainer width="100%" height={280}>
					<BarChart data={optimalData}>
						<CartesianGrid strokeDasharray="3 3" />
						<XAxis
							dataKey="name"
							fontSize={11}
							angle={-20}
							textAnchor="end"
							height={50}
						/>
						<YAxis fontSize={12} />
						<Tooltip />
						<Legend />
						<Bar dataKey="Tồn thực" fill="#2196F3" radius={[4, 4, 0, 0]} />
						<Bar dataKey="Tối ưu" fill="#FF9800" radius={[4, 4, 0, 0]} />
					</BarChart>
				</ResponsiveContainer>
			</div>
		</div>
	);
}
