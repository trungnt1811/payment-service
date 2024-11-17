package ucases

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
)

type paymentWalletUCase struct {
	paymentWalletRepository interfaces.PaymentWalletRepository
}

func NewPaymentWalletUCase(
	paymentWalletRepository interfaces.PaymentWalletRepository,
) interfaces.PaymentWalletUCase {
	return &paymentWalletUCase{
		paymentWalletRepository: paymentWalletRepository,
	}
}

func (u *paymentWalletUCase) CreatePaymentWallets(ctx context.Context, payloads []dto.PaymentWalletPayloadDTO) error {
	var wallets []domain.PaymentWallet
	for _, payload := range payloads {
		wallet := domain.PaymentWallet{
			ID:      payload.ID,
			Address: payload.Address,
			InUse:   payload.InUse,
		}
		wallets = append(wallets, wallet)
	}
	return u.paymentWalletRepository.CreatePaymentWallets(ctx, wallets)
}

func (u *paymentWalletUCase) IsRowExist(ctx context.Context) (bool, error) {
	return u.paymentWalletRepository.IsRowExist(ctx)
}
