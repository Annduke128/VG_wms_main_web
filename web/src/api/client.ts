import type { CreateWarehouseRequest, UpdateWarehouseRequest, Warehouse } from "../types/warehouse";

const API_BASE = import.meta.env.VITE_API_URL || "/api";

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || "Request failed");
  }
  return res.json();
}

export const api = {
  // Inventory
  inventoryGrid: (body: unknown) =>
    request("/inventory/grid", { method: "POST", body: JSON.stringify(body) }),

  updateInventory: (maHang: string, fields: Record<string, unknown>, warehouseId: number) =>
    request(`/inventory/${encodeURIComponent(maHang)}?warehouse_id=${warehouseId}`, {
      method: "PATCH",
      body: JSON.stringify(fields),
    }),

  bulkUpdate: (body: unknown) =>
    request("/inventory/bulk-update", {
      method: "POST",
      body: JSON.stringify(body),
    }),

  getJob: (jobId: string) => request(`/jobs/${jobId}`),

  // Import
  importFile: (type: string, file: File) => {
    const form = new FormData();
    form.append("file", file);
    return fetch(`${API_BASE}/import/${type}`, {
      method: "POST",
      body: form,
    }).then((r) => r.json());
  },

  downloadInventoryTemplate: async () => {
    const res = await fetch(`${API_BASE}/import/inventory/template`);
    if (!res.ok) throw new Error("Download failed");
    return res.blob();
  },

  getImportBatch: (batchId?: number) => {
    const path = batchId ? `/import/batches/${batchId}` : "/import/batches/latest";
    return request(path);
  },

  // Dashboard
  dashboardSummary: (warehouseId: number) =>
    request(`/dashboard/summary?warehouse_id=${warehouseId}`),
  dashboardCharts: (warehouseId: number, weeks = 4) =>
    request(`/dashboard/charts?weeks=${weeks}&warehouse_id=${warehouseId}`),

  // Dashboard extras
  dashboardZeroSales: (warehouseId: number) =>
    request(`/dashboard/zero-sales?warehouse_id=${warehouseId}`),
  dashboardRestockAlerts: (warehouseId: number) =>
    request(`/dashboard/restock-alerts?warehouse_id=${warehouseId}`),

  // Warehouses
  listWarehouses: () => request<Warehouse[]>("/warehouses"),
  getWarehouse: (id: number) => request<Warehouse>(`/warehouses/${id}`),
  createWarehouse: (data: CreateWarehouseRequest) =>
    request<Warehouse>("/warehouses", { method: "POST", body: JSON.stringify(data) }),
  updateWarehouse: (id: number, data: UpdateWarehouseRequest) =>
    request<Warehouse>(`/warehouses/${id}`, { method: "PATCH", body: JSON.stringify(data) }),

  // Inventory lots & alerts
  inventoryLots: (maHang: string, warehouseId: number) =>
    request(`/inventory/lots?ma_hang=${encodeURIComponent(maHang)}&warehouse_id=${warehouseId}`),
  inventoryAlerts: (warehouseId: number) =>
    request(`/inventory/alerts?warehouse_id=${warehouseId}`),

  // Orders
  listOrders: (
    filters: {
      type?: string;
      page?: number;
      limit?: number;
      date_from?: string;
      date_to?: string;
      month?: string;
      ma_bu?: string;
      ma_nhom_hang?: string;
      warehouse_id?: string | number;
    } = {},
  ) => {
    const params = new URLSearchParams({
      page: String(filters.page || 1),
      limit: String(filters.limit || 50),
    });
    if (filters.type) params.set("type", filters.type);
    if (filters.date_from) params.set("date_from", filters.date_from);
    if (filters.date_to) params.set("date_to", filters.date_to);
    if (filters.month) params.set("month", filters.month);
    if (filters.ma_bu) params.set("ma_bu", filters.ma_bu);
    if (filters.ma_nhom_hang) params.set("ma_nhom_hang", filters.ma_nhom_hang);
    if (filters.warehouse_id) params.set("warehouse_id", String(filters.warehouse_id));
    return request(`/orders?${params}`);
  },
  createOrder: (body: unknown) =>
    request("/orders", { method: "POST", body: JSON.stringify(body) }),

  // Admin
  recalcAll: () => request<{ job_id: string }>("/inventory/recalc-all", { method: "POST" }),

  resetAll: (confirmText: string) =>
    request("/admin/reset-all", {
      method: "POST",
      body: JSON.stringify({ confirm_text: confirmText }),
    }),

  // Filter options for inventory grid
  inventoryFilterOptions: (warehouseId: number) =>
    request<{ ma_bu: string[]; ma_nhom_hang: string[] }>(
      `/inventory/filter-options?warehouse_id=${warehouseId}`,
    ),

  // Export inventory to Excel
  exportInventory: async (
    maHangs: string[],
    columns: string[],
    warehouseId: number,
    filterModel?: Record<string, unknown>,
  ) => {
    const res = await fetch(`${API_BASE}/inventory/export`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        ma_hang: maHangs,
        columns,
        warehouse_id: warehouseId,
        filter_model: filterModel || {},
      }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }));
      throw new Error((err as { error?: string }).error || "Export failed");
    }
    return res.blob();
  },

  // Combo Masters
  listComboMasters: () => request("/combo/masters"),
  getComboDetail: (maCombo: string) => request(`/combo/masters/${encodeURIComponent(maCombo)}`),
  saveComboMaster: (body: unknown) =>
    request("/combo/masters", { method: "POST", body: JSON.stringify(body) }),
  deleteComboMaster: (maCombo: string) =>
    request(`/combo/masters/${encodeURIComponent(maCombo)}`, {
      method: "DELETE",
    }),

  // Combo Operations
  createCombo: (body: {
    ma_combo: string;
    so_luong: number;
    note?: string;
    warehouse_id: number;
  }) => request("/combo/create", { method: "POST", body: JSON.stringify(body) }),
  cancelCombo: (body: {
    ma_combo: string;
    so_luong: number;
    note?: string;
    warehouse_id: number;
  }) => request("/combo/cancel", { method: "POST", body: JSON.stringify(body) }),
  comboOut: (body: { ma_combo: string; so_luong: number; note?: string; warehouse_id: number }) =>
    request("/combo/out", { method: "POST", body: JSON.stringify(body) }),
  comboReturn: (body: {
    ma_combo: string;
    so_luong: number;
    note?: string;
    warehouse_id: number;
  }) => request("/combo/return", { method: "POST", body: JSON.stringify(body) }),

  // Combo Inventory & Transactions
  comboInventory: (warehouseId: number) => request(`/combo/inventory?warehouse_id=${warehouseId}`),
  comboTransactions: (warehouseId: number, page = 1, limit = 50) =>
    request(`/combo/transactions?warehouse_id=${warehouseId}&page=${page}&limit=${limit}`),

  // Accessories
  listAccessories: () => request("/accessories"),
  createAccessory: (body: { ma_phu_kien: string; ten_phu_kien: string; don_vi_tinh: string }) =>
    request("/accessories", { method: "POST", body: JSON.stringify(body) }),
  accessoryInventory: (warehouseId: number) =>
    request(`/accessories/inventory?warehouse_id=${warehouseId}`),
  accessoryStockIn: (body: {
    ma_phu_kien: string;
    so_luong: number;
    note?: string;
    warehouse_id: number;
  }) =>
    request("/accessories/stock-in", {
      method: "POST",
      body: JSON.stringify(body),
    }),

  // Thresholds
  getThresholds: (maHang: string, warehouseId: number) => {
    const params = `?ma_hang=${encodeURIComponent(maHang)}&warehouse_id=${warehouseId}`;
    return request(`/thresholds${params}`);
  },
  saveThreshold: (body: unknown) =>
    request("/thresholds", { method: "POST", body: JSON.stringify(body) }),
};
