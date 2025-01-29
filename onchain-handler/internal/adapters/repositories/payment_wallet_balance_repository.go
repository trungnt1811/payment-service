package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

func (r *paymentWalletBalanceRepository) AddPaymentWalletBalance(
	ctx context.Context,
	walletID uint64,
	newBalance string,
	network string,
	symbol string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Prepare the wallet balance record
		walletBalance := domain.PaymentWalletBalance{
			WalletID: walletID,
			Network:  network,
			Symbol:   symbol,
			Balance:  newBalance,
		}

		// Use ON CONFLICT clause to perform upsert
		err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "wallet_id"}, {Name: "network"}, {Name: "symbol"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"balance": gorm.Expr("payment_wallet_balance.balance + ?", newBalance),
			}),
		}).Create(&walletBalance).Error
		if err != nil {
			return fmt.Errorf("failed to upsert wallet balance: %w", err)
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
		// Prepare the wallet balance record
		walletBalance := domain.PaymentWalletBalance{
			WalletID: walletID,
			Network:  network,
			Symbol:   symbol,
			Balance:  "0", // Default balance for new records
		}

		// Use ON CONFLICT clause to perform upsert with subtraction
		err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "wallet_id"}, {Name: "network"}, {Name: "symbol"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"balance": gorm.Expr("GREATEST(payment_wallet_balance.balance - ?, 0)", amountToSubtract),
			}),
		}).Create(&walletBalance).Error
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
		// Use ON CONFLICT to either insert or update the balance
		err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "wallet_id"}, {Name: "network"}, {Name: "symbol"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"balance": newBalance, // Directly set balance to new value
			}),
		}).Create(&domain.PaymentWalletBalance{
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
