package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type paymentEventHistoryRepository struct {
	db *gorm.DB
}

func NewPaymentEventHistoryRepository(db *gorm.DB) repotypes.PaymentEventHistoryRepository {
	return &paymentEventHistoryRepository{
		db: db,
	}
}

// CreatePaymentEventHistory inserts multiple payment event history records in a single transaction
// and returns the created records.
func (r *paymentEventHistoryRepository) CreatePaymentEventHistory(
	ctx context.Context,
	paymentEvents []entities.PaymentEventHistory,
) ([]entities.PaymentEventHistory, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&paymentEvents).Error; err != nil {
			return fmt.Errorf("failed to create payment event history records: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Return the created models with updated fields (e.g., IDs, timestamps)
	return paymentEvents, nil
}
