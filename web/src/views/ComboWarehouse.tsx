import { useCallback, useEffect, useState } from "react";
import { api } from "../api/client";
import { useWarehouseStore } from "../stores/warehouseStore";
import type {
  Accessory,
  AccessoryInventory,
  ComboBOMAccessory,
  ComboBOMSemi,
  ComboDetail,
  ComboInventory,
  ComboMaster,
  ComboTransaction,
} from "../types/combo";

type Tab = "catalog" | "create" | "export" | "return" | "inventory" | "accessories";

const TABS: { key: Tab; label: string }[] = [
  { key: "catalog", label: "Danh mục combo" },
  { key: "create", label: "Tạo / Hủy combo" },
  { key: "export", label: "Xuất combo" },
  { key: "return", label: "Trả hàng combo" },
  { key: "inventory", label: "Tồn kho combo" },
  { key: "accessories", label: "Kho phụ kiện" },
];

/* ─── styles ───────────────────────────────────────── */

const card: React.CSSProperties = {
  background: "#fff",
  borderRadius: 8,
  padding: 20,
  border: "1px solid #e8eaed",
  marginBottom: 16,
};

const inputStyle: React.CSSProperties = {
  padding: "7px 10px",
  border: "1px solid #ddd",
  borderRadius: 6,
  fontSize: 13,
  width: "100%",
  boxSizing: "border-box",
};

const btnPrimary: React.CSSProperties = {
  padding: "8px 16px",
  background: "#6b7efa",
  color: "#fff",
  border: "none",
  borderRadius: 6,
  fontSize: 13,
  fontWeight: 600,
  cursor: "pointer",
};

const btnDanger: React.CSSProperties = {
  ...btnPrimary,
  background: "#e74c3c",
};

const btnOutline: React.CSSProperties = {
  padding: "8px 16px",
  background: "#fff",
  color: "#6b7efa",
  border: "1px solid #6b7efa",
  borderRadius: 6,
  fontSize: 13,
  fontWeight: 600,
  cursor: "pointer",
};

const thStyle: React.CSSProperties = {
  padding: "8px 12px",
  textAlign: "left",
  fontWeight: 600,
  fontSize: 12,
  color: "#7a7f8e",
  textTransform: "uppercase",
};

const tdStyle: React.CSSProperties = {
  padding: "8px 12px",
  fontSize: 13,
};

const labelStyle: React.CSSProperties = {
  fontSize: 12,
  fontWeight: 600,
  color: "#555",
  marginBottom: 4,
  display: "block",
};

/* ─── CatalogTab ───────────────────────────────────── */

function CatalogTab() {
  const { activeWarehouseId } = useWarehouseStore();
  const [masters, setMasters] = useState<ComboMaster[]>([]);
  const [selected, setSelected] = useState<ComboDetail | null>(null);
  const [loading, setLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({
    ma_combo: "",
    ten_combo: "",
    mo_ta: "",
  });
  const [bomSemi, setBomSemi] = useState<{ ma_component: string; so_luong: number }[]>([]);
  const [bomAccessory, setBomAccessory] = useState<{ ma_component: string; so_luong: number }[]>(
    [],
  );
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = (await api.listComboMasters()) as ComboMaster[];
      setMasters(data || []);
    } catch {
      setMasters([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const viewDetail = async (maCombo: string) => {
    try {
      const data = (await api.getComboDetail(maCombo)) as ComboDetail;
      setSelected(data);
    } catch (err) {
      console.error(err);
    }
  };

  const handleSave = async () => {
    if (!form.ma_combo || !form.ten_combo) {
      setError("Mã combo và Tên combo là bắt buộc");
      return;
    }
    setSaving(true);
    setError("");
    try {
      await api.saveComboMaster({
        ma_combo: form.ma_combo,
        warehouse_id: activeWarehouseId,
        ten_combo: form.ten_combo,
        mo_ta: form.mo_ta,
        bom_semi: bomSemi,
        bom_accessory: bomAccessory,
      });
      setShowForm(false);
      setForm({ ma_combo: "", ten_combo: "", mo_ta: "" });
      setBomSemi([]);
      setBomAccessory([]);
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Lỗi lưu combo");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (maCombo: string) => {
    if (!confirm(`Xóa combo "${maCombo}"?`)) return;
    try {
      await api.deleteComboMaster(maCombo);
      setSelected(null);
      load();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Lỗi xóa");
    }
  };

  const editCombo = (detail: ComboDetail) => {
    setForm({
      ma_combo: detail.ma_combo,
      ten_combo: detail.ten_combo,
      mo_ta: detail.mo_ta,
    });
    setBomSemi(
      (detail.bom_semi || []).map((b: ComboBOMSemi) => ({
        ma_component: b.ma_hang,
        so_luong: b.so_luong,
      })),
    );
    setBomAccessory(
      (detail.bom_accessory || []).map((b: ComboBOMAccessory) => ({
        ma_component: b.ma_phu_kien,
        so_luong: b.so_luong,
      })),
    );
    setShowForm(true);
    setSelected(null);
  };

  return (
    <div>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 16,
        }}
      >
        <h3 style={{ margin: 0, fontSize: 14, fontWeight: 600 }}>Danh mục combo</h3>
        <button
          style={btnPrimary}
          onClick={() => {
            setShowForm(!showForm);
            setForm({ ma_combo: "", ten_combo: "", mo_ta: "" });
            setBomSemi([]);
            setBomAccessory([]);
            setError("");
          }}
        >
          {showForm ? "Đóng" : "Thêm combo"}
        </button>
      </div>

      {showForm && (
        <div style={card}>
          <h4 style={{ margin: "0 0 12px", fontSize: 13, fontWeight: 600 }}>
            {form.ma_combo && masters.some((m) => m.ma_combo === form.ma_combo)
              ? "Sửa combo"
              : "Thêm combo mới"}
          </h4>
          {error && <p style={{ color: "#e74c3c", fontSize: 12, margin: "0 0 8px" }}>{error}</p>}

          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr 1fr",
              gap: 12,
              marginBottom: 12,
            }}
          >
            <div>
              <label style={labelStyle}>Mã combo</label>
              <input
                style={inputStyle}
                value={form.ma_combo}
                onChange={(e) => setForm({ ...form, ma_combo: e.target.value })}
                placeholder="VD: COMBO-001"
              />
            </div>
            <div>
              <label style={labelStyle}>Tên combo</label>
              <input
                style={inputStyle}
                value={form.ten_combo}
                onChange={(e) => setForm({ ...form, ten_combo: e.target.value })}
                placeholder="Tên hiển thị"
              />
            </div>
            <div>
              <label style={labelStyle}>Mô tả</label>
              <input
                style={inputStyle}
                value={form.mo_ta}
                onChange={(e) => setForm({ ...form, mo_ta: e.target.value })}
                placeholder="Ghi chú"
              />
            </div>
          </div>

          {/* BOM Semi */}
          <div style={{ marginBottom: 12 }}>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                marginBottom: 8,
              }}
            >
              <label style={{ ...labelStyle, margin: 0 }}>BOM — Bán thành phẩm (kho chính)</label>
              <button
                style={{ ...btnOutline, padding: "4px 10px", fontSize: 12 }}
                onClick={() => setBomSemi([...bomSemi, { ma_component: "", so_luong: 1 }])}
              >
                + Thêm
              </button>
            </div>
            {bomSemi.map((item, i) => (
              <div key={i} style={{ display: "flex", gap: 8, marginBottom: 4 }}>
                <input
                  style={{ ...inputStyle, flex: 2 }}
                  value={item.ma_component}
                  onChange={(e) => {
                    const n = [...bomSemi];
                    n[i] = { ...n[i], ma_component: e.target.value };
                    setBomSemi(n);
                  }}
                  placeholder="Mã hàng"
                />
                <input
                  style={{ ...inputStyle, flex: 1 }}
                  type="number"
                  min={0}
                  step="any"
                  value={item.so_luong}
                  onChange={(e) => {
                    const n = [...bomSemi];
                    n[i] = { ...n[i], so_luong: Number(e.target.value) };
                    setBomSemi(n);
                  }}
                  placeholder="SL"
                />
                <button
                  style={{
                    background: "none",
                    border: "none",
                    color: "#e74c3c",
                    cursor: "pointer",
                    fontSize: 16,
                  }}
                  onClick={() => setBomSemi(bomSemi.filter((_, j) => j !== i))}
                >
                  ✕
                </button>
              </div>
            ))}
          </div>

          {/* BOM Accessory */}
          <div style={{ marginBottom: 12 }}>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                marginBottom: 8,
              }}
            >
              <label style={{ ...labelStyle, margin: 0 }}>BOM — Phụ kiện</label>
              <button
                style={{ ...btnOutline, padding: "4px 10px", fontSize: 12 }}
                onClick={() =>
                  setBomAccessory([...bomAccessory, { ma_component: "", so_luong: 1 }])
                }
              >
                + Thêm
              </button>
            </div>
            {bomAccessory.map((item, i) => (
              <div key={i} style={{ display: "flex", gap: 8, marginBottom: 4 }}>
                <input
                  style={{ ...inputStyle, flex: 2 }}
                  value={item.ma_component}
                  onChange={(e) => {
                    const n = [...bomAccessory];
                    n[i] = { ...n[i], ma_component: e.target.value };
                    setBomAccessory(n);
                  }}
                  placeholder="Mã phụ kiện"
                />
                <input
                  style={{ ...inputStyle, flex: 1 }}
                  type="number"
                  min={0}
                  step="any"
                  value={item.so_luong}
                  onChange={(e) => {
                    const n = [...bomAccessory];
                    n[i] = { ...n[i], so_luong: Number(e.target.value) };
                    setBomAccessory(n);
                  }}
                  placeholder="SL"
                />
                <button
                  style={{
                    background: "none",
                    border: "none",
                    color: "#e74c3c",
                    cursor: "pointer",
                    fontSize: 16,
                  }}
                  onClick={() => setBomAccessory(bomAccessory.filter((_, j) => j !== i))}
                >
                  ✕
                </button>
              </div>
            ))}
          </div>

          <button style={btnPrimary} disabled={saving} onClick={handleSave}>
            {saving ? "Đang lưu..." : "Lưu combo"}
          </button>
        </div>
      )}

      {/* Detail view */}
      {selected && (
        <div style={card}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: 12,
            }}
          >
            <h4 style={{ margin: 0, fontSize: 13, fontWeight: 600 }}>
              {selected.ma_combo} — {selected.ten_combo}
            </h4>
            <div style={{ display: "flex", gap: 8 }}>
              <button style={btnOutline} onClick={() => editCombo(selected)}>
                Sửa
              </button>
              <button style={btnDanger} onClick={() => handleDelete(selected.ma_combo)}>
                Xóa
              </button>
              <button
                style={{
                  background: "none",
                  border: "none",
                  fontSize: 16,
                  cursor: "pointer",
                  color: "#7a7f8e",
                }}
                onClick={() => setSelected(null)}
              >
                ✕
              </button>
            </div>
          </div>
          {selected.mo_ta && (
            <p style={{ fontSize: 12, color: "#888", margin: "0 0 12px" }}>{selected.mo_ta}</p>
          )}

          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 }}>
            <div>
              <h5
                style={{
                  margin: "0 0 8px",
                  fontSize: 12,
                  fontWeight: 600,
                  color: "#555",
                }}
              >
                Bán thành phẩm
              </h5>
              {(selected.bom_semi || []).length === 0 ? (
                <p style={{ fontSize: 12, color: "#aaa" }}>Chưa có</p>
              ) : (
                <table style={{ width: "100%", borderCollapse: "collapse" }}>
                  <thead>
                    <tr style={{ borderBottom: "1px solid #eee" }}>
                      <th style={thStyle}>Mã hàng</th>
                      <th style={thStyle}>Tên</th>
                      <th style={{ ...thStyle, textAlign: "right" }}>SL/combo</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(selected.bom_semi || []).map((b: ComboBOMSemi) => (
                      <tr key={b.id} style={{ borderBottom: "1px solid #f5f5f5" }}>
                        <td style={tdStyle}>{b.ma_hang}</td>
                        <td style={tdStyle}>{b.ten_san_pham || "—"}</td>
                        <td style={{ ...tdStyle, textAlign: "right" }}>{b.so_luong}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
            <div>
              <h5
                style={{
                  margin: "0 0 8px",
                  fontSize: 12,
                  fontWeight: 600,
                  color: "#555",
                }}
              >
                Phụ kiện
              </h5>
              {(selected.bom_accessory || []).length === 0 ? (
                <p style={{ fontSize: 12, color: "#aaa" }}>Chưa có</p>
              ) : (
                <table style={{ width: "100%", borderCollapse: "collapse" }}>
                  <thead>
                    <tr style={{ borderBottom: "1px solid #eee" }}>
                      <th style={thStyle}>Mã PK</th>
                      <th style={thStyle}>Tên</th>
                      <th style={{ ...thStyle, textAlign: "right" }}>SL/combo</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(selected.bom_accessory || []).map((b: ComboBOMAccessory) => (
                      <tr key={b.id} style={{ borderBottom: "1px solid #f5f5f5" }}>
                        <td style={tdStyle}>{b.ma_phu_kien}</td>
                        <td style={tdStyle}>{b.ten_phu_kien || "—"}</td>
                        <td style={{ ...tdStyle, textAlign: "right" }}>{b.so_luong}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Master list */}
      {loading ? (
        <p style={{ color: "#888" }}>Đang tải...</p>
      ) : masters.length === 0 ? (
        <p style={{ color: "#888" }}>Chưa có combo nào. Nhấn "Thêm combo" để bắt đầu.</p>
      ) : (
        <table
          style={{
            width: "100%",
            borderCollapse: "collapse",
            background: "#fff",
            borderRadius: 8,
          }}
        >
          <thead>
            <tr style={{ borderBottom: "2px solid #eee" }}>
              <th style={thStyle}>Mã combo</th>
              <th style={thStyle}>Tên combo</th>
              <th style={thStyle}>Mô tả</th>
              <th style={thStyle}>Trạng thái</th>
              <th style={thStyle}>Ngày tạo</th>
            </tr>
          </thead>
          <tbody>
            {masters.map((m) => (
              <tr
                key={m.ma_combo}
                style={{ borderBottom: "1px solid #f0f0f0", cursor: "pointer" }}
                onClick={() => viewDetail(m.ma_combo)}
              >
                <td style={{ ...tdStyle, fontWeight: 600, color: "#6b7efa" }}>{m.ma_combo}</td>
                <td style={tdStyle}>{m.ten_combo}</td>
                <td style={tdStyle}>{m.mo_ta || "—"}</td>
                <td style={tdStyle}>
                  <span
                    style={{
                      padding: "2px 8px",
                      borderRadius: 10,
                      fontSize: 11,
                      fontWeight: 600,
                      background: m.active ? "#e8f5e9" : "#fce4ec",
                      color: m.active ? "#2e7d32" : "#c62828",
                    }}
                  >
                    {m.active ? "Hoạt động" : "Ngưng"}
                  </span>
                </td>
                <td style={tdStyle}>{new Date(m.created_at).toLocaleDateString("vi-VN")}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

/* ─── ComboActionTab (Create/Cancel, Out, Return) ─── */

function ComboActionTab({ actionType }: { actionType: "create" | "export" | "return" }) {
  const { activeWarehouseId } = useWarehouseStore();
  const [masters, setMasters] = useState<ComboMaster[]>([]);
  const [maCombo, setMaCombo] = useState("");
  const [soLuong, setSoLuong] = useState("");
  const [note, setNote] = useState("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<{ ok: boolean; msg: string } | null>(null);

  useEffect(() => {
    api
      .listComboMasters()
      .then((d) => setMasters((d as ComboMaster[]) || []))
      .catch(() => {});
  }, []);

  const titles: Record<
    string,
    { title: string; btnCreate: string; btnCancel?: string; color: string }
  > = {
    create: {
      title: "Tạo combo",
      btnCreate: "Tạo combo",
      btnCancel: "Hủy combo",
      color: "#27ae60",
    },
    export: { title: "Xuất combo", btnCreate: "Xuất combo", color: "#2196f3" },
    return: {
      title: "Trả hàng combo",
      btnCreate: "Trả hàng",
      color: "#ff9800",
    },
  };
  const cfg = titles[actionType];

  const execute = async (action: "create" | "cancel" | "out" | "return") => {
    if (!maCombo || !soLuong || Number(soLuong) <= 0) {
      setResult({ ok: false, msg: "Chọn combo và nhập số lượng > 0" });
      return;
    }
    setLoading(true);
    setResult(null);
    try {
      const body = {
        ma_combo: maCombo,
        so_luong: Number(soLuong),
        note,
        warehouse_id: activeWarehouseId,
      };
      if (action === "create") await api.createCombo(body);
      else if (action === "cancel") await api.cancelCombo(body);
      else if (action === "out") await api.comboOut(body);
      else await api.comboReturn(body);

      const labels = {
        create: "Tạo",
        cancel: "Hủy",
        out: "Xuất",
        return: "Trả",
      };
      setResult({
        ok: true,
        msg: `${labels[action]} ${soLuong} combo "${maCombo}" thành công`,
      });
      setSoLuong("");
      setNote("");
    } catch (err) {
      setResult({
        ok: false,
        msg: err instanceof Error ? err.message : "Lỗi thao tác",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={card}>
      <h3 style={{ margin: "0 0 16px", fontSize: 14, fontWeight: 600 }}>{cfg.title}</h3>

      {result && (
        <div
          style={{
            padding: "8px 12px",
            borderRadius: 6,
            marginBottom: 12,
            fontSize: 13,
            background: result.ok ? "#e8f5e9" : "#fce4ec",
            color: result.ok ? "#2e7d32" : "#c62828",
          }}
        >
          {result.msg}
        </div>
      )}

      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr 2fr",
          gap: 12,
          marginBottom: 12,
        }}
      >
        <div>
          <label style={labelStyle}>Combo</label>
          <select style={inputStyle} value={maCombo} onChange={(e) => setMaCombo(e.target.value)}>
            <option value="">-- Chọn combo --</option>
            {masters
              .filter((m) => m.active)
              .map((m) => (
                <option key={m.ma_combo} value={m.ma_combo}>
                  {m.ma_combo} — {m.ten_combo}
                </option>
              ))}
          </select>
        </div>
        <div>
          <label style={labelStyle}>Số lượng</label>
          <input
            style={inputStyle}
            type="number"
            min={1}
            step="any"
            value={soLuong}
            onChange={(e) => setSoLuong(e.target.value)}
            placeholder="VD: 10"
          />
        </div>
        <div>
          <label style={labelStyle}>Ghi chú</label>
          <input
            style={inputStyle}
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder="Ghi chú (không bắt buộc)"
          />
        </div>
      </div>

      <div style={{ display: "flex", gap: 8 }}>
        <button
          style={{ ...btnPrimary, background: cfg.color }}
          disabled={loading}
          onClick={() => {
            if (actionType === "create") execute("create");
            else if (actionType === "export") execute("out");
            else execute("return");
          }}
        >
          {loading ? "Đang xử lý..." : cfg.btnCreate}
        </button>
        {cfg.btnCancel && (
          <button style={btnDanger} disabled={loading} onClick={() => execute("cancel")}>
            {cfg.btnCancel}
          </button>
        )}
      </div>
    </div>
  );
}

/* ─── InventoryTab ─────────────────────────────────── */

function InventoryTab() {
  const { activeWarehouseId } = useWarehouseStore();
  const [items, setItems] = useState<ComboInventory[]>([]);
  const [transactions, setTransactions] = useState<ComboTransaction[]>([]);
  const [loading, setLoading] = useState(false);
  const [txPage, setTxPage] = useState(1);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [inv, txn] = await Promise.all([
        api.comboInventory(activeWarehouseId),
        api.comboTransactions(activeWarehouseId, txPage, 20),
      ]);
      setItems((inv as ComboInventory[]) || []);
      setTransactions((txn as ComboTransaction[]) || []);
    } catch {
      setItems([]);
      setTransactions([]);
    } finally {
      setLoading(false);
    }
  }, [txPage, activeWarehouseId]);

  useEffect(() => {
    load();
  }, [load]);

  const txTypeLabels: Record<string, { label: string; color: string }> = {
    CREATE: { label: "Tạo", color: "#27ae60" },
    CANCEL: { label: "Hủy", color: "#e74c3c" },
    OUT: { label: "Xuất", color: "#2196f3" },
    RETURN: { label: "Trả", color: "#ff9800" },
  };

  if (loading) return <p style={{ color: "#888" }}>Đang tải...</p>;

  return (
    <div>
      {/* Inventory table */}
      <div style={card}>
        <h3 style={{ margin: "0 0 12px", fontSize: 14, fontWeight: 600 }}>Tồn kho combo</h3>
        {items.length === 0 ? (
          <p style={{ color: "#888", fontSize: 13 }}>Chưa có dữ liệu tồn kho combo.</p>
        ) : (
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr style={{ borderBottom: "2px solid #eee" }}>
                <th style={thStyle}>Mã combo</th>
                <th style={thStyle}>Tên combo</th>
                <th style={{ ...thStyle, textAlign: "right" }}>Tồn</th>
                <th style={{ ...thStyle, textAlign: "right" }}>Đã tạo</th>
                <th style={{ ...thStyle, textAlign: "right" }}>Đã xuất</th>
                <th style={{ ...thStyle, textAlign: "right" }}>Đã trả</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.ma_combo} style={{ borderBottom: "1px solid #f0f0f0" }}>
                  <td style={{ ...tdStyle, fontWeight: 600 }}>{item.ma_combo}</td>
                  <td style={tdStyle}>{item.ten_combo || "—"}</td>
                  <td style={{ ...tdStyle, textAlign: "right", fontWeight: 600 }}>
                    {item.so_ton.toLocaleString("vi-VN")}
                  </td>
                  <td style={{ ...tdStyle, textAlign: "right" }}>
                    {item.so_nhap.toLocaleString("vi-VN")}
                  </td>
                  <td style={{ ...tdStyle, textAlign: "right" }}>
                    {item.so_xuat.toLocaleString("vi-VN")}
                  </td>
                  <td style={{ ...tdStyle, textAlign: "right" }}>
                    {item.so_tra.toLocaleString("vi-VN")}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Transactions */}
      <div style={card}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: 12,
          }}
        >
          <h3 style={{ margin: 0, fontSize: 14, fontWeight: 600 }}>Lịch sử nghiệp vụ</h3>
          <button style={btnOutline} onClick={load}>
            Tải lại
          </button>
        </div>
        {transactions.length === 0 ? (
          <p style={{ color: "#888", fontSize: 13 }}>Chưa có giao dịch nào.</p>
        ) : (
          <>
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <thead>
                <tr style={{ borderBottom: "2px solid #eee" }}>
                  <th style={thStyle}>ID</th>
                  <th style={thStyle}>Loại</th>
                  <th style={thStyle}>Combo</th>
                  <th style={{ ...thStyle, textAlign: "right" }}>Số lượng</th>
                  <th style={thStyle}>Ghi chú</th>
                  <th style={thStyle}>Thời gian</th>
                </tr>
              </thead>
              <tbody>
                {transactions.map((tx) => {
                  const t = txTypeLabels[tx.transaction_type] || {
                    label: tx.transaction_type,
                    color: "#666",
                  };
                  return (
                    <tr key={tx.id} style={{ borderBottom: "1px solid #f0f0f0" }}>
                      <td style={tdStyle}>{tx.id}</td>
                      <td style={tdStyle}>
                        <span
                          style={{
                            padding: "2px 8px",
                            borderRadius: 10,
                            fontSize: 11,
                            fontWeight: 600,
                            background: `${t.color}18`,
                            color: t.color,
                          }}
                        >
                          {t.label}
                        </span>
                      </td>
                      <td style={tdStyle}>
                        {tx.ma_combo}
                        {tx.ten_combo ? ` — ${tx.ten_combo}` : ""}
                      </td>
                      <td
                        style={{
                          ...tdStyle,
                          textAlign: "right",
                          fontWeight: 600,
                        }}
                      >
                        {tx.so_luong.toLocaleString("vi-VN")}
                      </td>
                      <td style={tdStyle}>{tx.note || "—"}</td>
                      <td style={tdStyle}>{new Date(tx.created_at).toLocaleString("vi-VN")}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            <div
              style={{
                display: "flex",
                gap: 8,
                marginTop: 12,
                justifyContent: "center",
              }}
            >
              <button
                style={btnOutline}
                disabled={txPage <= 1}
                onClick={() => setTxPage((p) => p - 1)}
              >
                ← Trước
              </button>
              <span style={{ padding: "8px 12px", fontSize: 13 }}>Trang {txPage}</span>
              <button
                style={btnOutline}
                disabled={transactions.length < 20}
                onClick={() => setTxPage((p) => p + 1)}
              >
                Sau →
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

/* ─── AccessoriesTab ───────────────────────────────── */

function AccessoriesTab() {
  const { activeWarehouseId } = useWarehouseStore();
  const [accessories, setAccessories] = useState<Accessory[]>([]);
  const [inventory, setInventory] = useState<AccessoryInventory[]>([]);
  const [loading, setLoading] = useState(false);
  const [showAdd, setShowAdd] = useState(false);
  const [addForm, setAddForm] = useState({
    ma_phu_kien: "",
    ten_phu_kien: "",
    don_vi_tinh: "",
  });
  const [stockForm, setStockForm] = useState({
    ma_phu_kien: "",
    so_luong: "",
    note: "",
  });
  const [result, setResult] = useState<{ ok: boolean; msg: string } | null>(null);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [acc, inv] = await Promise.all([
        api.listAccessories(),
        api.accessoryInventory(activeWarehouseId),
      ]);
      setAccessories((acc as Accessory[]) || []);
      setInventory((inv as AccessoryInventory[]) || []);
    } catch {
      setAccessories([]);
      setInventory([]);
    } finally {
      setLoading(false);
    }
  }, [activeWarehouseId]);

  useEffect(() => {
    load();
  }, [load]);

  const handleAddAccessory = async () => {
    if (!addForm.ma_phu_kien || !addForm.ten_phu_kien) {
      setResult({ ok: false, msg: "Mã và tên phụ kiện là bắt buộc" });
      return;
    }
    setSaving(true);
    try {
      await api.createAccessory(addForm);
      setShowAdd(false);
      setAddForm({ ma_phu_kien: "", ten_phu_kien: "", don_vi_tinh: "" });
      setResult({ ok: true, msg: "Thêm phụ kiện thành công" });
      load();
    } catch (err) {
      setResult({ ok: false, msg: err instanceof Error ? err.message : "Lỗi" });
    } finally {
      setSaving(false);
    }
  };

  const handleStockIn = async () => {
    if (!stockForm.ma_phu_kien || !stockForm.so_luong || Number(stockForm.so_luong) <= 0) {
      setResult({ ok: false, msg: "Chọn phụ kiện và nhập số lượng > 0" });
      return;
    }
    setSaving(true);
    try {
      await api.accessoryStockIn({
        ma_phu_kien: stockForm.ma_phu_kien,
        so_luong: Number(stockForm.so_luong),
        note: stockForm.note,
        warehouse_id: activeWarehouseId,
      });
      setResult({
        ok: true,
        msg: `Nhập ${stockForm.so_luong} phụ kiện "${stockForm.ma_phu_kien}" thành công`,
      });
      setStockForm({ ...stockForm, so_luong: "", note: "" });
      load();
    } catch (err) {
      setResult({ ok: false, msg: err instanceof Error ? err.message : "Lỗi" });
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <p style={{ color: "#888" }}>Đang tải...</p>;

  return (
    <div>
      {result && (
        <div
          style={{
            padding: "8px 12px",
            borderRadius: 6,
            marginBottom: 12,
            fontSize: 13,
            background: result.ok ? "#e8f5e9" : "#fce4ec",
            color: result.ok ? "#2e7d32" : "#c62828",
          }}
        >
          {result.msg}
        </div>
      )}

      {/* Add accessory */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 12,
        }}
      >
        <h3 style={{ margin: 0, fontSize: 14, fontWeight: 600 }}>Danh mục phụ kiện</h3>
        <button style={btnPrimary} onClick={() => setShowAdd(!showAdd)}>
          {showAdd ? "Đóng" : "Thêm phụ kiện"}
        </button>
      </div>

      {showAdd && (
        <div style={card}>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr 1fr auto",
              gap: 12,
              alignItems: "end",
            }}
          >
            <div>
              <label style={labelStyle}>Mã phụ kiện</label>
              <input
                style={inputStyle}
                value={addForm.ma_phu_kien}
                onChange={(e) => setAddForm({ ...addForm, ma_phu_kien: e.target.value })}
                placeholder="VD: PK-001"
              />
            </div>
            <div>
              <label style={labelStyle}>Tên phụ kiện</label>
              <input
                style={inputStyle}
                value={addForm.ten_phu_kien}
                onChange={(e) => setAddForm({ ...addForm, ten_phu_kien: e.target.value })}
                placeholder="Tên hiển thị"
              />
            </div>
            <div>
              <label style={labelStyle}>Đơn vị tính</label>
              <input
                style={inputStyle}
                value={addForm.don_vi_tinh}
                onChange={(e) => setAddForm({ ...addForm, don_vi_tinh: e.target.value })}
                placeholder="VD: cái, bộ"
              />
            </div>
            <button style={btnPrimary} disabled={saving} onClick={handleAddAccessory}>
              Lưu
            </button>
          </div>
        </div>
      )}

      {/* Accessory list */}
      {accessories.length === 0 ? (
        <p style={{ color: "#888", fontSize: 13, marginBottom: 16 }}>Chưa có phụ kiện nào.</p>
      ) : (
        <table
          style={{
            width: "100%",
            borderCollapse: "collapse",
            background: "#fff",
            borderRadius: 8,
            marginBottom: 16,
          }}
        >
          <thead>
            <tr style={{ borderBottom: "2px solid #eee" }}>
              <th style={thStyle}>Mã PK</th>
              <th style={thStyle}>Tên phụ kiện</th>
              <th style={thStyle}>ĐVT</th>
              <th style={thStyle}>Ngày tạo</th>
            </tr>
          </thead>
          <tbody>
            {accessories.map((a) => (
              <tr key={a.ma_phu_kien} style={{ borderBottom: "1px solid #f0f0f0" }}>
                <td style={{ ...tdStyle, fontWeight: 600 }}>{a.ma_phu_kien}</td>
                <td style={tdStyle}>{a.ten_phu_kien}</td>
                <td style={tdStyle}>{a.don_vi_tinh || "—"}</td>
                <td style={tdStyle}>{new Date(a.created_at).toLocaleDateString("vi-VN")}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {/* Stock in */}
      <div style={card}>
        <h3 style={{ margin: "0 0 12px", fontSize: 14, fontWeight: 600 }}>Nhập kho phụ kiện</h3>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr 2fr auto",
            gap: 12,
            alignItems: "end",
          }}
        >
          <div>
            <label style={labelStyle}>Phụ kiện</label>
            <select
              style={inputStyle}
              value={stockForm.ma_phu_kien}
              onChange={(e) => setStockForm({ ...stockForm, ma_phu_kien: e.target.value })}
            >
              <option value="">-- Chọn --</option>
              {accessories.map((a) => (
                <option key={a.ma_phu_kien} value={a.ma_phu_kien}>
                  {a.ma_phu_kien} — {a.ten_phu_kien}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label style={labelStyle}>Số lượng</label>
            <input
              style={inputStyle}
              type="number"
              min={1}
              step="any"
              value={stockForm.so_luong}
              onChange={(e) => setStockForm({ ...stockForm, so_luong: e.target.value })}
            />
          </div>
          <div>
            <label style={labelStyle}>Ghi chú</label>
            <input
              style={inputStyle}
              value={stockForm.note}
              onChange={(e) => setStockForm({ ...stockForm, note: e.target.value })}
              placeholder="Ghi chú nhập kho"
            />
          </div>
          <button
            style={{ ...btnPrimary, background: "#27ae60" }}
            disabled={saving}
            onClick={handleStockIn}
          >
            Nhập kho
          </button>
        </div>
      </div>

      {/* Inventory */}
      <div style={card}>
        <h3 style={{ margin: "0 0 12px", fontSize: 14, fontWeight: 600 }}>Tồn kho phụ kiện</h3>
        {inventory.length === 0 ? (
          <p style={{ color: "#888", fontSize: 13 }}>Chưa có dữ liệu tồn kho.</p>
        ) : (
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr style={{ borderBottom: "2px solid #eee" }}>
                <th style={thStyle}>Mã PK</th>
                <th style={thStyle}>Tên phụ kiện</th>
                <th style={thStyle}>ĐVT</th>
                <th style={{ ...thStyle, textAlign: "right" }}>Tồn kho</th>
                <th style={thStyle}>Cập nhật</th>
              </tr>
            </thead>
            <tbody>
              {inventory.map((inv) => (
                <tr key={inv.ma_phu_kien} style={{ borderBottom: "1px solid #f0f0f0" }}>
                  <td style={{ ...tdStyle, fontWeight: 600 }}>{inv.ma_phu_kien}</td>
                  <td style={tdStyle}>{inv.ten_phu_kien || "—"}</td>
                  <td style={tdStyle}>{inv.don_vi_tinh || "—"}</td>
                  <td style={{ ...tdStyle, textAlign: "right", fontWeight: 600 }}>
                    {inv.so_ton.toLocaleString("vi-VN")}
                  </td>
                  <td style={tdStyle}>{new Date(inv.updated_at).toLocaleDateString("vi-VN")}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

/* ─── Main export ──────────────────────────────────── */

export function ComboWarehouse() {
  const [tab, setTab] = useState<Tab>("catalog");

  return (
    <div>
      <h2
        style={{
          margin: "0 0 16px",
          fontSize: 18,
          fontWeight: 700,
          color: "#2d3142",
        }}
      >
        Kho combo
      </h2>

      {/* Tab bar */}
      <div
        style={{
          display: "flex",
          gap: 0,
          marginBottom: 20,
          borderBottom: "2px solid #eee",
        }}
      >
        {TABS.map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            style={{
              padding: "10px 18px",
              border: "none",
              background: "transparent",
              fontSize: 13,
              fontWeight: tab === t.key ? 600 : 400,
              color: tab === t.key ? "#6b7efa" : "#7a7f8e",
              cursor: "pointer",
              borderBottom: tab === t.key ? "2px solid #6b7efa" : "2px solid transparent",
              marginBottom: -2,
              transition: "all 0.15s ease",
            }}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      {tab === "catalog" && <CatalogTab />}
      {tab === "create" && <ComboActionTab actionType="create" />}
      {tab === "export" && <ComboActionTab actionType="export" />}
      {tab === "return" && <ComboActionTab actionType="return" />}
      {tab === "inventory" && <InventoryTab />}
      {tab === "accessories" && <AccessoriesTab />}
    </div>
  );
}
