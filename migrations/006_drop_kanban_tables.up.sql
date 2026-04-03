-- Drop Kanban feature tables (no longer used)
DROP TABLE IF EXISTS kanban_events;
DROP TABLE IF EXISTS kanban_inbound;
DROP TABLE IF EXISTS kanban_outbound;

DROP INDEX IF EXISTS idx_kanban_events_sku;
DROP INDEX IF EXISTS idx_kanban_inbound_stage;
DROP INDEX IF EXISTS idx_kanban_outbound_stage;
