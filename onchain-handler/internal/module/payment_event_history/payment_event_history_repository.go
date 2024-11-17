package payment_event_history

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
)

type paymentEventHistoryRepository struct {
	db *gorm.DB
}

func NewPaymentEventHistoryRepository(db *gorm.DB) interfaces.PaymentEventHistoryRepository {
	return &paymentEventHistoryRepository{
		db: db,
	}
}

// CreatePaymentEventHistory inserts multiple payment event history records in a single transaction.
func (r *paymentEventHistoryRepository) CreatePaymentEventHistory(ctx context.Context, paymentEvents []domain.PaymentEventHistory) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&paymentEvents).Error; err != nil {
			return fmt.Errorf("failed to create payment event history records: %w", err)
		}
		return nil
	})
}
