-- Add batch_code to inbound/outbound items
ALTER TABLE inbound_items ADD COLUMN IF NOT EXISTS batch_code TEXT NOT NULL DEFAULT '';
ALTER TABLE outbound_items ADD COLUMN IF NOT EXISTS batch_code TEXT NOT NULL DEFAULT '';

-- FIFO lot tracking
CREATE TABLE IF NOT EXISTS inventory_lots (
    id BIGSERIAL PRIMARY KEY,
    ma_hang TEXT NOT NULL REFERENCES products(ma_hang),
    batch_code TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    qty_in NUMERIC(15,2) NOT NULL DEFAULT 0,
    qty_out NUMERIC(15,2) NOT NULL DEFAULT 0,
    qty_remaining NUMERIC(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_lots_ma_hang_batch ON inventory_lots(ma_hang, batch_code);
CREATE INDEX idx_lots_ma_hang ON inventory_lots(ma_hang);
CREATE INDEX idx_lots_batch ON inventory_lots(batch_code);
CREATE INDEX idx_lots_received ON inventory_lots(received_at);
CREATE INDEX idx_lots_remaining ON inventory_lots(ma_hang, qty_remaining) WHERE qty_remaining > 0;

-- Inventory thresholds with version history
CREATE TABLE IF NOT EXISTS inventory_thresholds (
    id BIGSERIAL PRIMARY KEY,
    ma_hang TEXT NOT NULL REFERENCES products(ma_hang),
    min_qty NUMERIC(15,2) NOT NULL DEFAULT 0,
    optimal_qty NUMERIC(15,2) NOT NULL DEFAULT 0,
    max_age_days INT NOT NULL DEFAULT 0,
    source TEXT NOT NULL DEFAULT 'manual',
    model_version TEXT NOT NULL DEFAULT '',
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_to TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_thresholds_ma_hang ON inventory_thresholds(ma_hang);
CREATE INDEX idx_thresholds_effective ON inventory_thresholds(ma_hang, effective_from, effective_to);
