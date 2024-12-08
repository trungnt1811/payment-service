package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
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

func (r *paymentWalletRepository) CreatePaymentWallets(ctx context.Context, wallets []domain.PaymentWallet) error {
	err := r.db.WithContext(ctx).Create(&wallets).Error
	if err != nil {
		return fmt.Errorf("failed to create payment wallets: %w", err)
	}
	return nil
}

func (r *paymentWalletRepository) IsRowExist(ctx context.Context) (bool, error) {
	var wallet domain.PaymentWallet
	// Try to fetch the first row from the PaymentWallet table
	err := r.db.WithContext(ctx).Model(&domain.PaymentWallet{}).First(&wallet).Error
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

// ClaimFirstAvailableWallet attempts to find the first available wallet and mark it as in-use.
// If no available wallet is found, it creates a new one.
func (r *paymentWalletRepository) ClaimFirstAvailableWallet(tx *gorm.DB, ctx context.Context) (*domain.PaymentWallet, error) {
	var wallet domain.PaymentWallet

	// Attempt to find and lock the first available wallet
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("in_use = ?", false).
		Order("id").
		First(&wallet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No available wallet found; create a new one
			newWallet, createErr := r.createNewWallet(tx.WithContext(ctx))
			if createErr != nil {
				return nil, fmt.Errorf("failed to create new wallet: %w", createErr)
			}
			return newWallet, nil
		}
		// Other errors
		return nil, fmt.Errorf("failed to find available wallet: %w", err)
	}

	// Mark the found wallet as in-use in a separate query
	if updateErr := tx.Model(&wallet).Update("in_use", true).Error; updateErr != nil {
		return nil, fmt.Errorf("failed to mark wallet as in use: %w", updateErr)
	}

	// If we reach here, we have successfully locked and updated an existing wallet
	return &wallet, nil
}

func (r *paymentWalletRepository) createNewWallet(tx *gorm.DB) (*domain.PaymentWallet, error) {
	var maxID uint64
	if err := tx.Model(&domain.PaymentWallet{}).Select("MAX(id)").Scan(&maxID).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve max wallet ID: %w", err)
	}
	nextID := maxID + 1

	// Generate a new wallet account
	account, _, genErr := crypto.GenerateAccount(
		r.config.Wallet.Mnemonic,
		r.config.Wallet.Passphrase,
		r.config.Wallet.Salt,
		constants.PaymentWallet,
		nextID,
	)
	if genErr != nil {
		return nil, fmt.Errorf("failed to generate new wallet: %w", genErr)
	}

	// Create and persist the wallet
	newWallet := domain.PaymentWallet{
		ID:      nextID,
		Address: account.Address.Hex(),
		InUse:   true,
	}
	if err := tx.Create(&newWallet).Error; err != nil {
		return nil, fmt.Errorf("failed to insert new wallet: %w", err)
	}

	return &newWallet, nil
}

func (r *paymentWalletRepository) GetPaymentWalletByAddress(ctx context.Context, address string) (*domain.PaymentWallet, error) {
	var wallet domain.PaymentWallet
	if err := r.db.WithContext(ctx).Where("address = ?", address).First(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *paymentWalletRepository) GetPaymentWallets(ctx context.Context) ([]domain.PaymentWallet, error) {
	var wallets []domain.PaymentWallet
	if err := r.db.WithContext(ctx).Find(&wallets).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch payment wallets: %w", err)
	}
	return wallets, nil
}

func (r *paymentWalletRepository) GetPaymentWalletsWithBalances(ctx context.Context, nonZeroOnly bool, network *string) ([]domain.PaymentWallet, error) {
	var wallets []domain.PaymentWallet

	// Build the base query
	query := r.db.WithContext(ctx).
		Order("id ASC").
		Preload("PaymentWalletBalances", func(db *gorm.DB) *gorm.DB {
			// Apply filters to the balances
			if nonZeroOnly {
				db = db.Where("balance > ?", "0")
			}
			if network != nil {
				db = db.Where("network = ?", *network)
			}
			return db
		})

	// Apply join and filtering to ensure wallets with no balances are excluded
	query = query.Joins("JOIN payment_wallet_balance ON payment_wallet.id = payment_wallet_balance.wallet_id").
		Group("payment_wallet.id")

	if nonZeroOnly {
		query = query.Having("SUM(CASE WHEN payment_wallet_balance.balance > 0 THEN 1 ELSE 0 END) > 0")
	}

	if network != nil {
		query = query.Where("payment_wallet_balance.network = ?", *network)
	}

	// Execute the query
	if err := query.Find(&wallets).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch payment wallets with balances: %w", err)
	}

	return wallets, nil
}

func (r *paymentWalletRepository) ReleaseWalletsByIDs(ctx context.Context, walletIDs []uint64) error {
	if len(walletIDs) == 0 {
		return nil // No IDs provided, nothing to release
	}

	err := r.db.WithContext(ctx).Model(&domain.PaymentWallet{}).
		Where("id IN ?", walletIDs).
		Update("in_use", false).
		Error
	if err != nil {
		return fmt.Errorf("failed to release wallets: %w", err)
	}
	return nil
}
