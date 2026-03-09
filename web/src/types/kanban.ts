export type InboundStage =
	| "can_nhap"
	| "da_len_don"
	| "da_duyet"
	| "da_ve_hang";
export type OutboundStage = "can_day" | "da_chot_don" | "da_giao";

export interface KanbanInbound {
	id: number;
	ma_hang: string;
	ten_san_pham: string;
	so_luong: number;
	stage: InboundStage;
	note: string;
	created_at: string;
	updated_at: string;
}

export interface KanbanOutbound {
	id: number;
	ma_hang: string;
	ten_san_pham: string;
	so_luong: number;
	stage: OutboundStage;
	note: string;
	created_at: string;
	updated_at: string;
}

export interface CreateKanbanRequest {
	ma_hang: string;
	ten_san_pham: string;
	so_luong: number;
	note: string;
}

export interface MoveKanbanRequest {
	to_stage: string;
	user_id: string;
}

export const INBOUND_STAGES: { key: InboundStage; label: string }[] = [
	{ key: "can_nhap", label: "Cần nhập" },
	{ key: "da_len_don", label: "Đã lên đơn" },
	{ key: "da_duyet", label: "Đã duyệt" },
	{ key: "da_ve_hang", label: "Đã về hàng" },
];

export const OUTBOUND_STAGES: { key: OutboundStage; label: string }[] = [
	{ key: "can_day", label: "Cần đẩy" },
	{ key: "da_chot_don", label: "Đã chốt đơn" },
	{ key: "da_giao", label: "Đã giao" },
];
