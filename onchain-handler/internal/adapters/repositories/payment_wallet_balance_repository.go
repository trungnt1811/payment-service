package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
)

type paymentWalletBalanceRepository struct {
	db *gorm.DB
}

func NewPaymentWalletBalanceRepository(db *gorm.DB) interfaces.PaymentWalletBalanceRepository {
	return &paymentWalletBalanceRepository{
		db: db,
	}
}

// UpsertPaymentWalletBalances updates or inserts the balances of multiple wallets by their WalletIDs.
func (r *paymentWalletBalanceRepository) UpsertPaymentWalletBalances(
	ctx context.Context,
	walletIDs []uint64,
	newBalances []string,
	network string,
	symbol string,
) error {
	// Check if the lengths of walletIDs and newBalances match
	if len(walletIDs) != len(newBalances) {
		return fmt.Errorf("the number of wallet IDs and balances must be the same")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Loop through each wallet and either insert or update
		for i, walletID := range walletIDs {
			var existingCount int64
			// Step 1: Check if the record exists
			if err := tx.Table("payment_wallet_balance").
				Where("wallet_id = ? AND network = ? AND symbol = ?", walletID, network, symbol).
				Count(&existingCount).Error; err != nil {
				return fmt.Errorf("failed to check for existing record: %w", err)
			}

			if existingCount > 0 {
				// Step 2: If record exists, update it
				if err := tx.Model(&domain.PaymentWalletBalance{}).
					Where("wallet_id = ? AND network = ? AND symbol = ?", walletID, network, symbol).
					Updates(map[string]interface{}{
						"balance": newBalances[i],
					}).Error; err != nil {
					return fmt.Errorf("failed to update wallet balance: %w", err)
				}
			} else {
				// Step 3: If record does not exist, insert it
				if err := tx.Create(&domain.PaymentWalletBalance{
					WalletID:  walletID,
					Network:   network,
					Symbol:    symbol,
					Balance:   newBalances[i],
					UpdatedAt: time.Now(),
				}).Error; err != nil {
					return fmt.Errorf("failed to insert wallet balance: %w", err)
				}
			}
		}

		return nil
	})
}
