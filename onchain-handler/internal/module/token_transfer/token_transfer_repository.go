package token_transfer

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type tokenTransferRepository struct {
	db *gorm.DB
}

func NewTokenTransferRepository(db *gorm.DB) interfaces.TokenTransferRepository {
	return &tokenTransferRepository{
		db: db,
	}
}

func (r *tokenTransferRepository) CreateTokenTransferHistories(ctx context.Context, models []model.TokenTransferHistory) error {
	err := r.db.WithContext(ctx).Create(&models).Error
	if err != nil {
		return fmt.Errorf("failed to create transfer histories: %w", err)
	}
	return nil
}

func (r *tokenTransferRepository) GetTokenTransferHistories(ctx context.Context, page, size int) ([]model.TokenTransferHistory, error) {
	var tokenTransfers []model.TokenTransferHistory
	offset := (page - 1) * size

	// Query to fetch token transfer histories with pagination
	if err := r.db.WithContext(ctx).
		Limit(size + 1). // Fetch one more than the size to check if there's a next page
		Offset(offset).
		Find(&tokenTransfers).Error; err != nil {
		return nil, err
	}

	return tokenTransfers, nil
}
