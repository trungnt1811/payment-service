package types

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
)

type PaymentEventHistoryUCase interface {
	CreatePaymentEventHistory(ctx context.Context, payloads []dto.PaymentEventPayloadDTO) error
}
