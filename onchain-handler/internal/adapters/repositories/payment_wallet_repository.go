package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/google/uuid"

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
			inUse := true // New wallets are always in use
			newWallet, createErr := r.CreateNewWallet(tx.WithContext(ctx), inUse)
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

func (r *paymentWalletRepository) CreateNewWallet(tx *gorm.DB, inUse bool) (*domain.PaymentWallet, error) {
	// Generate a unique, temporary address
	uuidPart := strings.ReplaceAll(uuid.New().String(), "-", "") // Remove dashes from UUID
	tempAddress := "temp-" + uuidPart[:min(len(uuidPart), 37)]   // Ensure total length <= 42

	// Insert a placeholder wallet with the unique temporary address
	placeholderWallet := domain.PaymentWallet{
		InUse:   inUse,
		Address: tempAddress, // Unique temporary address within 42 characters
	}

	if err := tx.Create(&placeholderWallet).Error; err != nil {
		return nil, fmt.Errorf("failed to insert new wallet placeholder: %w", err)
	}

	// The placeholderWallet.ID now contains the auto-incremented ID from the DB.

	// Generate the real wallet account using the assigned ID.
	account, _, genErr := crypto.GenerateAccount(
		r.config.Wallet.Mnemonic,
		r.config.Wallet.Passphrase,
		r.config.Wallet.Salt,
		constants.PaymentWallet,
		placeholderWallet.ID,
	)
	if genErr != nil {
		return nil, fmt.Errorf("failed to generate new wallet: %w", genErr)
	}

	// Update the wallet record with the real address
	if err := tx.Model(&placeholderWallet).Update("address", account.Address.Hex()).Error; err != nil {
		return nil, fmt.Errorf("failed to update wallet address: %w", err)
	}

	// Return the wallet with its final state
	placeholderWallet.Address = account.Address.Hex()
	return &placeholderWallet, nil
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
