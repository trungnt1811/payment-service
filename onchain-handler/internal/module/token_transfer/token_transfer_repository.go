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

func (r *tokenTransferRepository) GetTokenTransferHistories(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]model.TokenTransferHistory, error) {
	var tokenTransfers []model.TokenTransferHistory

	// Start with pagination setup
	query := r.db.WithContext(ctx).Limit(limit).Offset(offset)

	// Apply filters only if they are provided (filters is not nil)
	if filters != nil {
		filterConditions := map[string]string{
			"transaction_hash": "transaction_hash = ?",
			"from_pool_name":   "from_pool_name = ?",
			"from_address":     "from_address = ?",
			"to_address":       "to_address = ?",
			"symbol":           "symbol = ?",
		}

		for key, condition := range filterConditions {
			if value, ok := filters[key]; ok && value != "" {
				query = query.Where(condition, value)
			}
		}
	}

	// Execute query
	if err := query.Find(&tokenTransfers).Error; err != nil {
		return nil, err
	}

	return tokenTransfers, nil
}
