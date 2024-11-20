package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

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
	requestIDs []string, // List of request IDs to filter
	startTime, endTime time.Time, // Range of time to filter by
) ([]domain.TokenTransferHistory, error) {
	var tokenTransfers []domain.TokenTransferHistory

	// Start with pagination setup
	query := r.db.WithContext(ctx).Limit(limit).Offset(offset).Order("id ASC")

	// Apply filter for request IDs if provided
	if len(requestIDs) > 0 {
		query = query.Where("request_id IN ?", requestIDs)
	}

	// Apply time range filter if both startTime and endTime are provided
	if !startTime.IsZero() && !endTime.IsZero() {
		query = query.Where("created_at BETWEEN ? AND ?", startTime, endTime)
	}

	// Execute query
	if err := query.Find(&tokenTransfers).Error; err != nil {
		return nil, err
	}

	return tokenTransfers, nil
}
