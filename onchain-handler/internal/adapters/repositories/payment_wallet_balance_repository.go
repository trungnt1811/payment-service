package repositories

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/genefriendway/onchain-handler/internal/domain/entities"
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

func (r *paymentWalletBalanceRepository) AddPaymentWalletBalance(
	ctx context.Context,
	walletID uint64,
	amountToAdd string,
	network string,
	symbol string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock the existing balance row
		var existingBalance string
		err := tx.Raw(`
			SELECT balance 
			FROM payment_wallet_balance 
			WHERE wallet_id = ? AND network = ? AND symbol = ? 
			FOR UPDATE
		`, walletID, network, symbol).Scan(&existingBalance).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to lock balance row: %w", err)
		}

		// Convert existing balance and amount to add
		existingBalanceFloat, _ := new(big.Float).SetString(existingBalance)
		newBalanceFloat, _ := new(big.Float).SetString(amountToAdd)

		// Calculate updated balance
		updatedBalance := new(big.Float).Add(existingBalanceFloat, newBalanceFloat)

		// Use ON CONFLICT to handle insert/update logic safely
		err = tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "wallet_id"}, {Name: "network"}, {Name: "symbol"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"balance": updatedBalance.String(),
			}),
		}).Create(&entities.PaymentWalletBalance{
			WalletID: walletID,
			Network:  network,
			Symbol:   symbol,
			Balance:  updatedBalance.String(),
		}).Error
		if err != nil {
			return fmt.Errorf("failed to add wallet balance: %w", err)
		}

		return nil
	})
}

func (r *paymentWalletBalanceRepository) SubtractPaymentWalletBalance(
	ctx context.Context,
	walletID uint64,
	amountToSubtract string,
	network string,
	symbol string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock the existing balance row
		var existingBalance string
		err := tx.Raw(`
			SELECT balance 
			FROM payment_wallet_balance 
			WHERE wallet_id = ? AND network = ? AND symbol = ? 
			FOR UPDATE
		`, walletID, network, symbol).Scan(&existingBalance).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to lock balance row: %w", err)
		}

		// Convert existing balance and amount to subtract
		existingBalanceFloat, _ := new(big.Float).SetString(existingBalance)
		amountToSubtractFloat, _ := new(big.Float).SetString(amountToSubtract)

		// Calculate new balance (ensuring it does not go negative)
		newBalance := new(big.Float).Sub(existingBalanceFloat, amountToSubtractFloat)
		if newBalance.Cmp(big.NewFloat(0)) < 0 {
			newBalance.SetFloat64(0)
		}

		// Use ON CONFLICT to handle insert/update logic safely
		err = tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "wallet_id"}, {Name: "network"}, {Name: "symbol"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"balance": newBalance.String(),
			}),
		}).Create(&entities.PaymentWalletBalance{
			WalletID: walletID,
			Network:  network,
			Symbol:   symbol,
			Balance:  newBalance.String(),
		}).Error
		if err != nil {
			return fmt.Errorf("failed to subtract wallet balance: %w", err)
		}

		return nil
	})
}

func (r *paymentWalletBalanceRepository) UpsertPaymentWalletBalance(
	ctx context.Context,
	walletID uint64,
	newBalance string,
	network string,
	symbol string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock row to prevent race conditions
		err := tx.Exec(`
			SELECT 1 FROM payment_wallet_balance 
			WHERE wallet_id = ? AND network = ? AND symbol = ? 
			FOR UPDATE
		`, walletID, network, symbol).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to lock balance row: %w", err)
		}

		// Overwrite balance using `ON CONFLICT`
		err = tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "wallet_id"}, {Name: "network"}, {Name: "symbol"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"balance": newBalance, // Directly overwrites balance
			}),
		}).Create(&entities.PaymentWalletBalance{
			WalletID: walletID,
			Network:  network,
			Symbol:   symbol,
			Balance:  newBalance,
		}).Error
		if err != nil {
			return fmt.Errorf("failed to upsert wallet balance: %w", err)
		}

		return nil
	})
}
