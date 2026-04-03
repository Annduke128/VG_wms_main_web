-- Recreate Kanban tables (rollback)
CREATE TABLE IF NOT EXISTS kanban_events (
    event_id BIGSERIAL PRIMARY KEY,
    sku TEXT NOT NULL,
    from_stage TEXT NOT NULL DEFAULT '',
    to_stage TEXT NOT NULL DEFAULT '',
    user_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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

CREATE INDEX idx_kanban_events_sku ON kanban_events(sku);
CREATE INDEX idx_kanban_inbound_stage ON kanban_inbound(stage);
CREATE INDEX idx_kanban_outbound_stage ON kanban_outbound(stage);
