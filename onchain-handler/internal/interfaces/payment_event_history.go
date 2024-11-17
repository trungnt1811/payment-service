package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/dto"
)

type PaymentEventHistoryRepository interface {
	CreatePaymentEventHistory(ctx context.Context, paymentEvents []domain.PaymentEventHistory) error
}

type PaymentEventHistoryUCase interface {
	CreatePaymentEventHistory(ctx context.Context, payloads []dto.PaymentEventPayloadDTO) error
}
