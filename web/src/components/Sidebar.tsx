import type { ReactNode } from "react";

type Page = "overview" | "inventory" | "orders" | "settings";

interface SidebarProps {
	active: Page;
	onChange: (page: Page) => void;
}

const NAV_ITEMS: { key: Page; label: string; icon: ReactNode }[] = [
	{ key: "overview", label: "Tổng quan", icon: "📊" },
	{ key: "inventory", label: "Kho", icon: "📦" },
	{ key: "orders", label: "Nhập / Xuất", icon: "🔄" },
	{ key: "settings", label: "Settings", icon: "⚙️" },
];

export type { Page };

export function Sidebar({ active, onChange }: SidebarProps) {
	return (
		<aside
			style={{
				width: 220,
				minHeight: "100vh",
				background: "#1a1a2e",
				color: "#e0e0e0",
				display: "flex",
				flexDirection: "column",
				flexShrink: 0,
			}}
		>
			<div
				style={{
					padding: "20px 16px",
					borderBottom: "1px solid #2a2a4a",
					fontWeight: 700,
					fontSize: 18,
					color: "#fff",
					letterSpacing: 0.5,
				}}
			>
				WMS Dashboard
			</div>

			<nav style={{ flex: 1, padding: "8px 0" }}>
				{NAV_ITEMS.map((item) => (
					<button
						key={item.key}
						onClick={() => onChange(item.key)}
						style={{
							display: "flex",
							alignItems: "center",
							gap: 10,
							width: "100%",
							padding: "12px 20px",
							border: "none",
							background: active === item.key ? "#16213e" : "transparent",
							color: active === item.key ? "#fff" : "#a0a0b8",
							fontSize: 14,
							fontWeight: active === item.key ? 600 : 400,
							cursor: "pointer",
							textAlign: "left",
							borderLeft:
								active === item.key
									? "3px solid #4fc3f7"
									: "3px solid transparent",
							transition: "all 0.15s ease",
						}}
					>
						<span style={{ fontSize: 16 }}>{item.icon}</span>
						{item.label}
					</button>
				))}
			</nav>
		</aside>
	);
}
