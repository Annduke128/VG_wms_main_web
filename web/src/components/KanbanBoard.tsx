import { useCallback, useEffect, useState } from "react";
import { api } from "../api/client";
import type {
	InboundStage,
	KanbanInbound,
	KanbanOutbound,
	OutboundStage,
} from "../types/kanban";
import { INBOUND_STAGES, OUTBOUND_STAGES } from "../types/kanban";

interface KanbanBoardProps {
	type: "inbound" | "outbound";
}

type CardType = KanbanInbound | KanbanOutbound;

export function KanbanBoard({ type }: KanbanBoardProps) {
	const [cards, setCards] = useState<CardType[]>([]);
	const [loading, setLoading] = useState(false);

	const stages = type === "inbound" ? INBOUND_STAGES : OUTBOUND_STAGES;

	const fetchCards = useCallback(async () => {
		setLoading(true);
		try {
			const data =
				type === "inbound"
					? ((await api.listKanbanInbound()) as KanbanInbound[])
					: ((await api.listKanbanOutbound()) as KanbanOutbound[]);
			setCards(data);
		} catch (err) {
			console.error("Fetch kanban error:", err);
		} finally {
			setLoading(false);
		}
	}, [type]);

	useEffect(() => {
		fetchCards();
	}, [fetchCards]);

	const moveCard = async (cardId: number, toStage: string) => {
		try {
			if (type === "inbound") {
				await api.moveKanbanInbound(cardId, {
					to_stage: toStage,
					user_id: "admin",
				});
			} else {
				const result = (await api.moveKanbanOutbound(cardId, {
					to_stage: toStage,
					user_id: "admin",
				})) as { warning?: string; message?: string };
				if (result.warning === "NEGATIVE_STOCK") {
					alert(`Warning: ${result.message}`);
				}
			}
			await fetchCards();
		} catch (err) {
			alert(
				`Move failed: ${err instanceof Error ? err.message : "Unknown error"}`,
			);
		}
	};

	const getNextStage = (currentStage: string): string | null => {
		const stageKeys = stages.map((s) => s.key);
		const idx = stageKeys.indexOf(currentStage as InboundStage & OutboundStage);
		if (idx < 0 || idx >= stageKeys.length - 1) return null;
		return stageKeys[idx + 1];
	};

	return (
		<div style={{ padding: 16 }}>
			<h2>{type === "inbound" ? "Kanban Nhập hàng" : "Kanban Xuất hàng"}</h2>
			{loading && <p>Loading...</p>}

			<div style={{ display: "flex", gap: 16, overflowX: "auto" }}>
				{stages.map((stage) => {
					const stageCards = cards.filter((c) => c.stage === stage.key);
					return (
						<div
							key={stage.key}
							style={{
								minWidth: 280,
								background: "#f5f5f5",
								borderRadius: 8,
								padding: 12,
								flex: "1 0 280px",
							}}
						>
							<h3
								style={{
									margin: "0 0 12px 0",
									borderBottom: "2px solid #ddd",
									paddingBottom: 8,
								}}
							>
								{stage.label}
								<span style={{ color: "#888", fontSize: 14, marginLeft: 8 }}>
									({stageCards.length})
								</span>
							</h3>

							{stageCards.map((card) => {
								const nextStage = getNextStage(card.stage);
								return (
									<div
										key={card.id}
										style={{
											background: "#fff",
											borderRadius: 6,
											padding: 12,
											marginBottom: 8,
											boxShadow: "0 1px 3px rgba(0,0,0,0.1)",
										}}
									>
										<div style={{ fontWeight: 600, marginBottom: 4 }}>
											{card.ma_hang}
										</div>
										<div
											style={{ fontSize: 13, color: "#555", marginBottom: 4 }}
										>
											{card.ten_san_pham}
										</div>
										<div style={{ fontSize: 13, marginBottom: 8 }}>
											SL: {card.so_luong.toLocaleString("vi-VN")}
										</div>
										{card.note && (
											<div
												style={{ fontSize: 12, color: "#888", marginBottom: 8 }}
											>
												{card.note}
											</div>
										)}
										{nextStage && (
											<button
												onClick={() => moveCard(card.id, nextStage)}
												style={{
													background: "#4CAF50",
													color: "#fff",
													border: "none",
													borderRadius: 4,
													padding: "6px 12px",
													cursor: "pointer",
													fontSize: 13,
												}}
											>
												→ {stages.find((s) => s.key === nextStage)?.label}
											</button>
										)}
									</div>
								);
							})}
						</div>
					);
				})}
			</div>
		</div>
	);
}
