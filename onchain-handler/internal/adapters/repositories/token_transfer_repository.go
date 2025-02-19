package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/constants"
	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type tokenTransferRepository struct {
	db *gorm.DB
}

func NewTokenTransferRepository(db *gorm.DB) repotypes.TokenTransferRepository {
	return &tokenTransferRepository{
		db: db,
	}
}

func (r *tokenTransferRepository) CreateTokenTransferHistories(ctx context.Context, models []entities.TokenTransferHistory) error {
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
) ([]entities.TokenTransferHistory, error) {
	var tokenTransfers []entities.TokenTransferHistory

	orderColumn := "id" // Default values for ordering
	if orderBy != nil && *orderBy != "" {
		orderColumn = *orderBy
	}

	orderDir := constants.Asc.String() // Default direction
	if orderDirection == constants.Desc {
		orderDir = constants.Desc.String()
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

func (r *tokenTransferRepository) GetTotalTokenAmount(
	ctx context.Context,
	startTime, endTime *time.Time,
	fromAddress, toAddress *string,
) (float64, error) {
	var totalTokenAmount float64

	// Build the base query for summing the token_amount
	query := r.db.WithContext(ctx).
		Table("onchain_token_transfer").
		Where("symbol = ?", constants.USDT)

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

	// Ensure SUM never returns NULL
	if err := query.Select("COALESCE(SUM(token_amount), 0)").Scan(&totalTokenAmount).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate total token amount: %w", err)
	}

	return totalTokenAmount, nil
}
