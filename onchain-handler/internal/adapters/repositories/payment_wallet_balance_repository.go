package repositories

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"gorm.io/gorm"

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
		// Lock the existing balance row, default balance to "0"
		existingBalance := "0"
		err := tx.Raw(`
			SELECT COALESCE(balance, '0') 
			FROM payment_wallet_balance 
			WHERE wallet_id = ? AND network = ? AND symbol = ? 
			FOR UPDATE
		`, walletID, network, symbol).Scan(&existingBalance).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to lock balance row: %w", err)
		}

		// Safely initialize big.Float values
		existingBalanceFloat := new(big.Float)
		if existingBalance != "" {
			if _, ok := existingBalanceFloat.SetString(existingBalance); !ok {
				return fmt.Errorf("invalid existing balance format: %s", existingBalance)
			}
		}

		newBalanceFloat := new(big.Float)
		if amountToAdd != "" {
			if _, ok := newBalanceFloat.SetString(amountToAdd); !ok {
				return fmt.Errorf("invalid amountToAdd format: %s", amountToAdd)
			}
		}

		// Calculate updated balance
		updatedBalance := new(big.Float).Add(existingBalanceFloat, newBalanceFloat)

		// Check if the balance row exists
		var count int64
		err = tx.Model(&entities.PaymentWalletBalance{}).
			Where("wallet_id = ? AND network = ? AND symbol = ?", walletID, network, symbol).
			Count(&count).Error
		if err != nil {
			return fmt.Errorf("failed to check existing balance: %w", err)
		}

		// If record exists, update it; otherwise, insert a new one
		if count > 0 {
			err = tx.Model(&entities.PaymentWalletBalance{}).
				Where("wallet_id = ? AND network = ? AND symbol = ?", walletID, network, symbol).
				Update("balance", updatedBalance.String()).Error
		} else {
			err = tx.Create(&entities.PaymentWalletBalance{
				WalletID: walletID,
				Network:  network,
				Symbol:   symbol,
				Balance:  updatedBalance.String(),
			}).Error
		}

		if err != nil {
			return fmt.Errorf("failed to update wallet balance: %w", err)
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
		// Lock the existing balance row, default balance to "0"
		existingBalance := "0"
		err := tx.Raw(`
			SELECT COALESCE(balance, '0') 
			FROM payment_wallet_balance 
			WHERE wallet_id = ? AND network = ? AND symbol = ? 
			FOR UPDATE
		`, walletID, network, symbol).Scan(&existingBalance).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to lock balance row: %w", err)
		}

		// Safely initialize big.Float values
		existingBalanceFloat := new(big.Float)
		if existingBalance != "" {
			existingBalanceFloat.SetString(existingBalance)
		} else {
			existingBalanceFloat.SetFloat64(0) // Default to 0
		}

		amountToSubtractFloat := new(big.Float)
		if amountToSubtract != "" {
			amountToSubtractFloat.SetString(amountToSubtract)
		} else {
			return fmt.Errorf("invalid amountToSubtract: %s", amountToSubtract)
		}

		// Ensure sufficient funds before subtracting
		if existingBalanceFloat.Cmp(amountToSubtractFloat) < 0 {
			return fmt.Errorf("insufficient funds in wallet %d (%s %s)", walletID, network, symbol)
		}

		// Calculate new balance
		newBalance := new(big.Float).Sub(existingBalanceFloat, amountToSubtractFloat)

		// Update the balance in the database
		err = tx.Model(&entities.PaymentWalletBalance{}).
			Where("wallet_id = ? AND network = ? AND symbol = ?", walletID, network, symbol).
			Update("balance", newBalance.String()).Error
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
		var exists int
		err := tx.Raw(`
			SELECT 1 FROM payment_wallet_balance 
			WHERE wallet_id = ? AND network = ? AND symbol = ? 
			FOR UPDATE
		`, walletID, network, symbol).Scan(&exists).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to lock balance row: %w", err)
		}

		// Upsert balance safely using `Save()`
		err = tx.Save(&entities.PaymentWalletBalance{
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
