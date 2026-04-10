import { useCallback, useState } from "react";
import { api } from "../api/client";
import { InventoryGrid } from "../components/InventoryGrid";
import { useWarehouseStore } from "../stores/warehouseStore";
import type { InventoryLot } from "../types/dashboard";

export function Inventory() {
  const { activeWarehouseId } = useWarehouseStore();
  const [selectedMaHang, setSelectedMaHang] = useState<string | null>(null);
  const [lots, setLots] = useState<InventoryLot[]>([]);
  const [lotsLoading, setLotsLoading] = useState(false);

  const handleRowSelect = useCallback(
    async (maHang: string) => {
      if (maHang === selectedMaHang) {
        setSelectedMaHang(null);
        setLots([]);
        return;
      }
      setSelectedMaHang(maHang);
      setLotsLoading(true);
      try {
        const data = (await api.inventoryLots(maHang, activeWarehouseId)) as InventoryLot[];
        setLots(data || []);
      } catch (err) {
        console.error("Fetch lots error:", err);
        setLots([]);
      } finally {
        setLotsLoading(false);
      }
    },
    [selectedMaHang, activeWarehouseId],
  );

  return (
    <div>
      <InventoryGrid
        key={activeWarehouseId}
        warehouseId={activeWarehouseId}
        onRowSelect={handleRowSelect}
      />

      {selectedMaHang && (
        <div
          style={{
            marginTop: 16,
            background: "#fff",
            borderRadius: 8,
            padding: 20,
            border: "1px solid #e8eaed",
          }}
        >
          <div
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              marginBottom: 12,
            }}
          >
            <h3
              style={{
                margin: 0,
                fontSize: 13,
                fontWeight: 600,
                color: "#3a3f4b",
              }}
            >
              Lô hàng — <span style={{ color: "#6b7efa" }}>{selectedMaHang}</span>
            </h3>
            <button
              onClick={() => {
                setSelectedMaHang(null);
                setLots([]);
              }}
              style={{
                background: "none",
                border: "none",
                fontSize: 16,
                cursor: "pointer",
                color: "#7a7f8e",
              }}
            >
              ✕
            </button>
          </div>

          {lotsLoading ? (
            <p style={{ color: "#888" }}>Đang tải...</p>
          ) : lots.length === 0 ? (
            <p style={{ color: "#888" }}>Chưa có lô hàng nào.</p>
          ) : (
            <table
              style={{
                width: "100%",
                borderCollapse: "collapse",
                fontSize: 13,
              }}
            >
              <thead>
                <tr style={{ borderBottom: "2px solid #eee", textAlign: "left" }}>
                  <th style={{ padding: "8px 12px" }}>Mã thùng (batch)</th>
                  <th style={{ padding: "8px 12px" }}>Ngày nhập</th>
                  <th style={{ padding: "8px 12px", textAlign: "right" }}>SL nhập</th>
                  <th style={{ padding: "8px 12px", textAlign: "right" }}>SL đã xuất</th>
                  <th style={{ padding: "8px 12px", textAlign: "right" }}>SL còn lại</th>
                </tr>
              </thead>
              <tbody>
                {lots.map((lot) => (
                  <tr key={lot.id} style={{ borderBottom: "1px solid #f0f0f0" }}>
                    <td style={{ padding: "8px 12px", fontWeight: 600 }}>{lot.batch_code}</td>
                    <td style={{ padding: "8px 12px" }}>
                      {new Date(lot.received_at).toLocaleDateString("vi-VN")}
                    </td>
                    <td style={{ padding: "8px 12px", textAlign: "right" }}>
                      {lot.qty_in.toLocaleString("vi-VN")}
                    </td>
                    <td style={{ padding: "8px 12px", textAlign: "right" }}>
                      {lot.qty_out.toLocaleString("vi-VN")}
                    </td>
                    <td
                      style={{
                        padding: "8px 12px",
                        textAlign: "right",
                        fontWeight: 600,
                      }}
                    >
                      {lot.qty_remaining.toLocaleString("vi-VN")}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}
    </div>
  );
}
