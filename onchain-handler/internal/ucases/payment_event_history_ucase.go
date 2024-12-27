package ucases

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
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
	var eventHistories []domain.PaymentEventHistory
	for _, payload := range payloads {
		eventHistory := domain.PaymentEventHistory{
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
	_, err := u.paymentEventHistoryRepository.CreatePaymentEventHistory(ctx, eventHistories)
	return err
}
