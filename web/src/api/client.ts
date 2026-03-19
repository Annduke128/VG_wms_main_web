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

	updateInventory: (maHang: string, fields: Record<string, unknown>) =>
		request(`/inventory/${encodeURIComponent(maHang)}`, {
			method: "PATCH",
			body: JSON.stringify(fields),
		}),

	bulkUpdate: (body: unknown) =>
		request("/inventory/bulk-update", {
			method: "POST",
			body: JSON.stringify(body),
		}),

	getJob: (jobId: string) => request(`/jobs/${jobId}`),

	// Kanban Inbound
	listKanbanInbound: () => request("/kanban/inbound"),
	createKanbanInbound: (body: unknown) =>
		request("/kanban/inbound", { method: "POST", body: JSON.stringify(body) }),
	moveKanbanInbound: (id: number, body: unknown) =>
		request(`/kanban/inbound/${id}/move`, {
			method: "POST",
			body: JSON.stringify(body),
		}),

	// Kanban Outbound
	listKanbanOutbound: () => request("/kanban/outbound"),
	createKanbanOutbound: (body: unknown) =>
		request("/kanban/outbound", { method: "POST", body: JSON.stringify(body) }),
	moveKanbanOutbound: (id: number, body: unknown) =>
		request(`/kanban/outbound/${id}/move`, {
			method: "POST",
			body: JSON.stringify(body),
		}),

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

	// Dashboard
	dashboardSummary: () => request("/dashboard/summary"),
	dashboardCharts: (weeks = 4) => request(`/dashboard/charts?weeks=${weeks}`),

	// Inventory lots & alerts
	inventoryLots: (maHang: string) =>
		request(`/inventory/lots?ma_hang=${encodeURIComponent(maHang)}`),
	inventoryAlerts: () => request("/inventory/alerts"),

	// Orders
	listOrders: (type?: string, page = 1, limit = 50) => {
		const params = new URLSearchParams({
			page: String(page),
			limit: String(limit),
		});
		if (type) params.set("type", type);
		return request(`/orders?${params}`);
	},
	createOrder: (body: unknown) =>
		request("/orders", { method: "POST", body: JSON.stringify(body) }),

	// Thresholds
	getThresholds: (maHang?: string) => {
		const params = maHang ? `?ma_hang=${encodeURIComponent(maHang)}` : "";
		return request(`/thresholds${params}`);
	},
	saveThreshold: (body: unknown) =>
		request("/thresholds", { method: "POST", body: JSON.stringify(body) }),
};
