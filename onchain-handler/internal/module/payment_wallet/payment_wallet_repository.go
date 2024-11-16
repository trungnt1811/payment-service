package payment_wallet

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
	"github.com/genefriendway/onchain-handler/pkg/crypto"
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

// ClaimFirstAvailableWallet fetches the first available wallet or creates a new one if none are available
func (r *paymentWalletRepository) ClaimFirstAvailableWallet(ctx context.Context) (*model.PaymentWallet, error) {
	var wallet model.PaymentWallet

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Attempt to find and lock the first available wallet
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("in_use = ?", false).
			Order("id").
			First(&wallet).Error; err == nil {
			// Wallet found, mark as in use
			return tx.Model(&wallet).Update("in_use", true).Error
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			// An error occurred other than 'record not found'
			return fmt.Errorf("failed to find available wallet: %w", err)
		}

		// If no wallet found, count current records to derive next ID
		var recordCount int64
		if err := tx.Model(&model.PaymentWallet{}).Count(&recordCount).Error; err != nil {
			return fmt.Errorf("failed to count wallets: %w", err)
		}
		nextID := uint64(recordCount + 1)

		// Generate a new wallet account
		account, _, genErr := crypto.GenerateAccount(
			r.config.Wallet.Mnemonic,
			r.config.Wallet.Passphrase,
			r.config.Wallet.Salt,
			constants.PaymentWallet,
			nextID,
		)
		if genErr != nil {
			return fmt.Errorf("failed to generate new wallet: %w", genErr)
		}

		// Initialize the new wallet struct
		wallet = model.PaymentWallet{
			ID:      nextID,
			Address: account.Address.Hex(),
			InUse:   true,
		}

		// Insert the new wallet into the database
		if err := tx.Create(&wallet).Error; err != nil {
			return fmt.Errorf("failed to create new wallet: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to claim available wallet: %w", err)
	}

	return &wallet, nil
}
