package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/dto"
)

type PaymentWalletRepository interface {
	CreatePaymentWallets(ctx context.Context, models []domain.PaymentWallet) error
	IsRowExist(ctx context.Context) (bool, error)
	ClaimFirstAvailableWallet(ctx context.Context) (*domain.PaymentWallet, error)
}

type PaymentWalletUCase interface {
	CreatePaymentWallets(ctx context.Context, payloads []dto.PaymentWalletPayloadDTO) error
	IsRowExist(ctx context.Context) (bool, error)
}
