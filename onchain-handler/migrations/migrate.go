package migrations

import (
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

func RunMigrations(db *gorm.DB) error {
	log.Println("Running migrations...")

	sqlScripts := []string{
		"migrations/sql/01_onchain_token_transfer.sql",
		"migrations/sql/02_block_state.sql",
		"migrations/sql/03_payment_wallet.sql",
		"migrations/sql/04_payment_order.sql",
		"migrations/sql/05_payment_event_history.sql",
		"migrations/sql/06_user_wallet.sql",
		"migrations/sql/07_payment_wallet_balance.sql",
		"migrations/sql/08_blockchain_network_metadata.sql",
		"migrations/sql/09_add_upcoming_block_height_to_payment_order.sql",
		"migrations/sql/10_add_index_for_from_address_in_payment_event_history.sql",
		"migrations/sql/11_addition_indexes_for_payment_order.sql",
		"migrations/sql/12_add_vendor_id_to_payment_order.sql",
	}

	for _, script := range sqlScripts {
		if err := applySQLScript(db, script); err != nil {
			return err
		}
	}

	log.Println("Migrations completed.")
	return nil
}
