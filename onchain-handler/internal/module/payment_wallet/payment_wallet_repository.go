package payment_wallet

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
	"github.com/genefriendway/onchain-handler/internal/utils"
)

type paymentWalletRepository struct {
	db     *gorm.DB
	config *conf.Configuration
}

func NewPaymentWalletRepository(db *gorm.DB, config *conf.Configuration) interfaces.PaymentWalletRepository {
	return &paymentWalletRepository{
		db:     db,
		config: config,
	}
}

func (r *paymentWalletRepository) CreatePaymentWallets(ctx context.Context, wallets []model.PaymentWallet) error {
	err := r.db.WithContext(ctx).Create(&wallets).Error
	if err != nil {
		return fmt.Errorf("failed to create payment wallets: %w", err)
	}
	return nil
}

func (r *paymentWalletRepository) IsRowExist(ctx context.Context) (bool, error) {
	var wallet model.PaymentWallet
	// Try to fetch the first row from the PaymentWallet table
	err := r.db.WithContext(ctx).Model(&model.PaymentWallet{}).First(&wallet).Error
	if err != nil {
		// If the error is "record not found", return false without error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		// Return error for any other failure
		return false, fmt.Errorf("failed to check rows existence: %w", err)
	}
	// If no error, it means a row exists
	return true, nil
}

// ClaimFirstAvailableWallet fetches the first wallet where InUse is false, and updates it to be in use
func (r *paymentWalletRepository) ClaimFirstAvailableWallet(ctx context.Context) (*model.PaymentWallet, error) {
	var wallet model.PaymentWallet

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find the first available wallet and lock it
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("in_use = ?", false).
			Order("id").
			First(&wallet).Error

		if err == nil {
			// If wallet is found, mark it as in use
			return tx.Model(&wallet).Update("in_use", true).Error
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to find available wallet: %w", err)
		}

		// If no available wallet found, generate a new wallet
		wallets, genErr := utils.GenerateWallets(r.config, 1)
		if genErr != nil {
			return fmt.Errorf("failed to generate new wallet: %w", genErr)
		}

		// Mark the first newly created wallet as in use
		wallets[0].InUse = true

		// Insert the new wallet into the database
		createErr := tx.Create(&wallets[0]).Error
		if createErr != nil {
			return fmt.Errorf("failed to create new wallet: %w", createErr)
		}

		wallet = wallets[0]
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to claim available wallet: %w", err)
	}

	return &wallet, nil
}
