package user_wallet

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type userWalletRepository struct {
	db *gorm.DB
}

func NewUserWalletRepository(db *gorm.DB) interfaces.UserWalletRepository {
	return &userWalletRepository{
		db: db,
	}
}

func (r *userWalletRepository) CreateUserWallets(ctx context.Context, userWallets []model.UserWallet) error {
	err := r.db.WithContext(ctx).Create(&userWallets).Error
	if err != nil {
		return fmt.Errorf("failed to create user wallets: %w", err)
	}
	return nil
}

// GetUserWallets retrieves user wallets with pagination and optional user ID filtering.
func (r *userWalletRepository) GetUserWallets(
	ctx context.Context,
	limit, offset int,
	userIDs []uint64,
) ([]model.UserWallet, error) {
	var wallets []model.UserWallet

	// Start the query with pagination
	query := r.db.WithContext(ctx).Limit(limit).Offset(offset)

	// Apply filter for user IDs if provided
	if len(userIDs) > 0 {
		query = query.Where("user_id IN ?", userIDs)
	}

	// Execute query
	if err := query.Find(&wallets).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve user wallets: %w", err)
	}

	return wallets, nil
}
