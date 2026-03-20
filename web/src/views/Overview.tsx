import { useCallback, useEffect, useState } from "react";
import { api } from "../api/client";
import { DashboardChartsView } from "../components/Charts";
import { KpiCards } from "../components/KpiCards";
import type {
	AlertItem,
	DashboardCharts,
	DashboardSummary,
	RestockAlertItem,
	ZeroSalesItem,
} from "../types/dashboard";

export function Overview() {
	const [summary, setSummary] = useState<DashboardSummary | null>(null);
	const [charts, setCharts] = useState<DashboardCharts | null>(null);
	const [alerts, setAlerts] = useState<AlertItem[]>([]);
	const [zeroSales, setZeroSales] = useState<ZeroSalesItem[]>([]);
	const [restockAlerts, setRestockAlerts] = useState<RestockAlertItem[]>([]);
	const [loading, setLoading] = useState(true);

	const fetchData = useCallback(async () => {
		setLoading(true);
		try {
			const [s, c, a, zs, ra] = await Promise.all([
				api.dashboardSummary() as Promise<DashboardSummary>,
				api.dashboardCharts(4) as Promise<DashboardCharts>,
				api.inventoryAlerts() as Promise<AlertItem[]>,
				api.dashboardZeroSales() as Promise<ZeroSalesItem[]>,
				api.dashboardRestockAlerts() as Promise<RestockAlertItem[]>,
			]);
			setSummary(s);
			setCharts(c);
			setAlerts(a || []);
			setZeroSales(zs || []);
			setRestockAlerts(ra || []);
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
			<h2
				style={{
					margin: "0 0 20px",
					fontSize: 16,
					fontWeight: 600,
					color: "#1e2330",
				}}
			>
				Tổng quan
			</h2>

			<KpiCards data={summary} loading={loading} />
			<DashboardChartsView data={charts} loading={loading} />

			{/* Zero Sales Section */}
			{zeroSales.length > 0 && (
				<div style={{ marginTop: 20 }}>
					<h3
						style={{
							margin: "0 0 12px",
							fontSize: 13,
							fontWeight: 600,
							color: "#3a3f4b",
						}}
					>
						Không có doanh số (LBBQ = 0)
					</h3>
					<div style={{ display: "grid", gap: 6 }}>
						{zeroSales.slice(0, 20).map((item) => (
							<div
								key={item.ma_hang}
								style={{
									display: "flex",
									alignItems: "center",
									gap: 12,
									background: "#fff",
									borderRadius: 6,
									padding: "8px 14px",
									border: "1px solid #e8eaed",
									borderLeft: "2px solid #8b8fa3",
								}}
							>
								<span
									style={{
										fontSize: 12,
										fontWeight: 600,
										minWidth: 110,
										color: "#1e2330",
									}}
								>
									{item.ma_hang}
								</span>
								<span style={{ fontSize: 12, color: "#5a5f6e", flex: 1 }}>
									{item.ten_san_pham}
								</span>
								<span style={{ fontSize: 11, color: "#7a7f8e" }}>
									Tồn: {item.so_ton.toLocaleString("vi-VN")}
								</span>
								<span
									style={{
										fontSize: 11,
										padding: "2px 6px",
										borderRadius: 3,
										background: "#f0f1f4",
										color: "#5a5f6e",
									}}
								>
									{item.latest_outbound_month
										? `Xuất cuối: ${item.latest_outbound_month}`
										: "Chưa có dữ liệu xuất hàng"}
								</span>
							</div>
						))}
					</div>
				</div>
			)}

			{/* Restock Alerts Section */}
			{restockAlerts.length > 0 && (
				<div style={{ marginTop: 20 }}>
					<h3
						style={{
							margin: "0 0 12px",
							fontSize: 13,
							fontWeight: 600,
							color: "#3a3f4b",
						}}
					>
						Cần nhập lại hàng
					</h3>
					<div style={{ display: "grid", gap: 6 }}>
						{restockAlerts.slice(0, 20).map((item) => (
							<div
								key={item.ma_hang}
								style={{
									display: "flex",
									alignItems: "center",
									gap: 12,
									background: "#fff",
									borderRadius: 6,
									padding: "8px 14px",
									border: "1px solid #e8eaed",
									borderLeft: "2px solid #e5a04b",
								}}
							>
								<span
									style={{
										fontSize: 12,
										fontWeight: 600,
										minWidth: 110,
										color: "#1e2330",
									}}
								>
									{item.ma_hang}
								</span>
								<span style={{ fontSize: 12, color: "#5a5f6e", flex: 1 }}>
									{item.ten_san_pham}
								</span>
								<span
									style={{
										fontSize: 11,
										padding: "2px 6px",
										borderRadius: 3,
										background: "#fef3e2",
										color: "#b07a2a",
									}}
								>
									{item.last_outbound_date
										? `Hết hàng ${item.days_since_last_out} ngày`
										: "Chưa có dữ liệu xuất hàng"}
								</span>
							</div>
						))}
					</div>
				</div>
			)}

			{/* Existing Alerts */}
			{alerts.length > 0 && (
				<div style={{ marginTop: 20 }}>
					<h3
						style={{
							margin: "0 0 12px",
							fontSize: 13,
							fontWeight: 600,
							color: "#3a3f4b",
						}}
					>
						Cảnh báo
					</h3>
					<div style={{ display: "grid", gap: 6 }}>
						{alerts.slice(0, 20).map((alert, i) => (
							<div
								key={`${alert.ma_hang}-${alert.alert_type}-${i}`}
								style={{
									display: "flex",
									alignItems: "center",
									gap: 12,
									background: "#fff",
									borderRadius: 6,
									padding: "8px 14px",
									border: "1px solid #e8eaed",
									borderLeft: `2px solid ${alert.alert_type === "ton_lau" ? "#e5a04b" : "#e06363"}`,
								}}
							>
								<span
									style={{
										fontSize: 12,
										fontWeight: 600,
										minWidth: 110,
										color: "#1e2330",
									}}
								>
									{alert.ma_hang}
								</span>
								<span style={{ fontSize: 12, color: "#5a5f6e", flex: 1 }}>
									{alert.ten_san_pham}
								</span>
								<span
									style={{
										fontSize: 11,
										padding: "2px 6px",
										borderRadius: 3,
										background:
											alert.alert_type === "ton_lau" ? "#fef3e2" : "#fde8e8",
										color:
											alert.alert_type === "ton_lau" ? "#b07a2a" : "#b83b3b",
									}}
								>
									{alert.alert_type === "ton_lau" ? "Tồn lâu" : "Thiếu hàng"}
								</span>
								<span style={{ fontSize: 11, color: "#7a7f8e" }}>
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
