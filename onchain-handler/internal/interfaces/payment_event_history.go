package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/domain/dto"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type PaymentEventHistoryRepository interface {
	CreatePaymentEventHistory(
		ctx context.Context,
		paymentEvents []entities.PaymentEventHistory,
	) ([]entities.PaymentEventHistory, error)
}

type PaymentEventHistoryUCase interface {
	CreatePaymentEventHistory(ctx context.Context, payloads []dto.PaymentEventPayloadDTO) error
}
