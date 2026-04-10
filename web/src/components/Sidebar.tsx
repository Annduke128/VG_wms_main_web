import { WarehouseSelector } from "./WarehouseSelector";

type Page = "overview" | "inventory" | "combo" | "orders" | "settings";

interface SidebarProps {
  active: Page;
  onChange: (page: Page) => void;
}

const NAV_ITEMS: { key: Page; label: string }[] = [
  { key: "overview", label: "Tổng quan" },
  { key: "inventory", label: "Kho chính" },
  { key: "combo", label: "Kho combo" },
  { key: "orders", label: "Nhập / Xuất" },
  { key: "settings", label: "Cài đặt" },
];

export type { Page };

export function Sidebar({ active, onChange }: SidebarProps) {
  return (
    <aside
      style={{
        width: 200,
        minHeight: "100vh",
        background: "#1e2330",
        color: "#c5c9d4",
        display: "flex",
        flexDirection: "column",
        flexShrink: 0,
      }}
    >
      <div
        style={{
          padding: "20px 16px",
          borderBottom: "1px solid rgba(255,255,255,0.06)",
          fontWeight: 700,
          fontSize: 15,
          color: "#fff",
          letterSpacing: 0.2,
        }}
      >
        Vĩnh Giang WMS
      </div>

      <WarehouseSelector />

      <nav style={{ flex: 1, padding: "8px 0" }}>
        {NAV_ITEMS.map((item) => (
          <button
            key={item.key}
            onClick={() => onChange(item.key)}
            style={{
              display: "block",
              width: "100%",
              padding: "10px 20px",
              border: "none",
              background: active === item.key ? "rgba(255,255,255,0.07)" : "transparent",
              color: active === item.key ? "#fff" : "#8b90a0",
              fontSize: 13,
              fontWeight: active === item.key ? 600 : 400,
              cursor: "pointer",
              textAlign: "left",
              borderLeft: active === item.key ? "2px solid #7c8cf5" : "2px solid transparent",
              transition: "all 0.12s ease",
            }}
          >
            {item.label}
          </button>
        ))}
      </nav>
    </aside>
  );
}
