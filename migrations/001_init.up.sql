CREATE TABLE IF NOT EXISTS products (
    ma_hang TEXT PRIMARY KEY,
    ten_san_pham TEXT NOT NULL DEFAULT '',
    ma_bu TEXT NOT NULL DEFAULT '',
    ma_cat TEXT NOT NULL DEFAULT '',
    ma_nhom_hang TEXT NOT NULL DEFAULT '',
    nhom_hang TEXT NOT NULL DEFAULT '',
    don_vi_tinh TEXT NOT NULL DEFAULT '',
    quy_cach TEXT NOT NULL DEFAULT '',
    don_gia NUMERIC(15,2) NOT NULL DEFAULT 0,
    vat NUMERIC(5,2) NOT NULL DEFAULT 0,
    gia_niv NUMERIC(15,2) NOT NULL DEFAULT 0,
    gia_nhap NUMERIC(15,2) NOT NULL DEFAULT 0,
    ngay_cap_nhat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    hoa_hong NUMERIC(5,2) NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS inventory_main (
    ma_hang TEXT PRIMARY KEY REFERENCES products(ma_hang),
    ten_san_pham TEXT NOT NULL DEFAULT '',
    so_ton NUMERIC(15,2) NOT NULL DEFAULT 0,
    so_nhap NUMERIC(15,2) NOT NULL DEFAULT 0,
    so_xuat NUMERIC(15,2) NOT NULL DEFAULT 0,
    tien_ton NUMERIC(15,2) NOT NULL DEFAULT 0,
    tien_nhap NUMERIC(15,2) NOT NULL DEFAULT 0,
    tien_xuat NUMERIC(15,2) NOT NULL DEFAULT 0,
    so_ngay_ton NUMERIC(10,2) NOT NULL DEFAULT 0,
    luong_ban_binh_quan_ngay NUMERIC(15,4) NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS inbound_items (
    id BIGSERIAL PRIMARY KEY,
    ma_hang TEXT NOT NULL REFERENCES products(ma_hang),
    ten_san_pham TEXT NOT NULL DEFAULT '',
    don_vi_tinh TEXT NOT NULL DEFAULT '',
    quy_cach TEXT NOT NULL DEFAULT '',
    so_luong NUMERIC(15,2) NOT NULL DEFAULT 0,
    doanh_so NUMERIC(15,2) NOT NULL DEFAULT 0,
    chiet_khau NUMERIC(15,2) NOT NULL DEFAULT 0,
    so_luong_tra_lai NUMERIC(15,2) NOT NULL DEFAULT 0,
    doanh_thu NUMERIC(15,2) NOT NULL DEFAULT 0,
    von NUMERIC(15,2) NOT NULL DEFAULT 0,
    lai_gop NUMERIC(15,2) NOT NULL DEFAULT 0,
    ti_le_lai_gop NUMERIC(5,2) NOT NULL DEFAULT 0,
    ngay_nhan_hang TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS outbound_items (
    id BIGSERIAL PRIMARY KEY,
    ma_hang TEXT NOT NULL REFERENCES products(ma_hang),
    ten_san_pham TEXT NOT NULL DEFAULT '',
    don_vi_tinh TEXT NOT NULL DEFAULT '',
    quy_cach TEXT NOT NULL DEFAULT '',
    so_luong NUMERIC(15,2) NOT NULL DEFAULT 0,
    doanh_so NUMERIC(15,2) NOT NULL DEFAULT 0,
    chiet_khau NUMERIC(15,2) NOT NULL DEFAULT 0,
    so_luong_tra_lai NUMERIC(15,2) NOT NULL DEFAULT 0,
    doanh_thu NUMERIC(15,2) NOT NULL DEFAULT 0,
    von NUMERIC(15,2) NOT NULL DEFAULT 0,
    lai_gop NUMERIC(15,2) NOT NULL DEFAULT 0,
    ti_le_lai_gop NUMERIC(5,2) NOT NULL DEFAULT 0,
    ngay_nhan_hang TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inventory_movements (
    movement_id BIGSERIAL PRIMARY KEY,
    ma_hang TEXT NOT NULL REFERENCES products(ma_hang),
    qty NUMERIC(15,2) NOT NULL DEFAULT 0,
    movement_type TEXT NOT NULL CHECK (movement_type IN ('IN','OUT')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kanban_events (
    event_id BIGSERIAL PRIMARY KEY,
    sku TEXT NOT NULL,
    from_stage TEXT NOT NULL DEFAULT '',
    to_stage TEXT NOT NULL DEFAULT '',
    user_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rule_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT ''
);
INSERT INTO rule_config (key, value) VALUES ('optimal_days', '14') ON CONFLICT DO NOTHING;
INSERT INTO rule_config (key, value) VALUES ('gap_ratio', '0.10') ON CONFLICT DO NOTHING;

-- Import batch tracking
CREATE TABLE IF NOT EXISTS import_batches (
    batch_id BIGSERIAL PRIMARY KEY,
    file_type TEXT NOT NULL,
    file_name TEXT NOT NULL DEFAULT '',
    total_rows INT NOT NULL DEFAULT 0,
    success_rows INT NOT NULL DEFAULT 0,
    error_rows INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    errors JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- Kanban cards for inbound
CREATE TABLE IF NOT EXISTS kanban_inbound (
    id BIGSERIAL PRIMARY KEY,
    ma_hang TEXT NOT NULL REFERENCES products(ma_hang),
    ten_san_pham TEXT NOT NULL DEFAULT '',
    so_luong NUMERIC(15,2) NOT NULL DEFAULT 0,
    stage TEXT NOT NULL DEFAULT 'can_nhap' CHECK (stage IN ('can_nhap','da_len_don','da_duyet','da_ve_hang')),
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Kanban cards for outbound
CREATE TABLE IF NOT EXISTS kanban_outbound (
    id BIGSERIAL PRIMARY KEY,
    ma_hang TEXT NOT NULL REFERENCES products(ma_hang),
    ten_san_pham TEXT NOT NULL DEFAULT '',
    so_luong NUMERIC(15,2) NOT NULL DEFAULT 0,
    stage TEXT NOT NULL DEFAULT 'can_day' CHECK (stage IN ('can_day','da_chot_don','da_giao')),
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Async jobs tracking
CREATE TABLE IF NOT EXISTS async_jobs (
    job_id TEXT PRIMARY KEY,
    job_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','running','completed','failed')),
    payload JSONB DEFAULT '{}',
    result JSONB DEFAULT '{}',
    error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inventory_main_ma_hang ON inventory_main(ma_hang);
CREATE INDEX idx_inbound_items_ma_hang ON inbound_items(ma_hang);
CREATE INDEX idx_inbound_items_ngay ON inbound_items(ngay_nhan_hang);
CREATE INDEX idx_outbound_items_ma_hang ON outbound_items(ma_hang);
CREATE INDEX idx_outbound_items_ngay ON outbound_items(ngay_nhan_hang);
CREATE INDEX idx_movements_ma_hang ON inventory_movements(ma_hang);
CREATE INDEX idx_movements_created ON inventory_movements(created_at);
CREATE INDEX idx_kanban_events_sku ON kanban_events(sku);
CREATE INDEX idx_kanban_inbound_stage ON kanban_inbound(stage);
CREATE INDEX idx_kanban_outbound_stage ON kanban_outbound(stage);
CREATE INDEX idx_async_jobs_status ON async_jobs(status);
