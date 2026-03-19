DROP TABLE IF EXISTS inventory_thresholds;
DROP TABLE IF EXISTS inventory_lots;
ALTER TABLE outbound_items DROP COLUMN IF EXISTS batch_code;
ALTER TABLE inbound_items DROP COLUMN IF EXISTS batch_code;
