import { useState } from "react";
import { ImportPanel } from "./components/ImportPanel";
import { InventoryGrid } from "./components/InventoryGrid";
import { KanbanBoard } from "./components/KanbanBoard";

type Tab = "inventory" | "kanban-inbound" | "kanban-outbound" | "import";

const TABS: { key: Tab; label: string }[] = [
	{ key: "inventory", label: "Inventory" },
	{ key: "kanban-inbound", label: "Kanban Inbound" },
	{ key: "kanban-outbound", label: "Kanban Outbound" },
	{ key: "import", label: "Import" },
];

function App() {
	const [activeTab, setActiveTab] = useState<Tab>("inventory");

	return (
		<div style={{ fontFamily: "system-ui, sans-serif" }}>
			<nav
				style={{
					display: "flex",
					gap: 0,
					borderBottom: "2px solid #e0e0e0",
					padding: "0 16px",
					background: "#fafafa",
				}}
			>
				{TABS.map((tab) => (
					<button
						key={tab.key}
						onClick={() => setActiveTab(tab.key)}
						style={{
							padding: "12px 24px",
							border: "none",
							borderBottom:
								activeTab === tab.key
									? "2px solid #1976d2"
									: "2px solid transparent",
							background: "none",
							color: activeTab === tab.key ? "#1976d2" : "#666",
							fontWeight: activeTab === tab.key ? 600 : 400,
							cursor: "pointer",
							fontSize: 14,
						}}
					>
						{tab.label}
					</button>
				))}
			</nav>

			<main style={{ padding: "0 16px" }}>
				{activeTab === "inventory" && <InventoryGrid />}
				{activeTab === "kanban-inbound" && <KanbanBoard type="inbound" />}
				{activeTab === "kanban-outbound" && <KanbanBoard type="outbound" />}
				{activeTab === "import" && <ImportPanel />}
			</main>
		</div>
	);
}

export default App;
