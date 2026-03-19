import { useCallback, useEffect, useState } from "react";
import { api } from "../api/client";
import type { OrderListItem } from "../types/dashboard";

export function Orders() {
	const [orders, setOrders] = useState<OrderListItem[]>([]);
	const [total, setTotal] = useState(0);
	const [page, setPage] = useState(1);
	const [typeFilter, setTypeFilter] = useState<string>("");
	const [loading, setLoading] = useState(false);
	const [showForm, setShowForm] = useState(false);
	const limit = 50;

	const fetchOrders = useCallback(async () => {
		setLoading(true);
		try {
			const resp = (await api.listOrders(
				typeFilter || undefined,
				page,
				limit,
			)) as {
				data: OrderListItem[];
				total: number;
			};
			setOrders(resp.data || []);
			setTotal(resp.total || 0);
		} catch (err) {
			console.error("Fetch orders error:", err);
		} finally {
			setLoading(false);
		}
	}, [typeFilter, page]);

	useEffect(() => {
		fetchOrders();
	}, [fetchOrders]);

	const totalPages = Math.ceil(total / limit);

	return (
		<div>
			<div
				style={{
					display: "flex",
					alignItems: "center",
					gap: 12,
					marginBottom: 16,
				}}
			>
				<h2
					style={{ margin: 0, fontSize: 16, fontWeight: 600, color: "#1e2330" }}
				>
					Nhập / Xuất
				</h2>
				<select
					value={typeFilter}
					onChange={(e) => {
						setTypeFilter(e.target.value);
						setPage(1);
					}}
					style={{
						padding: "6px 10px",
						borderRadius: 5,
						border: "1px solid #d5d8de",
						fontSize: 12,
						color: "#3a3f4b",
					}}
				>
					<option value="">Tất cả</option>
					<option value="in">Nhập</option>
					<option value="out">Xuất</option>
				</select>
				<button
					onClick={() => setShowForm(!showForm)}
					style={{
						marginLeft: "auto",
						padding: "7px 14px",
						background: "#3a3f4b",
						color: "#fff",
						border: "none",
						borderRadius: 5,
						cursor: "pointer",
						fontSize: 12,
						fontWeight: 500,
					}}
				>
					+ Tạo đơn
				</button>
			</div>

			{showForm && (
				<CreateOrderForm
					onCreated={() => {
						setShowForm(false);
						fetchOrders();
					}}
				/>
			)}

			{loading ? (
				<p style={{ color: "#888" }}>Đang tải...</p>
			) : (
				<>
					<table
						style={{
							width: "100%",
							borderCollapse: "collapse",
							background: "#fff",
							borderRadius: 6,
							overflow: "hidden",
							border: "1px solid #e8eaed",
							fontSize: 12,
						}}
					>
						<thead>
							<tr
								style={{
									borderBottom: "1px solid #e8eaed",
									textAlign: "left",
									background: "#fafbfc",
								}}
							>
								<th style={{ padding: "10px 12px" }}>ID</th>
								<th style={{ padding: "10px 12px" }}>Loại</th>
								<th style={{ padding: "10px 12px" }}>Mã hàng</th>
								<th style={{ padding: "10px 12px" }}>Tên sản phẩm</th>
								<th style={{ padding: "10px 12px" }}>Mã thùng</th>
								<th style={{ padding: "10px 12px", textAlign: "right" }}>
									Số lượng
								</th>
								<th style={{ padding: "10px 12px" }}>Ngày</th>
							</tr>
						</thead>
						<tbody>
							{orders.map((o) => (
								<tr
									key={`${o.type}-${o.id}`}
									style={{ borderBottom: "1px solid #f0f0f0" }}
								>
									<td style={{ padding: "8px 12px" }}>{o.id}</td>
									<td style={{ padding: "8px 12px" }}>
										<span
											style={{
												padding: "2px 6px",
												borderRadius: 3,
												fontSize: 11,
												fontWeight: 500,
												background: o.type === "in" ? "#e8f5ed" : "#fde8e8",
												color: o.type === "in" ? "#3a7d4f" : "#b83b3b",
											}}
										>
											{o.type === "in" ? "Nhập" : "Xuất"}
										</span>
									</td>
									<td style={{ padding: "8px 12px" }}>{o.ma_hang}</td>
									<td style={{ padding: "8px 12px" }}>{o.ten_san_pham}</td>
									<td
										style={{
											padding: "8px 12px",
											fontWeight: o.type === "out" ? 700 : 400,
										}}
									>
										{o.batch_code || "—"}
									</td>
									<td style={{ padding: "8px 12px", textAlign: "right" }}>
										{o.so_luong.toLocaleString("vi-VN")}
									</td>
									<td style={{ padding: "8px 12px" }}>
										{new Date(o.ngay_nhan_hang).toLocaleDateString("vi-VN")}
									</td>
								</tr>
							))}
							{orders.length === 0 && (
								<tr>
									<td
										colSpan={7}
										style={{ padding: 20, textAlign: "center", color: "#888" }}
									>
										Chưa có đơn nào.
									</td>
								</tr>
							)}
						</tbody>
					</table>

					{totalPages > 1 && (
						<div
							style={{
								display: "flex",
								justifyContent: "center",
								gap: 8,
								marginTop: 16,
							}}
						>
							<button
								disabled={page <= 1}
								onClick={() => setPage(page - 1)}
								style={{
									padding: "6px 12px",
									borderRadius: 4,
									border: "1px solid #ccc",
									cursor: page <= 1 ? "not-allowed" : "pointer",
									background: "#fff",
								}}
							>
								← Trước
							</button>
							<span style={{ padding: "6px 12px", fontSize: 13 }}>
								{page} / {totalPages}
							</span>
							<button
								disabled={page >= totalPages}
								onClick={() => setPage(page + 1)}
								style={{
									padding: "6px 12px",
									borderRadius: 4,
									border: "1px solid #ccc",
									cursor: page >= totalPages ? "not-allowed" : "pointer",
									background: "#fff",
								}}
							>
								Sau →
							</button>
						</div>
					)}
				</>
			)}
		</div>
	);
}

// --- Create Order Form ---

function CreateOrderForm({ onCreated }: { onCreated: () => void }) {
	const [type, setType] = useState<"in" | "out">("in");
	const [maHang, setMaHang] = useState("");
	const [batchCode, setBatchCode] = useState("");
	const [soLuong, setSoLuong] = useState("");
	const [submitting, setSubmitting] = useState(false);
	const [result, setResult] = useState<string | null>(null);

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		if (!maHang || !soLuong) return;

		setSubmitting(true);
		setResult(null);
		try {
			const body: Record<string, unknown> = {
				type,
				ma_hang: maHang,
				so_luong: parseFloat(soLuong),
			};
			if (type === "in") {
				body.batch_code = batchCode || maHang;
			}

			const resp = (await api.createOrder(body)) as {
				allocations?: { batch_code: string; qty: number }[];
			};

			if (type === "out" && resp.allocations) {
				const detail = resp.allocations
					.map((a) => `${a.batch_code}: ${a.qty}`)
					.join(", ");
				setResult(`Xuất thành công — FIFO: ${detail}`);
			} else {
				setResult("Tạo đơn thành công!");
			}

			setMaHang("");
			setBatchCode("");
			setSoLuong("");
			setTimeout(onCreated, 1500);
		} catch (err) {
			setResult(`Lỗi: ${err instanceof Error ? err.message : "Unknown"}`);
		} finally {
			setSubmitting(false);
		}
	};

	return (
		<form
			onSubmit={handleSubmit}
			style={{
				background: "#fff",
				borderRadius: 6,
				padding: 18,
				marginBottom: 16,
				border: "1px solid #e8eaed",
				display: "flex",
				gap: 10,
				alignItems: "flex-end",
				flexWrap: "wrap",
			}}
		>
			<div>
				<label
					style={{
						fontSize: 11,
						color: "#7a7f8e",
						display: "block",
						marginBottom: 3,
					}}
				>
					Loại
				</label>
				<select
					value={type}
					onChange={(e) => setType(e.target.value as "in" | "out")}
					style={{
						padding: "7px 10px",
						borderRadius: 5,
						border: "1px solid #d5d8de",
						fontSize: 12,
						color: "#3a3f4b",
					}}
				>
					<option value="in">Nhập</option>
					<option value="out">Xuất</option>
				</select>
			</div>

			<div>
				<label
					style={{
						fontSize: 11,
						color: "#7a7f8e",
						display: "block",
						marginBottom: 3,
					}}
				>
					Mã hàng
				</label>
				<input
					value={maHang}
					onChange={(e) => setMaHang(e.target.value)}
					placeholder="VD: SP001"
					required
					style={{
						padding: "7px 10px",
						borderRadius: 5,
						border: "1px solid #d5d8de",
						fontSize: 12,
						width: 140,
						color: "#3a3f4b",
					}}
				/>
			</div>

			{type === "in" && (
				<div>
					<label
						style={{
							fontSize: 11,
							color: "#7a7f8e",
							display: "block",
							marginBottom: 3,
						}}
					>
						Mã thùng
					</label>
					<input
						value={batchCode}
						onChange={(e) => setBatchCode(e.target.value)}
						placeholder="VD: BATCH-001"
						style={{
							padding: "7px 10px",
							borderRadius: 5,
							border: "1px solid #d5d8de",
							fontSize: 12,
							width: 140,
							color: "#3a3f4b",
						}}
					/>
				</div>
			)}

			<div>
				<label
					style={{
						fontSize: 11,
						color: "#7a7f8e",
						display: "block",
						marginBottom: 3,
					}}
				>
					Số lượng
				</label>
				<input
					type="number"
					value={soLuong}
					onChange={(e) => setSoLuong(e.target.value)}
					placeholder="0"
					required
					min="0.01"
					step="0.01"
					style={{
						padding: "7px 10px",
						borderRadius: 5,
						border: "1px solid #d5d8de",
						fontSize: 12,
						width: 100,
						color: "#3a3f4b",
					}}
				/>
			</div>

			<button
				type="submit"
				disabled={submitting}
				style={{
					padding: "7px 16px",
					background: type === "in" ? "#5bb98c" : "#e06363",
					color: "#fff",
					border: "none",
					borderRadius: 5,
					cursor: submitting ? "not-allowed" : "pointer",
					fontSize: 12,
					fontWeight: 500,
				}}
			>
				{submitting
					? "Đang xử lý..."
					: type === "in"
						? "Nhập hàng"
						: "Xuất hàng"}
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
		</form>
	);
}
