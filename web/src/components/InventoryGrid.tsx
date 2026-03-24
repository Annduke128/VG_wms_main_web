import DataEditor, {
	CompactSelection,
	type GridCell,
	GridCellKind,
	type GridColumn,
	type GridSelection,
	type Item,
} from "@glideapps/glide-data-grid";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import "@glideapps/glide-data-grid/dist/index.css";
import { api } from "../api/client";
import type {
	FilterItem,
	GridRequest,
	GridResponse,
	SortItem,
} from "../types/grid";
import type { InventoryMain } from "../types/inventory";

const PAGE_SIZE = 200;

const DEFAULT_COLUMNS: GridColumn[] = [
	{ id: "ma_hang", title: "Mã hàng", width: 120 },
	{ id: "ten_san_pham", title: "Tên sản phẩm", width: 250 },
	{ id: "don_gia", title: "Đơn giá", width: 110 },
	{ id: "so_ton", title: "Số tồn", width: 100 },
	{ id: "so_nhap", title: "Số nhập", width: 100 },
	{ id: "so_xuat", title: "Số xuất", width: 100 },
	{ id: "tien_ton", title: "Tiền tồn", width: 120 },
	{ id: "tien_nhap", title: "Tiền nhập", width: 120 },
	{ id: "tien_xuat", title: "Tiền xuất", width: 120 },
	{ id: "so_ngay_ton", title: "Số ngày tồn", width: 110 },
	{ id: "luong_ban_binh_quan_ngay", title: "LBBQ/ngày", width: 120 },
	{ id: "so_ngay_ton_ban", title: "Ngày tồn bán", width: 110 },
];

// All exportable columns in default order (BU + Nhóm after Tên SP)
const ALL_EXPORT_COLUMNS: { id: string; title: string }[] = [
	{ id: "ma_hang", title: "Mã hàng" },
	{ id: "ten_san_pham", title: "Tên sản phẩm" },
	{ id: "ma_bu", title: "Mã BU" },
	{ id: "ma_nhom_hang", title: "Mã nhóm hàng" },
	{ id: "don_gia", title: "Đơn giá" },
	{ id: "so_ton", title: "Số tồn" },
	{ id: "so_nhap", title: "Số nhập" },
	{ id: "so_xuat", title: "Số xuất" },
	{ id: "tien_ton", title: "Tiền tồn" },
	{ id: "tien_nhap", title: "Tiền nhập" },
	{ id: "tien_xuat", title: "Tiền xuất" },
	{ id: "so_ngay_ton", title: "Số ngày tồn" },
	{ id: "luong_ban_binh_quan_ngay", title: "LBBQ/ngày" },
	{ id: "so_ngay_ton_ban", title: "Ngày tồn bán" },
];

// Columns that should not be editable in the grid
const READONLY_COLUMNS = new Set([
	"ma_hang",
	"don_gia",
	"tien_ton",
	"tien_nhap",
	"tien_xuat",
	"so_ngay_ton",
	"luong_ban_binh_quan_ngay",
	"so_ngay_ton_ban",
]);

interface InventoryGridProps {
	onRowSelect?: (maHang: string) => void;
}

export function InventoryGrid({ onRowSelect }: InventoryGridProps = {}) {
	const [data, setData] = useState<InventoryMain[]>([]);
	const [totalRows, setTotalRows] = useState(0);
	const [loading, setLoading] = useState(false);
	const [sortModel] = useState<SortItem[]>([]);

	// Filter state
	const [buFilter, setBuFilter] = useState("");
	const [nhomFilter, setNhomFilter] = useState("");
	const [filterOptions, setFilterOptions] = useState<{
		ma_bu: string[];
		ma_nhom_hang: string[];
	}>({ ma_bu: [], ma_nhom_hang: [] });

	// Compute filterModel from filter inputs
	const filterModel = useMemo(() => {
		const fm: Record<string, FilterItem> = {};
		if (buFilter)
			fm.ma_bu = { filterType: "text", type: "contains", filter: buFilter };
		if (nhomFilter)
			fm.ma_nhom_hang = {
				filterType: "text",
				type: "contains",
				filter: nhomFilter,
			};
		return fm;
	}, [buFilter, nhomFilter]);

	// Column state (for resize/reorder — session only)
	const [cols, setCols] = useState<GridColumn[]>(DEFAULT_COLUMNS);

	// Selection
	const [selection, setSelection] = useState<GridSelection>({
		columns: CompactSelection.empty(),
		rows: CompactSelection.empty(),
	});

	// Export modal
	const [showExportModal, setShowExportModal] = useState(false);
	const [exportColumns, setExportColumns] = useState<Set<string>>(
		() => new Set(ALL_EXPORT_COLUMNS.map((c) => c.id)),
	);
	const [exporting, setExporting] = useState(false);

	// Tooltip
	const [tooltip, setTooltip] = useState<{
		text: string;
		x: number;
		y: number;
	} | null>(null);
	const containerRef = useRef<HTMLDivElement>(null);
	const mousePos = useRef({ x: 0, y: 0 });

	const loadedRanges = useRef<Set<string>>(new Set());

	// Load filter options once
	useEffect(() => {
		api
			.inventoryFilterOptions()
			.then((opts) => setFilterOptions(opts))
			.catch(console.error);
	}, []);

	// Fetch data range
	const fetchRange = useCallback(
		async (startRow: number, endRow: number) => {
			const key = `${startRow}-${endRow}`;
			if (loadedRanges.current.has(key)) return;

			setLoading(true);
			try {
				const req: GridRequest = { startRow, endRow, sortModel, filterModel };
				const resp = (await api.inventoryGrid(
					req,
				)) as GridResponse<InventoryMain>;

				setData((prev) => {
					const next = [...prev];
					resp.rowsData.forEach((row, i) => {
						next[startRow + i] = row;
					});
					return next;
				});
				setTotalRows(resp.totalRowCount);
				loadedRanges.current.add(key);
			} catch (err) {
				console.error("Grid fetch error:", err);
			} finally {
				setLoading(false);
			}
		},
		[sortModel, filterModel],
	);

	// Reset and reload when filter changes
	useEffect(() => {
		loadedRanges.current.clear();
		setData([]);
		setSelection({
			columns: CompactSelection.empty(),
			rows: CompactSelection.empty(),
		});
		fetchRange(0, PAGE_SIZE);
	}, [fetchRange]);

	// Visible region changed → fetch more data
	const onVisibleRegionChanged = useCallback(
		(range: { x: number; y: number; width: number; height: number }) => {
			const start = Math.floor(range.y / PAGE_SIZE) * PAGE_SIZE;
			const end = start + PAGE_SIZE;
			fetchRange(start, end);
			if (end < totalRows) {
				fetchRange(end, end + PAGE_SIZE);
			}
		},
		[fetchRange, totalRows],
	);

	// Get cell content — uses cols[col].id for field lookup
	const getCellContent = useCallback(
		([col, row]: Item): GridCell => {
			const item = data[row];
			if (!item) {
				return { kind: GridCellKind.Loading, allowOverlay: false };
			}

			const colDef = cols[col];
			const value = item[colDef.id as keyof InventoryMain];

			if (typeof value === "number") {
				return {
					kind: GridCellKind.Number,
					data: value,
					displayData: value.toLocaleString("vi-VN"),
					allowOverlay: !READONLY_COLUMNS.has(colDef.id as string),
					readonly: READONLY_COLUMNS.has(colDef.id as string),
				};
			}

			return {
				kind: GridCellKind.Text,
				data: String(value ?? ""),
				displayData: String(value ?? ""),
				allowOverlay: !READONLY_COLUMNS.has(colDef.id as string),
				readonly: READONLY_COLUMNS.has(colDef.id as string),
			};
		},
		[data, cols],
	);

	// Cell edited → optimistic update + API call
	const onCellEdited = useCallback(
		([col, row]: Item, newValue: { data: unknown }) => {
			const item = data[row];
			if (!item) return;

			const colDef = cols[col];
			const field = colDef.id as string;
			if (!field) return;

			const oldValue = item[field as keyof InventoryMain];
			setData((prev) => {
				const next = [...prev];
				next[row] = { ...item, [field]: newValue.data };
				return next;
			});

			api
				.updateInventory(item.ma_hang, { [field]: newValue.data })
				.catch(() => {
					setData((prev) => {
						const next = [...prev];
						next[row] = { ...item, [field]: oldValue };
						return next;
					});
				});
		},
		[data, cols],
	);

	// Column resize (session only)
	const onColumnResize = useCallback((column: GridColumn, newSize: number) => {
		setCols((prev) =>
			prev.map((c) => (c.id === column.id ? { ...c, width: newSize } : c)),
		);
	}, []);

	// Column reorder (session only)
	const onColumnMoved = useCallback((startIndex: number, endIndex: number) => {
		setCols((prev) => {
			const next = [...prev];
			const [moved] = next.splice(startIndex, 1);
			next.splice(endIndex, 0, moved);
			return next;
		});
	}, []);

	// Track mouse position for tooltip
	const onMouseMove = useCallback((e: React.MouseEvent) => {
		if (containerRef.current) {
			const rect = containerRef.current.getBoundingClientRect();
			mousePos.current = {
				x: e.clientX - rect.left,
				y: e.clientY - rect.top,
			};
		}
	}, []);

	// Tooltip for LBBQ=0
	const lbbqColIndex = useMemo(
		() => cols.findIndex((c) => c.id === "luong_ban_binh_quan_ngay"),
		[cols],
	);

	const onItemHovered = useCallback(
		(args: { kind: string; location?: readonly [number, number] }) => {
			if (args.kind === "cell" && args.location) {
				const [col, row] = args.location;
				if (col === lbbqColIndex) {
					const item = data[row];
					if (item && item.luong_ban_binh_quan_ngay === 0) {
						setTooltip({
							text: "CẦN ĐẨY HÀNG",
							x: mousePos.current.x,
							y: mousePos.current.y,
						});
						return;
					}
				}
			}
			setTooltip(null);
		},
		[data, lbbqColIndex],
	);

	// Get selected ma_hang values for export
	const selectedMaHangs = useMemo(() => {
		const result: string[] = [];
		if (selection.rows.length === 0) return result;
		for (let i = 0; i < data.length; i++) {
			if (data[i] && selection.rows.hasIndex(i)) {
				result.push(data[i].ma_hang);
			}
		}
		return result;
	}, [selection.rows, data]);

	// Export handler
	const handleExport = useCallback(async () => {
		const selectedCols = ALL_EXPORT_COLUMNS.filter((c) =>
			exportColumns.has(c.id),
		).map((c) => c.id);
		if (selectedCols.length === 0) return;

		setExporting(true);
		try {
			const blob = await api.exportInventory(
				selectedMaHangs,
				selectedCols,
				filterModel,
			);
			const url = URL.createObjectURL(blob);
			const a = document.createElement("a");
			a.href = url;
			a.download = `BaoCaoTonKho_${new Date().toISOString().slice(0, 10).replace(/-/g, "")}.xlsx`;
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);
			setShowExportModal(false);
		} catch (err) {
			alert(
				`Xuất báo cáo thất bại: ${err instanceof Error ? err.message : "Unknown"}`,
			);
		} finally {
			setExporting(false);
		}
	}, [exportColumns, selectedMaHangs, filterModel]);

	return (
		<div
			ref={containerRef}
			onMouseMove={onMouseMove}
			style={{
				width: "100%",
				height: "calc(100vh - 200px)",
				position: "relative",
			}}
		>
			{/* Header: title + filters + export */}
			<div
				style={{
					marginBottom: 8,
					display: "flex",
					alignItems: "center",
					gap: 10,
					flexWrap: "wrap",
				}}
			>
				<h2
					style={{
						margin: 0,
						fontSize: 16,
						fontWeight: 600,
						color: "#1e2330",
					}}
				>
					Kho hàng
				</h2>
				{loading && (
					<span style={{ color: "#7a7f8e", fontSize: 12 }}>Loading...</span>
				)}
				<span style={{ color: "#7a7f8e", fontSize: 12 }}>
					{totalRows.toLocaleString()} rows
					{selectedMaHangs.length > 0 && ` · ${selectedMaHangs.length} đã chọn`}
				</span>

				<div
					style={{
						marginLeft: "auto",
						display: "flex",
						alignItems: "center",
						gap: 8,
					}}
				>
					<FilterDropdown
						label="Mã BU"
						value={buFilter}
						onChange={setBuFilter}
						options={filterOptions.ma_bu}
					/>
					<FilterDropdown
						label="Mã nhóm hàng"
						value={nhomFilter}
						onChange={setNhomFilter}
						options={filterOptions.ma_nhom_hang}
					/>
					{(buFilter || nhomFilter) && (
						<button
							type="button"
							onClick={() => {
								setBuFilter("");
								setNhomFilter("");
							}}
							style={{
								padding: "5px 10px",
								fontSize: 11,
								background: "#fff",
								border: "1px solid #d5d8de",
								borderRadius: 4,
								cursor: "pointer",
								color: "#7a7f8e",
							}}
						>
							Xóa lọc
						</button>
					)}
					<button
						type="button"
						onClick={() => {
							setExportColumns(new Set(ALL_EXPORT_COLUMNS.map((c) => c.id)));
							setShowExportModal(true);
						}}
						style={{
							padding: "5px 12px",
							fontSize: 11,
							background: "#3a3f4b",
							color: "#fff",
							border: "none",
							borderRadius: 4,
							cursor: "pointer",
							fontWeight: 500,
						}}
					>
						Xuất báo cáo
					</button>
				</div>
			</div>

			<DataEditor
				columns={cols}
				rows={totalRows}
				getCellContent={getCellContent}
				onCellEdited={onCellEdited}
				onVisibleRegionChanged={onVisibleRegionChanged}
				onItemHovered={onItemHovered}
				onColumnResize={onColumnResize}
				onColumnMoved={onColumnMoved}
				gridSelection={selection}
				onGridSelectionChange={(sel) => {
					setSelection(sel);
					if (onRowSelect && sel.current?.cell) {
						const rowIdx = sel.current.cell[1];
						const item = data[rowIdx];
						if (item) onRowSelect(item.ma_hang);
					}
				}}
				smoothScrollX
				smoothScrollY
				rowMarkers="checkbox"
				width="100%"
				height="100%"
			/>

			{/* Tooltip */}
			{tooltip && (
				<div
					style={{
						position: "absolute",
						left: tooltip.x + 12,
						top: tooltip.y - 30,
						background: "#1e2330",
						color: "#fff",
						padding: "4px 10px",
						borderRadius: 4,
						fontSize: 11,
						whiteSpace: "nowrap",
						pointerEvents: "none",
						zIndex: 1000,
					}}
				>
					{tooltip.text}
				</div>
			)}

			{/* Export Modal */}
			{showExportModal && (
				<div
					style={{
						position: "fixed",
						inset: 0,
						background: "rgba(0,0,0,0.4)",
						display: "flex",
						alignItems: "center",
						justifyContent: "center",
						zIndex: 1000,
					}}
					onClick={() => setShowExportModal(false)}
				>
					<div
						style={{
							background: "#fff",
							borderRadius: 10,
							padding: 28,
							width: 420,
							maxWidth: "90vw",
							boxShadow: "0 8px 32px rgba(0,0,0,0.18)",
						}}
						onClick={(e) => e.stopPropagation()}
					>
						<h3
							style={{
								margin: "0 0 10px",
								fontSize: 15,
								fontWeight: 700,
								color: "#1e2330",
							}}
						>
							Xuất báo cáo tồn kho
						</h3>
						<p
							style={{
								fontSize: 12,
								color: "#7a7f8e",
								marginBottom: 16,
							}}
						>
							{selectedMaHangs.length > 0
								? `Đã chọn ${selectedMaHangs.length} dòng.`
								: "Chưa chọn dòng — sẽ xuất toàn bộ dữ liệu đang lọc."}{" "}
							Chọn cột cần xuất:
						</p>
						<div
							style={{
								display: "grid",
								gridTemplateColumns: "1fr 1fr",
								gap: "6px 16px",
								marginBottom: 16,
							}}
						>
							{ALL_EXPORT_COLUMNS.map((col) => (
								<label
									key={col.id}
									style={{
										display: "flex",
										alignItems: "center",
										gap: 6,
										fontSize: 12,
										color: "#3a3f4b",
										cursor: "pointer",
									}}
								>
									<input
										type="checkbox"
										checked={exportColumns.has(col.id)}
										onChange={() => {
											setExportColumns((prev) => {
												const next = new Set(prev);
												if (next.has(col.id)) {
													next.delete(col.id);
												} else {
													next.add(col.id);
												}
												return next;
											});
										}}
									/>
									{col.title}
								</label>
							))}
						</div>
						<div
							style={{
								display: "flex",
								gap: 8,
								justifyContent: "space-between",
							}}
						>
							<div style={{ display: "flex", gap: 8 }}>
								<button
									type="button"
									onClick={() =>
										setExportColumns(
											new Set(ALL_EXPORT_COLUMNS.map((c) => c.id)),
										)
									}
									style={{
										padding: "5px 10px",
										fontSize: 11,
										background: "#fff",
										border: "1px solid #d5d8de",
										borderRadius: 4,
										cursor: "pointer",
										color: "#7a7f8e",
									}}
								>
									Chọn tất cả
								</button>
								<button
									type="button"
									onClick={() => setExportColumns(new Set())}
									style={{
										padding: "5px 10px",
										fontSize: 11,
										background: "#fff",
										border: "1px solid #d5d8de",
										borderRadius: 4,
										cursor: "pointer",
										color: "#7a7f8e",
									}}
								>
									Bỏ chọn
								</button>
							</div>
							<div style={{ display: "flex", gap: 8 }}>
								<button
									type="button"
									onClick={() => setShowExportModal(false)}
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
									Hủy
								</button>
								<button
									type="button"
									onClick={handleExport}
									disabled={exporting || exportColumns.size === 0}
									style={{
										padding: "7px 16px",
										background: exportColumns.size > 0 ? "#3a3f4b" : "#e8eaed",
										color: exportColumns.size > 0 ? "#fff" : "#7a7f8e",
										border: "none",
										borderRadius: 5,
										cursor:
											exporting || exportColumns.size === 0
												? "not-allowed"
												: "pointer",
										fontSize: 12,
										fontWeight: 500,
									}}
								>
									{exporting ? "Đang xuất..." : "Xuất Excel"}
								</button>
							</div>
						</div>
					</div>
				</div>
			)}
		</div>
	);
}

// --- FilterDropdown component ---

function FilterDropdown({
	label,
	value,
	onChange,
	options,
}: {
	label: string;
	value: string;
	onChange: (v: string) => void;
	options: string[];
}) {
	const [open, setOpen] = useState(false);
	const ref = useRef<HTMLDivElement>(null);

	const filtered = useMemo(() => {
		if (!value) return options;
		return options.filter((o) => o.toLowerCase().includes(value.toLowerCase()));
	}, [options, value]);

	// Close on click outside
	useEffect(() => {
		const handler = (e: MouseEvent) => {
			if (ref.current && !ref.current.contains(e.target as Node))
				setOpen(false);
		};
		document.addEventListener("mousedown", handler);
		return () => document.removeEventListener("mousedown", handler);
	}, []);

	return (
		<div ref={ref} style={{ position: "relative" }}>
			<input
				placeholder={label}
				value={value}
				onChange={(e) => {
					onChange(e.target.value);
					setOpen(true);
				}}
				onFocus={() => setOpen(true)}
				style={{
					padding: "5px 8px",
					fontSize: 11,
					border: "1px solid #d5d8de",
					borderRadius: 4,
					width: 130,
					color: "#3a3f4b",
				}}
			/>
			{open && filtered.length > 0 && (
				<div
					style={{
						position: "absolute",
						top: "100%",
						left: 0,
						width: 180,
						maxHeight: 200,
						overflowY: "auto",
						background: "#fff",
						border: "1px solid #e8eaed",
						borderRadius: 4,
						boxShadow: "0 4px 12px rgba(0,0,0,0.1)",
						zIndex: 100,
					}}
				>
					{filtered.map((o) => (
						<div
							key={o}
							onClick={() => {
								onChange(o);
								setOpen(false);
							}}
							onMouseEnter={(e) =>
								(e.currentTarget.style.background = "#f5f6f8")
							}
							onMouseLeave={(e) => (e.currentTarget.style.background = "#fff")}
							style={{
								padding: "6px 10px",
								fontSize: 11,
								cursor: "pointer",
								color: "#3a3f4b",
							}}
						>
							{o}
						</div>
					))}
				</div>
			)}
		</div>
	);
}
