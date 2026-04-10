import { useEffect } from "react";
import { useWarehouseStore } from "../stores/warehouseStore";

export function WarehouseSelector() {
  const { warehouses, activeWarehouseId, setActiveWarehouse, fetchWarehouses, loading } =
    useWarehouseStore();

  useEffect(() => {
    fetchWarehouses();
  }, [fetchWarehouses]);

  if (loading && warehouses.length === 0) {
    return (
      <div style={{ padding: "8px 16px", fontSize: 12, color: "#8b90a0" }}>Đang tải kho...</div>
    );
  }

  return (
    <div
      style={{
        padding: "10px 16px",
        borderBottom: "1px solid rgba(255,255,255,0.06)",
      }}
    >
      <label
        style={{
          display: "block",
          fontSize: 10,
          color: "#8b90a0",
          marginBottom: 4,
          textTransform: "uppercase",
          letterSpacing: 0.5,
        }}
      >
        Kho
      </label>
      <select
        value={activeWarehouseId}
        onChange={(e) => setActiveWarehouse(Number(e.target.value))}
        style={{
          width: "100%",
          padding: "6px 8px",
          fontSize: 13,
          background: "#282d3e",
          color: "#fff",
          border: "1px solid rgba(255,255,255,0.1)",
          borderRadius: 4,
          outline: "none",
          cursor: "pointer",
        }}
      >
        {warehouses.map((w) => (
          <option key={w.id} value={w.id}>
            {w.name}
          </option>
        ))}
      </select>
    </div>
  );
}
