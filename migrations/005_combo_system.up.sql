-- 005: Combo system (kho combo + phụ kiện)

-- Accessories (phụ kiện)
CREATE TABLE IF NOT EXISTS accessories (
    ma_phu_kien  VARCHAR(50) PRIMARY KEY,
    ten_phu_kien VARCHAR(200) NOT NULL,
    don_vi_tinh  VARCHAR(20) DEFAULT '',
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS accessory_inventory (
    ma_phu_kien VARCHAR(50) PRIMARY KEY REFERENCES accessories(ma_phu_kien),
    so_ton      NUMERIC(15,4) NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS accessory_movements (
    id             SERIAL PRIMARY KEY,
    ma_phu_kien    VARCHAR(50) NOT NULL REFERENCES accessories(ma_phu_kien),
    movement_type  VARCHAR(10) NOT NULL CHECK (movement_type IN ('IN', 'OUT', 'RETURN')),
    so_luong       NUMERIC(15,4) NOT NULL,
    note           TEXT DEFAULT '',
    created_at     TIMESTAMPTZ DEFAULT NOW()
);

-- Combo master (danh mục combo)
CREATE TABLE IF NOT EXISTS combo_master (
    ma_combo   VARCHAR(50) PRIMARY KEY,
    ten_combo  VARCHAR(200) NOT NULL,
    mo_ta      TEXT DEFAULT '',
    active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- BOM: bán thành phẩm (từ kho chính)
CREATE TABLE IF NOT EXISTS combo_bom_semi (
    id        SERIAL PRIMARY KEY,
    ma_combo  VARCHAR(50) NOT NULL REFERENCES combo_master(ma_combo) ON DELETE CASCADE,
    ma_hang   VARCHAR(50) NOT NULL REFERENCES products(ma_hang),
    so_luong  NUMERIC(15,4) NOT NULL CHECK (so_luong > 0),
    UNIQUE(ma_combo, ma_hang)
);

-- BOM: phụ kiện
CREATE TABLE IF NOT EXISTS combo_bom_accessory (
    id          SERIAL PRIMARY KEY,
    ma_combo    VARCHAR(50) NOT NULL REFERENCES combo_master(ma_combo) ON DELETE CASCADE,
    ma_phu_kien VARCHAR(50) NOT NULL REFERENCES accessories(ma_phu_kien),
    so_luong    NUMERIC(15,4) NOT NULL CHECK (so_luong > 0),
    UNIQUE(ma_combo, ma_phu_kien)
);

-- Tồn kho combo
CREATE TABLE IF NOT EXISTS combo_inventory (
    ma_combo   VARCHAR(50) PRIMARY KEY REFERENCES combo_master(ma_combo),
    so_ton     NUMERIC(15,4) NOT NULL DEFAULT 0,
    so_nhap    NUMERIC(15,4) NOT NULL DEFAULT 0,
    so_xuat    NUMERIC(15,4) NOT NULL DEFAULT 0,
    so_tra     NUMERIC(15,4) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Lịch sử nghiệp vụ combo
CREATE TABLE IF NOT EXISTS combo_transactions (
    id                SERIAL PRIMARY KEY,
    ma_combo          VARCHAR(50) NOT NULL REFERENCES combo_master(ma_combo),
    transaction_type  VARCHAR(10) NOT NULL CHECK (transaction_type IN ('CREATE', 'CANCEL', 'OUT', 'RETURN')),
    so_luong          NUMERIC(15,4) NOT NULL CHECK (so_luong > 0),
    note              TEXT DEFAULT '',
    created_at        TIMESTAMPTZ DEFAULT NOW()
);

-- Chi tiết tiêu thụ NVL per transaction
CREATE TABLE IF NOT EXISTS combo_component_movements (
    id                    SERIAL PRIMARY KEY,
    combo_transaction_id  INTEGER NOT NULL REFERENCES combo_transactions(id) ON DELETE CASCADE,
    component_type        VARCHAR(10) NOT NULL CHECK (component_type IN ('SEMI', 'ACCESSORY')),
    ma_component          VARCHAR(50) NOT NULL,
    so_luong              NUMERIC(15,4) NOT NULL,
    lot_id                INTEGER
);

-- Refresh inventory_grid view to include new tables if needed
-- (view unchanged — combo has its own inventory table)
