package repo

import (
	"context"
	"fmt"
)

// ResetAllData truncates all business data tables, preserving schema and migrations.
func (r *PostgresRepo) ResetAllData(ctx context.Context) error {
	// Order matters: truncate tables with FK dependencies first, or use CASCADE.
	// Using a single TRUNCATE with CASCADE for safety.
	_, err := r.Pool.Exec(ctx, `
		TRUNCATE TABLE
			combo_component_movements,
			combo_transactions,
			combo_inventory,
			combo_bom_accessory,
			combo_bom_semi,
			combo_master,
			accessory_movements,
			accessory_inventory,
			accessories,
			inventory_lbbq_history,
			inventory_movements,
			inventory_lots,
			inbound_items,
			outbound_items,
			kanban_events,
			kanban_inbound,
			kanban_outbound,
			inventory_main,
			inventory_thresholds,
			import_batches,
			async_jobs,
			products
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		return fmt.Errorf("reset all data: %w", err)
	}
	return nil
}
