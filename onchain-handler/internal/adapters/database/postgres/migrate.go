package postgres

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

func applySQLScript(db *gorm.DB, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := db.Exec(string(content)).Error; err != nil {
		return err
	}

	log.Printf("Successfully applied migration: %s", filepath.Base(filePath))
	return nil
}

func RunMigrations(db *gorm.DB, basePath string) error {
	log.Println("Running migrations...")

	// List of migration SQL files
	scriptFiles := []string{
		"01_onchain_token_transfer.sql",
		"02_block_state.sql",
		"03_payment_wallet.sql",
		"04_payment_order.sql",
		"05_payment_event_history.sql",
		"06_user_wallet.sql",
		"07_payment_wallet_balance.sql",
		"08_blockchain_network_metadata.sql",
		"09_add_upcoming_block_height_to_payment_order.sql",
		"10_add_index_for_from_address_in_payment_event_history.sql",
		"11_addition_indexes_for_payment_order.sql",
		"12_add_vendor_id_to_payment_order.sql",
		"13_payment_statistics.sql",
		"14_update_transfer_type.sql",
	}

	// Iterate over scripts and execute each
	for _, script := range scriptFiles {
		scriptPath := filepath.Join(basePath, script)

		log.Printf("Applying migration: %s\n", scriptPath)
		if err := applySQLScript(db, scriptPath); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", scriptPath, err)
		}
	}

	log.Println("Migrations completed successfully.")
	return nil
}
