package payment_event_history

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type paymentEventHistoryUCase struct {
	paymentEventHistoryRepository interfaces.PaymentEventHistoryRepository
}

func NewPaymentEventHistoryUCase(
	paymentEventHistoryRepository interfaces.PaymentEventHistoryRepository,
) interfaces.PaymentEventHistoryUCase {
	return &paymentEventHistoryUCase{
		paymentEventHistoryRepository: paymentEventHistoryRepository,
	}
}

func (u *paymentEventHistoryUCase) CreatePaymentEventHistory(
	ctx context.Context,
	payloads []dto.PaymentEventPayloadDTO,
) error {
	var eventHistories []model.PaymentEventHistory
	for _, payload := range payloads {
		eventHistory := model.PaymentEventHistory{
			PaymentOrderID:  payload.PaymentOrderID,
			TransactionHash: payload.TransactionHash,
			FromAddress:     payload.FromAddress,
			ToAddress:       payload.ToAddress,
			ContractAddress: payload.ContractAddress,
			TokenSymbol:     payload.TokenSymbol,
			Amount:          payload.Amount,
			Network:         payload.Network,
		}
		eventHistories = append(eventHistories, eventHistory)
	}
	return u.paymentEventHistoryRepository.CreatePaymentEventHistory(ctx, eventHistories)
}
