package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type PaymentWalletRepository interface {
	CreatePaymentWallets(ctx context.Context, models []model.PaymentWallet) error
	IsRowExist(ctx context.Context) (bool, error)
	ClaimFirstAvailableWallet(ctx context.Context) (*model.PaymentWallet, error)
}

type PaymentWalletUCase interface {
	CreatePaymentWallets(ctx context.Context, payloads []dto.PaymentWalletPayloadDTO) error
	IsRowExist(ctx context.Context) (bool, error)
}
