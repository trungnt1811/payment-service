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
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/crypto"
)

type paymentWalletRepository struct {
	db *gorm.DB
}

func NewPaymentWalletRepository(db *gorm.DB) interfaces.PaymentWalletRepository {
	return &paymentWalletRepository{
		db: db,
	}
}

func (r *paymentWalletRepository) IsRowExist(ctx context.Context) (bool, error) {
	var wallet entities.PaymentWallet
	// Try to fetch the first row from the PaymentWallet table
	err := r.db.WithContext(ctx).Model(&entities.PaymentWallet{}).First(&wallet).Error
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
func (r *paymentWalletRepository) ClaimFirstAvailableWallet(tx *gorm.DB, ctx context.Context) (*entities.PaymentWallet, error) {
	var wallet entities.PaymentWallet

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

func (r *paymentWalletRepository) CreateNewWallet(tx *gorm.DB, inUse bool) (*entities.PaymentWallet, error) {
	// Generate a unique, temporary address
	uuidPart := strings.ReplaceAll(uuid.New().String(), "-", "") // Remove dashes from UUID
	tempAddress := "temp-" + uuidPart[:min(len(uuidPart), 37)]   // Ensure total length <= 42

	// Insert a placeholder wallet with the unique temporary address
	placeholderWallet := entities.PaymentWallet{
		InUse:   inUse,
		Address: tempAddress, // Unique temporary address within 42 characters
	}

	if err := tx.Create(&placeholderWallet).Error; err != nil {
		return nil, fmt.Errorf("failed to insert new wallet placeholder: %w", err)
	}

	// The placeholderWallet.ID now contains the auto-incremented ID from the DB.

	// Generate the real wallet account using the assigned ID.
	walletConfig := conf.GetWalletConfiguration()
	account, _, genErr := crypto.GenerateAccount(
		walletConfig.Mnemonic,
		walletConfig.Passphrase,
		walletConfig.Salt,
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

func (r *paymentWalletRepository) GetPaymentWalletByAddress(ctx context.Context, address string) (*entities.PaymentWallet, error) {
	var wallet entities.PaymentWallet
	if err := r.db.WithContext(ctx).Where("address = ?", address).First(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *paymentWalletRepository) GetPaymentWallets(ctx context.Context) ([]entities.PaymentWallet, error) {
	var wallets []entities.PaymentWallet
	if err := r.db.WithContext(ctx).Find(&wallets).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch payment wallets: %w", err)
	}
	return wallets, nil
}

func (r *paymentWalletRepository) GetPaymentWalletsWithBalances(
	ctx context.Context,
	limit, offset int,
	network *string,
) ([]entities.PaymentWallet, error) {
	var wallets []entities.PaymentWallet

	query := r.db.WithContext(ctx).
		Select("payment_wallet.*").
		Joins("JOIN payment_wallet_balance ON payment_wallet.id = payment_wallet_balance.wallet_id").
		Where("payment_wallet_balance.symbol = ?", constants.USDT). // USDT filter
		Group("payment_wallet.id").
		Having("SUM(payment_wallet_balance.balance) > 0"). // Ensure only wallets with USDT appear
		Order("payment_wallet.id ASC")

	// Apply optional network filtering
	if network != nil {
		query = query.Where("payment_wallet_balance.network = ?", *network)
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// Execute query and Preload balances (ensures no null values)
	if err := query.Preload("PaymentWalletBalances", func(db *gorm.DB) *gorm.DB {
		// Preload only USDT balances that are > 0
		return db.Where("payment_wallet_balance.symbol = ?", constants.USDT).
			Where("payment_wallet_balance.balance > 0")
	}).Find(&wallets).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch payment wallets with balances: %w", err)
	}

	return wallets, nil
}

func (r *paymentWalletRepository) GetPaymentWalletWithBalancesByAddress(
	ctx context.Context, address *string,
) (entities.PaymentWallet, error) {
	var wallet entities.PaymentWallet

	query := r.db.WithContext(ctx).
		Select("payment_wallet.*").
		Joins("LEFT JOIN payment_wallet_balance ON payment_wallet.id = payment_wallet_balance.wallet_id AND payment_wallet_balance.symbol = ?", constants.USDT). // Use LEFT JOIN to include zero balances
		Group("payment_wallet.id").
		Order("payment_wallet.id ASC")

	// Apply required address filtering (should not be nil)
	if address != nil {
		query = query.Where("payment_wallet.address = ?", *address)
	} else {
		return entities.PaymentWallet{}, fmt.Errorf("address is required")
	}

	// Execute query and Preload balances (ensures no null values)
	err := query.Preload("PaymentWalletBalances", func(db *gorm.DB) *gorm.DB {
		// Preload all USDT balances (even if 0)
		return db.Where("payment_wallet_balance.symbol = ?", constants.USDT)
	}).First(&wallet).Error // Use `First` instead of `Find`
	// Handle case when wallet is not found
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entities.PaymentWallet{}, fmt.Errorf("wallet not found: %w", err)
		}
		return entities.PaymentWallet{}, fmt.Errorf("failed to fetch payment wallet with balances: %w", err)
	}

	return wallet, nil
}

func (r *paymentWalletRepository) GetTotalBalancePerNetwork(ctx context.Context, network *string) (map[string]string, error) {
	var totalBalances []struct {
		Network      string
		TotalBalance string
	}

	// Start query to sum only USDT balances per network
	query := r.db.WithContext(ctx).
		Table("payment_wallet_balance").
		Select("network, COALESCE(SUM(balance), '0') as total_balance"). //	Ensures zero if no balance exists
		Where("symbol = ?", constants.USDT).                             // Filter by USDT only
		Group("network")

	// Apply optional network filter
	if network != nil {
		query = query.Where("network = ?", *network)
	}

	// Execute query
	if err := query.Scan(&totalBalances).Error; err != nil {
		return nil, fmt.Errorf("failed to compute total balance per network: %w", err)
	}

	// Convert to map
	totalBalanceMap := make(map[string]string)
	for _, row := range totalBalances {
		totalBalanceMap[row.Network] = row.TotalBalance
	}

	return totalBalanceMap, nil
}

func (r *paymentWalletRepository) ReleaseWalletsByIDs(ctx context.Context, walletIDs []uint64) error {
	if len(walletIDs) == 0 {
		return nil // No IDs provided, nothing to release
	}

	err := r.db.WithContext(ctx).Model(&entities.PaymentWallet{}).
		Where("id IN ?", walletIDs).
		Update("in_use", false).
		Error
	if err != nil {
		return fmt.Errorf("failed to release wallets: %w", err)
	}
	return nil
}

func (r *paymentWalletRepository) GetWalletIDByAddress(ctx context.Context, address string) (uint64, error) {
	var walletID uint64
	err := r.db.WithContext(ctx).
		Model(&entities.PaymentWallet{}).
		Select("id").
		Where("address = ?", address).
		Limit(1).
		Scan(&walletID).
		Error
	if err != nil {
		return 0, fmt.Errorf("failed to get wallet ID: %w", err)
	}

	// If walletID remains 0, it means no wallet was found, return an explicit error
	if walletID == 0 {
		return 0, fmt.Errorf("wallet not found for address: %s", address)
	}

	return walletID, nil
}
