// --- Combo Master ---

export interface ComboMaster {
	ma_combo: string;
	ten_combo: string;
	mo_ta: string;
	active: boolean;
	created_at: string;
	updated_at: string;
}

export interface ComboBOMSemi {
	id: number;
	ma_combo: string;
	ma_hang: string;
	so_luong: number;
	ten_san_pham?: string;
}

export interface ComboBOMAccessory {
	id: number;
	ma_combo: string;
	ma_phu_kien: string;
	so_luong: number;
	ten_phu_kien?: string;
}

export interface ComboDetail extends ComboMaster {
	bom_semi: ComboBOMSemi[];
	bom_accessory: ComboBOMAccessory[];
}

// --- Combo Inventory ---

export interface ComboInventory {
	ma_combo: string;
	so_ton: number;
	so_nhap: number;
	so_xuat: number;
	so_tra: number;
	updated_at: string;
	ten_combo?: string;
}

// --- Combo Transaction ---

export interface ComboTransaction {
	id: number;
	ma_combo: string;
	transaction_type: "CREATE" | "CANCEL" | "OUT" | "RETURN";
	so_luong: number;
	note: string;
	created_at: string;
	ten_combo?: string;
}

// --- Accessories ---

export interface Accessory {
	ma_phu_kien: string;
	ten_phu_kien: string;
	don_vi_tinh: string;
	created_at: string;
}

export interface AccessoryInventory {
	ma_phu_kien: string;
	so_ton: number;
	updated_at: string;
	ten_phu_kien?: string;
	don_vi_tinh?: string;
}
