import { useCallback, useEffect, useState } from "react";
import { api } from "../api/client";
import type { ThresholdEntry } from "../types/dashboard";

export function Settings() {
	const [maHang, setMaHang] = useState("");
	const [history, setHistory] = useState<ThresholdEntry[]>([]);
	const [historyLoading, setHistoryLoading] = useState(false);

	// Form state
	const [formMaHang, setFormMaHang] = useState("");
	const [minQty, setMinQty] = useState("");
	const [optimalQty, setOptimalQty] = useState("");
	const [maxAgeDays, setMaxAgeDays] = useState("");
	const [effectiveFrom, setEffectiveFrom] = useState(
		new Date().toISOString().slice(0, 10),
	);
	const [effectiveTo, setEffectiveTo] = useState("");
	const [submitting, setSubmitting] = useState(false);
	const [result, setResult] = useState<string | null>(null);

	const fetchHistory = useCallback(async (sku: string) => {
		if (!sku) {
			setHistory([]);
			return;
		}
		setHistoryLoading(true);
		try {
			const data = (await api.getThresholds(sku)) as ThresholdEntry[];
			setHistory(data || []);
		} catch (err) {
			console.error("Fetch thresholds error:", err);
			setHistory([]);
		} finally {
			setHistoryLoading(false);
		}
	}, []);

	useEffect(() => {
		if (maHang.length >= 2) {
			const timer = setTimeout(() => fetchHistory(maHang), 400);
			return () => clearTimeout(timer);
		}
		setHistory([]);
	}, [maHang, fetchHistory]);

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		if (!formMaHang || !minQty || !optimalQty || !maxAgeDays) return;

		setSubmitting(true);
		setResult(null);
		try {
			await api.saveThreshold({
				ma_hang: formMaHang,
				min_qty: parseFloat(minQty),
				optimal_qty: parseFloat(optimalQty),
				max_age_days: parseFloat(maxAgeDays),
				source: "manual",
				effective_from: effectiveFrom
					? new Date(effectiveFrom).toISOString()
					: new Date().toISOString(),
				effective_to: effectiveTo ? new Date(effectiveTo).toISOString() : null,
			});
			setResult("Lưu threshold thành công!");
			setFormMaHang("");
			setMinQty("");
			setOptimalQty("");
			setMaxAgeDays("");
			setEffectiveTo("");
			// Refresh history if viewing same SKU
			if (maHang === formMaHang) fetchHistory(maHang);
		} catch (err) {
			setResult(`Lỗi: ${err instanceof Error ? err.message : "Unknown"}`);
		} finally {
			setSubmitting(false);
		}
	};

	return (
		<div>
			<h2 style={{ margin: "0 0 20px", fontSize: 20 }}>Cài đặt ngưỡng</h2>

			{/* Form */}
			<form
				onSubmit={handleSubmit}
				style={{
					background: "#fff",
					borderRadius: 10,
					padding: 24,
					boxShadow: "0 1px 4px rgba(0,0,0,0.08)",
					marginBottom: 24,
				}}
			>
				<h3 style={{ margin: "0 0 16px", fontSize: 15 }}>Nhập threshold mới</h3>
				<div
					style={{
						display: "grid",
						gridTemplateColumns: "repeat(3, 1fr)",
						gap: 16,
					}}
				>
					<Field label="Mã hàng (barcode)" required>
						<input
							value={formMaHang}
							onChange={(e) => setFormMaHang(e.target.value)}
							placeholder="VD: SP001"
							required
							style={inputStyle}
						/>
					</Field>
					<Field label="Min qty (thiếu hàng khi < giá trị này)" required>
						<input
							type="number"
							value={minQty}
							onChange={(e) => setMinQty(e.target.value)}
							placeholder="0"
							required
							min="0"
							step="0.01"
							style={inputStyle}
						/>
					</Field>
					<Field label="Optimal qty (tối ưu)" required>
						<input
							type="number"
							value={optimalQty}
							onChange={(e) => setOptimalQty(e.target.value)}
							placeholder="0"
							required
							min="0"
							step="0.01"
							style={inputStyle}
						/>
					</Field>
					<Field label="Max age days (tồn lâu khi >= ngày này)" required>
						<input
							type="number"
							value={maxAgeDays}
							onChange={(e) => setMaxAgeDays(e.target.value)}
							placeholder="30"
							required
							min="1"
							step="1"
							style={inputStyle}
						/>
					</Field>
					<Field label="Hiệu lực từ">
						<input
							type="date"
							value={effectiveFrom}
							onChange={(e) => setEffectiveFrom(e.target.value)}
							style={inputStyle}
						/>
					</Field>
					<Field label="Hiệu lực đến (trống = vô thời hạn)">
						<input
							type="date"
							value={effectiveTo}
							onChange={(e) => setEffectiveTo(e.target.value)}
							style={inputStyle}
						/>
					</Field>
				</div>

				<div
					style={{
						marginTop: 16,
						display: "flex",
						alignItems: "center",
						gap: 12,
					}}
				>
					<button
						type="submit"
						disabled={submitting}
						style={{
							padding: "10px 24px",
							background: "#1976d2",
							color: "#fff",
							border: "none",
							borderRadius: 6,
							cursor: submitting ? "not-allowed" : "pointer",
							fontSize: 14,
							fontWeight: 600,
						}}
					>
						{submitting ? "Đang lưu..." : "Lưu threshold"}
					</button>
					{result && (
						<span
							style={{
								fontSize: 13,
								color: result.startsWith("Lỗi") ? "#C62828" : "#2E7D32",
								fontWeight: 500,
							}}
						>
							{result}
						</span>
					)}
				</div>
			</form>

			{/* History lookup */}
			<div
				style={{
					background: "#fff",
					borderRadius: 10,
					padding: 24,
					boxShadow: "0 1px 4px rgba(0,0,0,0.08)",
				}}
			>
				<h3 style={{ margin: "0 0 12px", fontSize: 15 }}>Lịch sử threshold</h3>
				<div style={{ marginBottom: 16 }}>
					<input
						value={maHang}
						onChange={(e) => setMaHang(e.target.value)}
						placeholder="Nhập mã hàng để tra cứu..."
						style={{ ...inputStyle, width: 300 }}
					/>
				</div>

				{historyLoading ? (
					<p style={{ color: "#888" }}>Đang tải...</p>
				) : history.length === 0 ? (
					<p style={{ color: "#888" }}>
						{maHang
							? "Chưa có threshold cho SKU này."
							: "Nhập mã hàng để xem lịch sử."}
					</p>
				) : (
					<table
						style={{ width: "100%", borderCollapse: "collapse", fontSize: 13 }}
					>
						<thead>
							<tr style={{ borderBottom: "2px solid #eee", textAlign: "left" }}>
								<th style={{ padding: "8px 12px" }}>Min qty</th>
								<th style={{ padding: "8px 12px" }}>Optimal qty</th>
								<th style={{ padding: "8px 12px" }}>Max age (days)</th>
								<th style={{ padding: "8px 12px" }}>Source</th>
								<th style={{ padding: "8px 12px" }}>Hiệu lực từ</th>
								<th style={{ padding: "8px 12px" }}>Hiệu lực đến</th>
								<th style={{ padding: "8px 12px" }}>Ngày tạo</th>
							</tr>
						</thead>
						<tbody>
							{history.map((t) => (
								<tr key={t.id} style={{ borderBottom: "1px solid #f0f0f0" }}>
									<td style={{ padding: "8px 12px" }}>{t.min_qty}</td>
									<td style={{ padding: "8px 12px" }}>{t.optimal_qty}</td>
									<td style={{ padding: "8px 12px" }}>{t.max_age_days}</td>
									<td style={{ padding: "8px 12px" }}>{t.source}</td>
									<td style={{ padding: "8px 12px" }}>
										{new Date(t.effective_from).toLocaleDateString("vi-VN")}
									</td>
									<td style={{ padding: "8px 12px" }}>
										{t.effective_to
											? new Date(t.effective_to).toLocaleDateString("vi-VN")
											: "—"}
									</td>
									<td style={{ padding: "8px 12px" }}>
										{new Date(t.created_at).toLocaleDateString("vi-VN")}
									</td>
								</tr>
							))}
						</tbody>
					</table>
				)}
			</div>
		</div>
	);
}

// --- Helpers ---

function Field({
	label,
	required,
	children,
}: {
	label: string;
	required?: boolean;
	children: React.ReactNode;
}) {
	return (
		<div>
			<label
				style={{
					fontSize: 12,
					color: "#888",
					display: "block",
					marginBottom: 4,
				}}
			>
				{label} {required && <span style={{ color: "#f44336" }}>*</span>}
			</label>
			{children}
		</div>
	);
}

const inputStyle: React.CSSProperties = {
	width: "100%",
	padding: "8px 12px",
	borderRadius: 6,
	border: "1px solid #ccc",
	fontSize: 13,
};
