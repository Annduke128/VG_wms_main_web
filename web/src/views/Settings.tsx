import { useCallback, useEffect, useRef, useState } from "react";
import { api } from "../api/client";
import type { ThresholdEntry } from "../types/dashboard";
import type { AsyncJob, ImportBatch } from "../types/grid";

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

	// Import state
	const [uploading, setUploading] = useState(false);
	const [importJob, setImportJob] = useState<AsyncJob | null>(null);
	const [importBatch, setImportBatch] = useState<ImportBatch | null>(null);
	const [showErrors, setShowErrors] = useState(false);
	const fileRef = useRef<HTMLInputElement | null>(null);

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
			setResult("Lưu thành công");
			setFormMaHang("");
			setMinQty("");
			setOptimalQty("");
			setMaxAgeDays("");
			setEffectiveTo("");
			if (maHang === formMaHang) fetchHistory(maHang);
		} catch (err) {
			setResult(`Lỗi: ${err instanceof Error ? err.message : "Unknown"}`);
		} finally {
			setSubmitting(false);
		}
	};

	// --- Import handlers ---

	const handleDownloadTemplate = useCallback(async () => {
		try {
			const blob = await api.downloadInventoryTemplate();
			const url = URL.createObjectURL(blob);
			const a = document.createElement("a");
			a.href = url;
			a.download = "MauTonKho.xlsx";
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);
		} catch (err) {
			alert(
				`Tải mẫu thất bại: ${err instanceof Error ? err.message : "Unknown"}`,
			);
		}
	}, []);

	const handleUpload = useCallback(async () => {
		const input = fileRef.current;
		if (!input?.files?.[0]) {
			alert("Chọn file trước khi upload");
			return;
		}

		setUploading(true);
		setImportJob(null);
		setImportBatch(null);
		setShowErrors(false);
		try {
			const res = (await api.importFile("inventory", input.files[0])) as {
				job_id: string;
			};
			pollJob(res.job_id);
		} catch (err) {
			alert(
				`Upload thất bại: ${err instanceof Error ? err.message : "Unknown"}`,
			);
		} finally {
			setUploading(false);
		}
	}, []);

	const pollJob = useCallback((jobId: string) => {
		const poll = async () => {
			try {
				const job = (await api.getJob(jobId)) as AsyncJob;
				setImportJob(job);
				if (job.status === "pending" || job.status === "running") {
					setTimeout(poll, 1000);
				} else {
					// Job finished → fetch batch info for errors
					try {
						const batch = (await api.getImportBatch()) as ImportBatch;
						setImportBatch(batch);
					} catch {
						// ignore if batch fetch fails
					}
				}
			} catch {
				console.error("Poll error");
			}
		};
		poll();
	}, []);

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
				Cài đặt
			</h2>

			{/* --- Import Inventory Section --- */}
			<div
				style={{
					background: "#fff",
					borderRadius: 8,
					padding: 20,
					border: "1px solid #e8eaed",
					marginBottom: 20,
				}}
			>
				<h3
					style={{
						margin: "0 0 12px",
						fontSize: 13,
						fontWeight: 600,
						color: "#3a3f4b",
					}}
				>
					Import kho (gộp sản phẩm + tồn kho + lô)
				</h3>
				<p style={{ fontSize: 11, color: "#7a7f8e", marginBottom: 12 }}>
					Upload file .xlsx (17 cột): Mã vạch, Tên sản phẩm, BU, Mã cat, Mã nhóm
					hàng, Nhóm hàng, ĐVT, Quy cách, Đơn giá, VAT, Ngày cập nhật, Hoa hồng,
					Mã lô hàng, Ngày nhập, Số tồn, Số nhập, Số xuất. Giá NIV, Đơn giá
					nhập, Tiền tồn/nhập/xuất được tự động tính. Mỗi mã vạch có nhiều lô →
					nhập nhiều dòng. Ngày dùng dd/mm/yyyy, dd-mm-yyyy, hoặc dd-mm-yy. Mã
					lô hàng nếu để trống sẽ tự tạo LOT-&lt;mã vạch&gt;-&lt;yyyymmdd&gt;.
				</p>
				<div style={{ display: "flex", alignItems: "center", gap: 10 }}>
					<button
						onClick={handleDownloadTemplate}
						style={{
							padding: "7px 16px",
							background: "#fff",
							color: "#3a3f4b",
							border: "1px solid #d5d8de",
							borderRadius: 5,
							cursor: "pointer",
							fontSize: 12,
							fontWeight: 500,
						}}
					>
						Tải mẫu .xlsx
					</button>
					<input
						ref={(el) => {
							fileRef.current = el;
						}}
						type="file"
						accept=".xlsx"
						style={{ fontSize: 12 }}
					/>
					<button
						onClick={handleUpload}
						disabled={uploading}
						style={{
							padding: "7px 16px",
							background: "#3a3f4b",
							color: "#fff",
							border: "none",
							borderRadius: 5,
							cursor: uploading ? "not-allowed" : "pointer",
							fontSize: 12,
							fontWeight: 500,
						}}
					>
						{uploading ? "Đang upload..." : "Upload"}
					</button>
				</div>

				{importJob && (
					<div
						style={{
							marginTop: 12,
							padding: "10px 14px",
							background: "#fafbfc",
							borderRadius: 5,
							border: "1px solid #e8eaed",
							fontSize: 12,
							position: "relative",
						}}
					>
						<div
							style={{
								display: "flex",
								alignItems: "center",
								justifyContent: "space-between",
							}}
						>
							<div>
								<span style={{ color: "#7a7f8e" }}>Trạng thái: </span>
								<span
									style={{
										fontWeight: 600,
										color:
											importJob.status === "completed"
												? "#3a7d4f"
												: importJob.status === "failed"
													? "#b83b3b"
													: "#7a7f8e",
									}}
								>
									{importJob.status === "completed"
										? "Hoàn tất"
										: importJob.status === "failed"
											? "Thất bại"
											: importJob.status === "running"
												? "Đang xử lý..."
												: "Đang chờ..."}
								</span>
								{importJob.error && (
									<span style={{ color: "#b83b3b", marginLeft: 8 }}>
										{importJob.error}
									</span>
								)}
							</div>

							{/* Error button */}
							{importBatch &&
								importBatch.error_rows > 0 &&
								(() => {
									let errors: string[] = [];
									try {
										errors = JSON.parse(importBatch.errors) as string[];
									} catch {
										/* empty */
									}
									if (errors.length === 0) return null;
									const maxShow = 10;
									const shown = errors.slice(0, maxShow);
									const remaining = errors.length - maxShow;
									return (
										<div style={{ position: "relative" }}>
											<button
												onClick={() => setShowErrors((v) => !v)}
												onMouseEnter={() => setShowErrors(true)}
												onMouseLeave={() => {
													// Don't hide if pinned (clicked)
												}}
												style={{
													padding: "3px 10px",
													background: "#fef3f2",
													color: "#b83b3b",
													border: "1px solid #fecaca",
													borderRadius: 4,
													cursor: "pointer",
													fontSize: 11,
													fontWeight: 500,
												}}
											>
												Xem lỗi ({importBatch.error_rows})
											</button>
											{showErrors && (
												<div
													onMouseEnter={() => setShowErrors(true)}
													onMouseLeave={() => setShowErrors(false)}
													style={{
														position: "absolute",
														top: "calc(100% + 6px)",
														right: 0,
														width: 380,
														maxHeight: 260,
														overflowY: "auto",
														background: "#fff",
														border: "1px solid #e8eaed",
														borderRadius: 6,
														boxShadow: "0 4px 16px rgba(0,0,0,0.12)",
														padding: "10px 12px",
														zIndex: 100,
														fontSize: 11,
														color: "#3a3f4b",
													}}
												>
													{shown.map((err, i) => (
														<div
															key={i}
															style={{
																padding: "3px 0",
																borderBottom:
																	i < shown.length - 1
																		? "1px solid #f2f3f5"
																		: "none",
																wordBreak: "break-word",
															}}
														>
															{err}
														</div>
													))}
													{remaining > 0 && (
														<div
															style={{
																padding: "6px 0 0",
																color: "#7a7f8e",
																fontStyle: "italic",
															}}
														>
															+{remaining} dòng khác
														</div>
													)}
												</div>
											)}
										</div>
									);
								})()}
						</div>

						{/* Warning banner: 0 success rows */}
						{importBatch &&
							importBatch.success_rows === 0 &&
							importBatch.error_rows > 0 && (
								<div
									style={{
										marginTop: 8,
										padding: "8px 12px",
										background: "#fef9c3",
										border: "1px solid #fde68a",
										borderRadius: 4,
										fontSize: 11,
										color: "#92400e",
										fontWeight: 500,
									}}
								>
									Không có dòng nào import thành công. Kiểm tra định dạng file
									và dữ liệu.
								</div>
							)}

						{/* Summary row */}
						{importBatch &&
							(importJob.status === "completed" ||
								importJob.status === "failed") && (
								<div style={{ marginTop: 6, fontSize: 11, color: "#7a7f8e" }}>
									Tổng: {importBatch.total_rows} | Thành công:{" "}
									{importBatch.success_rows} | Lỗi: {importBatch.error_rows}
								</div>
							)}
					</div>
				)}
			</div>

			{/* --- Threshold Form --- */}
			<form
				onSubmit={handleSubmit}
				style={{
					background: "#fff",
					borderRadius: 8,
					padding: 20,
					border: "1px solid #e8eaed",
					marginBottom: 20,
				}}
			>
				<h3
					style={{
						margin: "0 0 14px",
						fontSize: 13,
						fontWeight: 600,
						color: "#3a3f4b",
					}}
				>
					Nhập ngưỡng mới
				</h3>
				<div
					style={{
						display: "grid",
						gridTemplateColumns: "repeat(3, 1fr)",
						gap: 14,
					}}
				>
					<Field label="Mã hàng" required>
						<input
							value={formMaHang}
							onChange={(e) => setFormMaHang(e.target.value)}
							placeholder="VD: SP001"
							required
							style={inputStyle}
						/>
					</Field>
					<Field label="Min qty" required>
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
					<Field label="Optimal qty" required>
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
					<Field label="Max age (ngày)" required>
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
					<Field label="Hiệu lực đến">
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
						marginTop: 14,
						display: "flex",
						alignItems: "center",
						gap: 10,
					}}
				>
					<button
						type="submit"
						disabled={submitting}
						style={{
							padding: "7px 18px",
							background: "#3a3f4b",
							color: "#fff",
							border: "none",
							borderRadius: 5,
							cursor: submitting ? "not-allowed" : "pointer",
							fontSize: 12,
							fontWeight: 500,
						}}
					>
						{submitting ? "Đang lưu..." : "Lưu"}
					</button>
					{result && (
						<span
							style={{
								fontSize: 12,
								color: result.startsWith("Lỗi") ? "#b83b3b" : "#3a7d4f",
								fontWeight: 500,
							}}
						>
							{result}
						</span>
					)}
				</div>
			</form>

			{/* --- History Lookup --- */}
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
						margin: "0 0 10px",
						fontSize: 13,
						fontWeight: 600,
						color: "#3a3f4b",
					}}
				>
					Lịch sử ngưỡng
				</h3>
				<div style={{ marginBottom: 14 }}>
					<input
						value={maHang}
						onChange={(e) => setMaHang(e.target.value)}
						placeholder="Nhập mã hàng để tra cứu..."
						style={{ ...inputStyle, width: 260 }}
					/>
				</div>

				{historyLoading ? (
					<p style={{ color: "#7a7f8e", fontSize: 12 }}>Đang tải...</p>
				) : history.length === 0 ? (
					<p style={{ color: "#7a7f8e", fontSize: 12 }}>
						{maHang
							? "Chưa có ngưỡng cho mã này."
							: "Nhập mã hàng để xem lịch sử."}
					</p>
				) : (
					<table
						style={{ width: "100%", borderCollapse: "collapse", fontSize: 12 }}
					>
						<thead>
							<tr
								style={{ borderBottom: "1px solid #e8eaed", textAlign: "left" }}
							>
								<th
									style={{
										padding: "6px 10px",
										color: "#7a7f8e",
										fontWeight: 500,
									}}
								>
									Min qty
								</th>
								<th
									style={{
										padding: "6px 10px",
										color: "#7a7f8e",
										fontWeight: 500,
									}}
								>
									Optimal qty
								</th>
								<th
									style={{
										padding: "6px 10px",
										color: "#7a7f8e",
										fontWeight: 500,
									}}
								>
									Max age
								</th>
								<th
									style={{
										padding: "6px 10px",
										color: "#7a7f8e",
										fontWeight: 500,
									}}
								>
									Source
								</th>
								<th
									style={{
										padding: "6px 10px",
										color: "#7a7f8e",
										fontWeight: 500,
									}}
								>
									Từ
								</th>
								<th
									style={{
										padding: "6px 10px",
										color: "#7a7f8e",
										fontWeight: 500,
									}}
								>
									Đến
								</th>
								<th
									style={{
										padding: "6px 10px",
										color: "#7a7f8e",
										fontWeight: 500,
									}}
								>
									Ngày tạo
								</th>
							</tr>
						</thead>
						<tbody>
							{history.map((t) => (
								<tr key={t.id} style={{ borderBottom: "1px solid #f2f3f5" }}>
									<td style={{ padding: "6px 10px" }}>{t.min_qty}</td>
									<td style={{ padding: "6px 10px" }}>{t.optimal_qty}</td>
									<td style={{ padding: "6px 10px" }}>{t.max_age_days}</td>
									<td style={{ padding: "6px 10px" }}>{t.source}</td>
									<td style={{ padding: "6px 10px" }}>
										{new Date(t.effective_from).toLocaleDateString("vi-VN")}
									</td>
									<td style={{ padding: "6px 10px" }}>
										{t.effective_to
											? new Date(t.effective_to).toLocaleDateString("vi-VN")
											: "—"}
									</td>
									<td style={{ padding: "6px 10px" }}>
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
					fontSize: 11,
					color: "#7a7f8e",
					display: "block",
					marginBottom: 3,
				}}
			>
				{label} {required && <span style={{ color: "#e06363" }}>*</span>}
			</label>
			{children}
		</div>
	);
}

const inputStyle: React.CSSProperties = {
	width: "100%",
	padding: "7px 10px",
	borderRadius: 5,
	border: "1px solid #d5d8de",
	fontSize: 12,
	color: "#3a3f4b",
};
