import { useState } from "react";
import type { Page } from "./components/Sidebar";
import { Sidebar } from "./components/Sidebar";
import { ComboWarehouse } from "./views/ComboWarehouse";
import { Inventory } from "./views/Inventory";
import { Orders } from "./views/Orders";
import { Overview } from "./views/Overview";
import { Settings } from "./views/Settings";

function App() {
	const [page, setPage] = useState<Page>("overview");

	return (
		<div style={{ display: "flex", minHeight: "100vh", background: "#f7f8fa" }}>
			<Sidebar active={page} onChange={setPage} />

			<main style={{ flex: 1, padding: 24, overflow: "auto" }}>
				{page === "overview" && <Overview />}
				{page === "inventory" && <Inventory />}
				{page === "combo" && <ComboWarehouse />}
				{page === "orders" && <Orders />}
				{page === "settings" && <Settings />}
			</main>
		</div>
	);
}

export default App;
