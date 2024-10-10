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

func (r *tokenTransferRepository) GetTokenTransferHistories(ctx context.Context, filters map[string]interface{}, page, size int) ([]model.TokenTransferHistory, error) {
	var tokenTransfers []model.TokenTransferHistory
	offset := (page - 1) * size

	query := r.db.WithContext(ctx).Limit(size + 1).Offset(offset) // Start with pagination setup

	// Apply filters only if they are provided (filters is not nil)
	if filters != nil {
		if transactionHash, ok := filters["transaction_hash"]; ok && transactionHash != "" {
			query = query.Where("transaction_hash = ?", transactionHash)
		}
		if fromAddress, ok := filters["from_address"]; ok && fromAddress != "" {
			query = query.Where("from_address = ?", fromAddress)
		}
		if toAddress, ok := filters["to_address"]; ok && toAddress != "" {
			query = query.Where("to_address = ?", toAddress)
		}
		if symbol, ok := filters["symbol"]; ok && symbol != "" {
			query = query.Where("symbol = ?", symbol)
		}
	}

	// Execute query
	if err := query.Find(&tokenTransfers).Error; err != nil {
		return nil, err
	}

	return tokenTransfers, nil
}
