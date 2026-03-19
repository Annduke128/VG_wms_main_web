import { useCallback, useEffect, useState } from "react";
import { api } from "../api/client";
import { DashboardChartsView } from "../components/Charts";
import { KpiCards } from "../components/KpiCards";
import type {
	AlertItem,
	DashboardCharts,
	DashboardSummary,
} from "../types/dashboard";

export function Overview() {
	const [summary, setSummary] = useState<DashboardSummary | null>(null);
	const [charts, setCharts] = useState<DashboardCharts | null>(null);
	const [alerts, setAlerts] = useState<AlertItem[]>([]);
	const [loading, setLoading] = useState(true);

	const fetchData = useCallback(async () => {
		setLoading(true);
		try {
			const [s, c, a] = await Promise.all([
				api.dashboardSummary() as Promise<DashboardSummary>,
				api.dashboardCharts(4) as Promise<DashboardCharts>,
				api.inventoryAlerts() as Promise<AlertItem[]>,
			]);
			setSummary(s);
			setCharts(c);
			setAlerts(a || []);
		} catch (err) {
			console.error("Dashboard fetch error:", err);
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		fetchData();
	}, [fetchData]);

	return (
		<div>
			<h2 style={{ margin: "0 0 20px", fontSize: 20, color: "#222" }}>
				Tổng quan
			</h2>

			<KpiCards data={summary} loading={loading} />
			<DashboardChartsView data={charts} loading={loading} />

			{/* Alerts section */}
			{alerts.length > 0 && (
				<div style={{ marginTop: 24 }}>
					<h3 style={{ margin: "0 0 12px", fontSize: 15, color: "#333" }}>
						Cảnh báo
					</h3>
					<div style={{ display: "grid", gap: 8 }}>
						{alerts.slice(0, 20).map((alert, i) => (
							<div
								key={`${alert.ma_hang}-${alert.alert_type}-${i}`}
								style={{
									display: "flex",
									alignItems: "center",
									gap: 12,
									background: "#fff",
									borderRadius: 8,
									padding: "10px 16px",
									boxShadow: "0 1px 3px rgba(0,0,0,0.06)",
									borderLeft: `3px solid ${alert.alert_type === "ton_lau" ? "#FF9800" : "#f44336"}`,
								}}
							>
								<span style={{ fontSize: 13, fontWeight: 600, minWidth: 120 }}>
									{alert.ma_hang}
								</span>
								<span style={{ fontSize: 13, color: "#555", flex: 1 }}>
									{alert.ten_san_pham}
								</span>
								<span
									style={{
										fontSize: 12,
										padding: "2px 8px",
										borderRadius: 4,
										background:
											alert.alert_type === "ton_lau" ? "#FFF3E0" : "#FFEBEE",
										color:
											alert.alert_type === "ton_lau" ? "#E65100" : "#C62828",
									}}
								>
									{alert.alert_type === "ton_lau" ? "Tồn lâu" : "Thiếu hàng"}
								</span>
								<span style={{ fontSize: 12, color: "#888" }}>
									{alert.message}
								</span>
							</div>
						))}
					</div>
				</div>
			)}
		</div>
	);
}
