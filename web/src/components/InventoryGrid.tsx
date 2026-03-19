import DataEditor, {
	CompactSelection,
	type GridCell,
	GridCellKind,
	type GridColumn,
	type GridSelection,
	type Item,
} from "@glideapps/glide-data-grid";
import { useCallback, useEffect, useRef, useState } from "react";
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

const columns: GridColumn[] = [
	{ id: "ma_hang", title: "Mã hàng", width: 120 },
	{ id: "ten_san_pham", title: "Tên sản phẩm", width: 250 },
	{ id: "so_ton", title: "Số tồn", width: 100 },
	{ id: "so_nhap", title: "Số nhập", width: 100 },
	{ id: "so_xuat", title: "Số xuất", width: 100 },
	{ id: "tien_ton", title: "Tiền tồn", width: 120 },
	{ id: "tien_nhap", title: "Tiền nhập", width: 120 },
	{ id: "tien_xuat", title: "Tiền xuất", width: 120 },
	{ id: "so_ngay_ton", title: "Số ngày tồn", width: 110 },
	{ id: "luong_ban_binh_quan_ngay", title: "LBBQ/ngày", width: 120 },
];

interface InventoryGridProps {
	onRowSelect?: (maHang: string) => void;
}

export function InventoryGrid({ onRowSelect }: InventoryGridProps = {}) {
	const [data, setData] = useState<InventoryMain[]>([]);
	const [totalRows, setTotalRows] = useState(0);
	const [loading, setLoading] = useState(false);
	const [sortModel] = useState<SortItem[]>([]);
	const [filterModel] = useState<Record<string, FilterItem>>({});
	const [selection, setSelection] = useState<GridSelection>({
		columns: CompactSelection.empty(),
		rows: CompactSelection.empty(),
	});

	const loadedRanges = useRef<Set<string>>(new Set());

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

	// Initial load
	useEffect(() => {
		loadedRanges.current.clear();
		setData([]);
		fetchRange(0, PAGE_SIZE);
	}, [fetchRange]);

	// Visible region changed → fetch more data
	const onVisibleRegionChanged = useCallback(
		(range: { x: number; y: number; width: number; height: number }) => {
			const start = Math.floor(range.y / PAGE_SIZE) * PAGE_SIZE;
			const end = start + PAGE_SIZE;
			fetchRange(start, end);

			// Prefetch next page
			if (end < totalRows) {
				fetchRange(end, end + PAGE_SIZE);
			}
		},
		[fetchRange, totalRows],
	);

	// Get cell content
	const getCellContent = useCallback(
		([col, row]: Item): GridCell => {
			const item = data[row];
			if (!item) {
				return { kind: GridCellKind.Loading, allowOverlay: false };
			}

			const colDef = columns[col];
			const value = item[colDef.id as keyof InventoryMain];

			if (typeof value === "number") {
				return {
					kind: GridCellKind.Number,
					data: value,
					displayData: value.toLocaleString("vi-VN"),
					allowOverlay: true,
					readonly: colDef.id === "ma_hang",
				};
			}

			return {
				kind: GridCellKind.Text,
				data: String(value ?? ""),
				displayData: String(value ?? ""),
				allowOverlay: true,
				readonly: colDef.id === "ma_hang",
			};
		},
		[data],
	);

	// Cell edited → optimistic update + API call
	const onCellEdited = useCallback(
		([col, row]: Item, newValue: { data: unknown }) => {
			const item = data[row];
			if (!item) return;

			const colDef = columns[col];
			const field = colDef.id as string;
			if (!field) return;

			// Optimistic update
			const oldValue = item[field as keyof InventoryMain];
			setData((prev) => {
				const next = [...prev];
				next[row] = { ...item, [field]: newValue.data };
				return next;
			});

			// API call
			api
				.updateInventory(item.ma_hang, { [field]: newValue.data })
				.catch(() => {
					// Rollback on error
					setData((prev) => {
						const next = [...prev];
						next[row] = { ...item, [field]: oldValue };
						return next;
					});
				});
		},
		[data],
	);

	return (
		<div style={{ width: "100%", height: "calc(100vh - 200px)" }}>
			<div
				style={{
					marginBottom: 8,
					display: "flex",
					alignItems: "center",
					gap: 12,
				}}
			>
				<h2 style={{ margin: 0, fontSize: 20 }}>Kho hàng</h2>
				{loading && <span style={{ color: "#888" }}>Loading...</span>}
				<span style={{ color: "#888" }}>{totalRows.toLocaleString()} rows</span>
			</div>
			<DataEditor
				columns={columns}
				rows={totalRows}
				getCellContent={getCellContent}
				onCellEdited={onCellEdited}
				onVisibleRegionChanged={onVisibleRegionChanged}
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
				rowMarkers="both"
				width="100%"
				height="100%"
			/>
		</div>
	);
}
