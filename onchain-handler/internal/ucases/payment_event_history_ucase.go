package ucases

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/domain/dto"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
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
	var eventHistories []entities.PaymentEventHistory
	for _, payload := range payloads {
		eventHistory := entities.PaymentEventHistory{
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
