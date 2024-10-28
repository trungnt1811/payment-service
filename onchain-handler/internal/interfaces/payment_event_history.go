package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type PaymentEventHistoryRepository interface {
	CreatePaymentEventHistory(ctx context.Context, paymentEvents []model.PaymentEventHistory) error
}

type PaymentEventHistoryUCase interface {
	CreatePaymentEventHistory(ctx context.Context, payloads []dto.PaymentEventPayloadDTO) error
}
