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
			<div style={{ color: "#7a7f8e", padding: 20, fontSize: 13 }}>
				Đang tải biểu đồ...
			</div>
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
				gap: 20,
				marginTop: 20,
			}}
		>
			<div
				style={{
					background: "#fff",
					borderRadius: 8,
					padding: 20,
					border: "1px solid #e8eaed",
				}}
			>
				<h3
					style={{
						margin: "0 0 16px",
						fontSize: 13,
						fontWeight: 600,
						color: "#3a3f4b",
					}}
				>
					Nhập / Xuất theo tuần
				</h3>
				<ResponsiveContainer width="100%" height={260}>
					<LineChart data={weekData}>
						<CartesianGrid strokeDasharray="3 3" stroke="#eee" />
						<XAxis
							dataKey="week_label"
							fontSize={11}
							tick={{ fill: "#7a7f8e" }}
						/>
						<YAxis fontSize={11} tick={{ fill: "#7a7f8e" }} />
						<Tooltip />
						<Legend />
						<Line
							type="monotone"
							dataKey="total_in"
							name="Nhập"
							stroke="#5bb98c"
							strokeWidth={2}
							dot={{ r: 3 }}
						/>
						<Line
							type="monotone"
							dataKey="total_out"
							name="Xuất"
							stroke="#e06363"
							strokeWidth={2}
							dot={{ r: 3 }}
						/>
					</LineChart>
				</ResponsiveContainer>
			</div>

			<div
				style={{
					background: "#fff",
					borderRadius: 8,
					padding: 20,
					border: "1px solid #e8eaed",
				}}
			>
				<h3
					style={{
						margin: "0 0 16px",
						fontSize: 13,
						fontWeight: 600,
						color: "#3a3f4b",
					}}
				>
					Tồn thực vs Tối ưu
				</h3>
				<ResponsiveContainer width="100%" height={260}>
					<BarChart data={optimalData}>
						<CartesianGrid strokeDasharray="3 3" stroke="#eee" />
						<XAxis
							dataKey="name"
							fontSize={10}
							angle={-20}
							textAnchor="end"
							height={50}
							tick={{ fill: "#7a7f8e" }}
						/>
						<YAxis fontSize={11} tick={{ fill: "#7a7f8e" }} />
						<Tooltip />
						<Legend />
						<Bar dataKey="Tồn thực" fill="#6b7efa" radius={[3, 3, 0, 0]} />
						<Bar dataKey="Tối ưu" fill="#e5a04b" radius={[3, 3, 0, 0]} />
					</BarChart>
				</ResponsiveContainer>
			</div>
		</div>
	);
}
