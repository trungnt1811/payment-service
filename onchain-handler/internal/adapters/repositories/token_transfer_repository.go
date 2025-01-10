package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
)

type tokenTransferRepository struct {
	db *gorm.DB
}

func NewTokenTransferRepository(db *gorm.DB) interfaces.TokenTransferRepository {
	return &tokenTransferRepository{
		db: db,
	}
}

func (r *tokenTransferRepository) CreateTokenTransferHistories(ctx context.Context, models []domain.TokenTransferHistory) error {
	err := r.db.WithContext(ctx).Create(&models).Error
	if err != nil {
		return fmt.Errorf("failed to create transfer histories: %w", err)
	}
	return nil
}

func (r *tokenTransferRepository) GetTokenTransferHistories(
	ctx context.Context,
	limit, offset int,
	orderBy *string, // Order by field
	orderDirection constants.OrderDirection,
	startTime, endTime *time.Time, // Range of time to filter by
	fromAddress, toAddress *string, // Address filters
) ([]domain.TokenTransferHistory, error) {
	var tokenTransfers []domain.TokenTransferHistory

	orderColumn := "id" // Default values for ordering
	if orderBy != nil && *orderBy != "" {
		orderColumn = *orderBy
	}

	orderDir := string(constants.Asc) // Default direction
	if orderDirection == constants.Desc {
		orderDir = string(constants.Desc)
	}

	// Start with pagination setup
	query := r.db.WithContext(ctx).
		Where("symbol = ?", constants.USDT).
		Limit(limit).
		Offset(offset).
		Order(fmt.Sprintf("%s %s", orderColumn, orderDir))

	// Apply time range filter if both startTime and endTime are provided
	if startTime != nil && endTime != nil {
		query = query.Where("created_at BETWEEN ? AND ?", startTime, endTime)
	}

	// Apply from_address filter if provided
	if fromAddress != nil && *fromAddress != "" {
		query = query.Where("from_address = ?", *fromAddress)
	}

	// Apply to_address filter if provided
	if toAddress != nil && *toAddress != "" {
		query = query.Where("to_address = ?", *toAddress)
	}

	// Execute query
	if err := query.Find(&tokenTransfers).Error; err != nil {
		return nil, err
	}

	return tokenTransfers, nil
}
