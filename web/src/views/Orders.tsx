import { useCallback, useEffect, useState } from "react";
import { api } from "../api/client";
import { useWarehouseStore } from "../stores/warehouseStore";
import type { OrderListItem } from "../types/dashboard";

// ── Shared filter state ──
interface OrderFilters {
  dateFrom: string;
  dateTo: string;
  month: string;
  maBu: string;
  maNhomHang: string;
}

const EMPTY_FILTERS: OrderFilters = {
  dateFrom: "",
  dateTo: "",
  month: "",
  maBu: "",
  maNhomHang: "",
};

export function Orders() {
  const { activeWarehouseId } = useWarehouseStore();
  const [filters, setFilters] = useState<OrderFilters>(EMPTY_FILTERS);
  const [buOptions, setBuOptions] = useState<string[]>([]);
  const [nhomOptions, setNhomOptions] = useState<string[]>([]);

  // Load filter dropdown options once
  useEffect(() => {
    api
      .inventoryFilterOptions(activeWarehouseId)
      .then((resp) => {
        const r = resp as { ma_bu: string[]; ma_nhom_hang: string[] };
        setBuOptions(r.ma_bu || []);
        setNhomOptions(r.ma_nhom_hang || []);
      })
      .catch(console.error);
  }, [activeWarehouseId]);

  const updateFilter = (key: keyof OrderFilters, val: string) => {
    setFilters((prev) => {
      const next = { ...prev, [key]: val };
      // If month is set, clear dateFrom/dateTo
      if (key === "month" && val) {
        next.dateFrom = "";
        next.dateTo = "";
      }
      // If dateFrom or dateTo is set, clear month
      if ((key === "dateFrom" || key === "dateTo") && val) {
        next.month = "";
      }
      return next;
    });
  };

  // Build month options: last 12 months
  const monthOptions: string[] = [];
  const now = new Date();
  for (let i = 0; i < 12; i++) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1);
    const mm = String(d.getMonth() + 1).padStart(2, "0");
    monthOptions.push(`${mm}/${d.getFullYear()}`);
  }

  return (
    <div>
      <h2
        style={{
          margin: "0 0 16px 0",
          fontSize: 16,
          fontWeight: 600,
          color: "#1e2330",
        }}
      >
        Nhập / Xuất
      </h2>

      {/* ── Filter bar ── */}
      <div
        style={{
          display: "flex",
          gap: 10,
          marginBottom: 16,
          flexWrap: "wrap",
          alignItems: "flex-end",
          background: "#fff",
          padding: "12px 16px",
          borderRadius: 6,
          border: "1px solid #e8eaed",
        }}
      >
        <FilterField label="Từ ngày">
          <input
            type="date"
            value={filters.dateFrom}
            onChange={(e) => updateFilter("dateFrom", e.target.value)}
            style={inputStyle}
          />
        </FilterField>
        <FilterField label="Đến ngày">
          <input
            type="date"
            value={filters.dateTo}
            onChange={(e) => updateFilter("dateTo", e.target.value)}
            style={inputStyle}
          />
        </FilterField>
        <FilterField label="Tháng/Năm">
          <select
            value={filters.month}
            onChange={(e) => updateFilter("month", e.target.value)}
            style={inputStyle}
          >
            <option value="">Tất cả</option>
            {monthOptions.map((m) => (
              <option key={m} value={m}>
                {m}
              </option>
            ))}
          </select>
        </FilterField>
        <FilterField label="Mã BU">
          <FilterDropdown
            value={filters.maBu}
            options={buOptions}
            onChange={(v) => updateFilter("maBu", v)}
            placeholder="Tất cả"
          />
        </FilterField>
        <FilterField label="Mã nhóm hàng">
          <FilterDropdown
            value={filters.maNhomHang}
            options={nhomOptions}
            onChange={(v) => updateFilter("maNhomHang", v)}
            placeholder="Tất cả"
          />
        </FilterField>
        <button
          onClick={() => setFilters(EMPTY_FILTERS)}
          style={{
            padding: "7px 12px",
            border: "1px solid #d5d8de",
            borderRadius: 5,
            background: "#fff",
            cursor: "pointer",
            fontSize: 12,
            color: "#7a7f8e",
          }}
        >
          Xóa bộ lọc
        </button>
      </div>

      {/* ── Two-column layout ── */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr",
          gap: 16,
        }}
      >
        <OrderPanel
          type="in"
          title="Nhập kho"
          color="#3a7d4f"
          bgColor="#e8f5ed"
          filters={filters}
        />
        <OrderPanel
          type="out"
          title="Xuất kho"
          color="#b83b3b"
          bgColor="#fde8e8"
          filters={filters}
        />
      </div>
    </div>
  );
}

// ── Order Panel (one per side) ──

function OrderPanel({
  type,
  title,
  color,
  bgColor,
  filters,
}: {
  type: "in" | "out";
  title: string;
  color: string;
  bgColor: string;
  filters: OrderFilters;
}) {
  const { activeWarehouseId } = useWarehouseStore();
  const [orders, setOrders] = useState<OrderListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const limit = 50;

  const fetchOrders = useCallback(async () => {
    setLoading(true);
    try {
      // Build API filter params
      const apiFilters: Record<string, string | number> = {
        type,
        page,
        limit,
        warehouse_id: activeWarehouseId,
      };
      if (filters.month) {
        apiFilters.month = filters.month;
      } else {
        if (filters.dateFrom) {
          // Convert yyyy-mm-dd → dd/mm/yyyy for API
          const [y, m, d] = filters.dateFrom.split("-");
          apiFilters.date_from = `${d}/${m}/${y}`;
        }
        if (filters.dateTo) {
          const [y, m, d] = filters.dateTo.split("-");
          apiFilters.date_to = `${d}/${m}/${y}`;
        }
      }
      if (filters.maBu) apiFilters.ma_bu = filters.maBu;
      if (filters.maNhomHang) apiFilters.ma_nhom_hang = filters.maNhomHang;

      const resp = (await api.listOrders(apiFilters)) as {
        data: OrderListItem[];
        total: number;
      };
      setOrders(resp.data || []);
      setTotal(resp.total || 0);
    } catch (err) {
      console.error(`Fetch ${type} orders error:`, err);
    } finally {
      setLoading(false);
    }
  }, [type, page, filters, activeWarehouseId]);

  useEffect(() => {
    setPage(1);
  }, [filters]);

  useEffect(() => {
    fetchOrders();
  }, [fetchOrders]);

  const totalPages = Math.ceil(total / limit);

  return (
    <div
      style={{
        background: "#fff",
        borderRadius: 6,
        border: "1px solid #e8eaed",
        overflow: "hidden",
      }}
    >
      {/* Header */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          padding: "12px 16px",
          borderBottom: "1px solid #e8eaed",
          background: bgColor,
        }}
      >
        <h3
          style={{
            margin: 0,
            fontSize: 14,
            fontWeight: 600,
            color,
          }}
        >
          {title}{" "}
          <span style={{ fontWeight: 400, fontSize: 12, color: "#7a7f8e" }}>
            ({total.toLocaleString("vi-VN")})
          </span>
        </h3>
        <button
          onClick={() => setShowForm(!showForm)}
          style={{
            padding: "5px 12px",
            background: color,
            color: "#fff",
            border: "none",
            borderRadius: 4,
            cursor: "pointer",
            fontSize: 11,
            fontWeight: 500,
          }}
        >
          + Tạo đơn
        </button>
      </div>

      {/* Form */}
      {showForm && (
        <CreateOrderForm
          type={type}
          color={color}
          onCreated={() => {
            setShowForm(false);
            fetchOrders();
          }}
        />
      )}

      {/* Table */}
      <div style={{ padding: "0 0 12px 0" }}>
        {loading ? (
          <p
            style={{
              color: "#888",
              textAlign: "center",
              padding: 20,
              margin: 0,
            }}
          >
            Đang tải...
          </p>
        ) : (
          <table
            style={{
              width: "100%",
              borderCollapse: "collapse",
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
                <th style={thStyle}>Mã hàng</th>
                <th style={thStyle}>Tên sản phẩm</th>
                <th style={thStyle}>Mã thùng</th>
                <th style={{ ...thStyle, textAlign: "right" }}>Số lượng</th>
                <th style={thStyle}>Ngày</th>
              </tr>
            </thead>
            <tbody>
              {orders.map((o) => (
                <tr key={`${type}-${o.id}`} style={{ borderBottom: "1px solid #f0f0f0" }}>
                  <td style={tdStyle}>{o.ma_hang}</td>
                  <td style={tdStyle}>{o.ten_san_pham}</td>
                  <td style={tdStyle}>{o.batch_code || "\u2014"}</td>
                  <td style={{ ...tdStyle, textAlign: "right" }}>
                    {o.so_luong.toLocaleString("vi-VN")}
                  </td>
                  <td style={tdStyle}>{new Date(o.ngay_nhan_hang).toLocaleDateString("vi-VN")}</td>
                </tr>
              ))}
              {orders.length === 0 && (
                <tr>
                  <td
                    colSpan={5}
                    style={{
                      padding: 20,
                      textAlign: "center",
                      color: "#888",
                    }}
                  >
                    {type === "in" ? "Chưa có đơn nhập." : "Chưa có đơn xuất."}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div
            style={{
              display: "flex",
              justifyContent: "center",
              gap: 8,
              marginTop: 10,
            }}
          >
            <button
              disabled={page <= 1}
              onClick={() => setPage(page - 1)}
              style={paginationBtnStyle(page <= 1)}
            >
              &#8592;
            </button>
            <span style={{ padding: "4px 8px", fontSize: 12, color: "#7a7f8e" }}>
              {page}/{totalPages}
            </span>
            <button
              disabled={page >= totalPages}
              onClick={() => setPage(page + 1)}
              style={paginationBtnStyle(page >= totalPages)}
            >
              &#8594;
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

// ── Create Order Form (type-specific) ──

function CreateOrderForm({
  type,
  color,
  onCreated,
}: {
  type: "in" | "out";
  color: string;
  onCreated: () => void;
}) {
  const { activeWarehouseId } = useWarehouseStore();
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
        warehouse_id: activeWarehouseId,
      };
      if (type === "in") {
        body.batch_code = batchCode || maHang;
      }

      const resp = (await api.createOrder(body)) as {
        allocations?: { batch_code: string; qty: number }[];
      };

      if (type === "out" && resp.allocations) {
        const detail = resp.allocations.map((a) => `${a.batch_code}: ${a.qty}`).join(", ");
        setResult(`FIFO: ${detail}`);
      } else {
        setResult("Thành công!");
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
        padding: "12px 16px",
        borderBottom: "1px solid #e8eaed",
        display: "flex",
        gap: 8,
        alignItems: "flex-end",
        flexWrap: "wrap",
      }}
    >
      <FormField label="Mã hàng">
        <input
          value={maHang}
          onChange={(e) => setMaHang(e.target.value)}
          placeholder="SP001"
          required
          style={{ ...inputStyle, width: 110 }}
        />
      </FormField>

      {type === "in" && (
        <FormField label="Mã thùng">
          <input
            value={batchCode}
            onChange={(e) => setBatchCode(e.target.value)}
            placeholder="BATCH-001"
            style={{ ...inputStyle, width: 110 }}
          />
        </FormField>
      )}

      <FormField label="Số lượng">
        <input
          type="number"
          value={soLuong}
          onChange={(e) => setSoLuong(e.target.value)}
          placeholder="0"
          required
          min="0.01"
          step="0.01"
          style={{ ...inputStyle, width: 80 }}
        />
      </FormField>

      <button
        type="submit"
        disabled={submitting}
        style={{
          padding: "7px 14px",
          background: color,
          color: "#fff",
          border: "none",
          borderRadius: 5,
          cursor: submitting ? "not-allowed" : "pointer",
          fontSize: 12,
          fontWeight: 500,
        }}
      >
        {submitting ? "..." : type === "in" ? "Nhập" : "Xuất"}
      </button>

      {result && (
        <span
          style={{
            fontSize: 11,
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

// ── Filter Dropdown (type-to-filter) ──

function FilterDropdown({
  value,
  options,
  onChange,
  placeholder,
}: {
  value: string;
  options: string[];
  onChange: (v: string) => void;
  placeholder: string;
}) {
  const [search, setSearch] = useState("");
  const [open, setOpen] = useState(false);

  const filtered = search
    ? options.filter((o) => o.toLowerCase().includes(search.toLowerCase()))
    : options;

  return (
    <div style={{ position: "relative" }}>
      <input
        value={value || search}
        onChange={(e) => {
          setSearch(e.target.value);
          if (!e.target.value) onChange("");
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
        onBlur={() => setTimeout(() => setOpen(false), 200)}
        placeholder={placeholder}
        style={{ ...inputStyle, width: 120 }}
      />
      {open && filtered.length > 0 && (
        <div
          style={{
            position: "absolute",
            top: "100%",
            left: 0,
            right: 0,
            maxHeight: 200,
            overflowY: "auto",
            background: "#fff",
            border: "1px solid #d5d8de",
            borderRadius: 4,
            zIndex: 10,
            boxShadow: "0 2px 8px rgba(0,0,0,0.1)",
          }}
        >
          <div
            style={dropdownItemStyle}
            onMouseDown={() => {
              onChange("");
              setSearch("");
              setOpen(false);
            }}
          >
            <em style={{ color: "#7a7f8e" }}>{placeholder}</em>
          </div>
          {filtered.map((opt) => (
            <div
              key={opt}
              style={{
                ...dropdownItemStyle,
                fontWeight: opt === value ? 600 : 400,
                background: opt === value ? "#f0f4ff" : "#fff",
              }}
              onMouseDown={() => {
                onChange(opt);
                setSearch("");
                setOpen(false);
              }}
            >
              {opt}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

// ── Small helpers ──

function FilterField({ label, children }: { label: string; children: React.ReactNode }) {
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
        {label}
      </label>
      {children}
    </div>
  );
}

function FormField({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <label
        style={{
          fontSize: 10,
          color: "#7a7f8e",
          display: "block",
          marginBottom: 2,
        }}
      >
        {label}
      </label>
      {children}
    </div>
  );
}

// ── Styles ──

const inputStyle: React.CSSProperties = {
  padding: "6px 8px",
  borderRadius: 4,
  border: "1px solid #d5d8de",
  fontSize: 12,
  color: "#3a3f4b",
};

const thStyle: React.CSSProperties = {
  padding: "8px 12px",
  fontSize: 11,
  fontWeight: 600,
  color: "#7a7f8e",
};

const tdStyle: React.CSSProperties = {
  padding: "7px 12px",
};

const dropdownItemStyle: React.CSSProperties = {
  padding: "6px 10px",
  fontSize: 12,
  cursor: "pointer",
};

function paginationBtnStyle(disabled: boolean): React.CSSProperties {
  return {
    padding: "4px 10px",
    borderRadius: 4,
    border: "1px solid #ccc",
    cursor: disabled ? "not-allowed" : "pointer",
    background: "#fff",
    fontSize: 12,
  };
}
