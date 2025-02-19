package ucases

import (
	"context"

	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
)

type paymentEventHistoryUCase struct {
	paymentEventHistoryRepository repotypes.PaymentEventHistoryRepository
}

func NewPaymentEventHistoryUCase(
	paymentEventHistoryRepository repotypes.PaymentEventHistoryRepository,
) ucasetypes.PaymentEventHistoryUCase {
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
