package types

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type PaymentEventHistoryRepository interface {
	CreatePaymentEventHistory(
		ctx context.Context,
		paymentEvents []entities.PaymentEventHistory,
	) ([]entities.PaymentEventHistory, error)
}
