export interface InventoryMain {
	ma_hang: string;
	ten_san_pham: string;
	so_ton: number;
	so_nhap: number;
	so_xuat: number;
	tien_ton: number;
	tien_nhap: number;
	tien_xuat: number;
	so_ngay_ton: number;
	luong_ban_binh_quan_ngay: number;
	so_ngay_ton_ban: number;
	// Joined from products table via inventory_grid view
	don_gia: number;
	ma_bu: string;
	ma_nhom_hang: string;
}
