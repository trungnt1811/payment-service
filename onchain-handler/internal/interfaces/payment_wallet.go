package interfaces

import (
	"context"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/dto"
)

type PaymentWalletRepository interface {
	CreatePaymentWallets(ctx context.Context, models []domain.PaymentWallet) error
	IsRowExist(ctx context.Context) (bool, error)
	ClaimFirstAvailableWallet(ctx context.Context) (*domain.PaymentWallet, error)
	GetPaymentWalletByAddress(ctx context.Context, address string) (*domain.PaymentWallet, error)
	GetPaymentWallets(ctx context.Context) ([]domain.PaymentWallet, error)
	GetPaymentWalletsWithBalances(ctx context.Context, nonZeroOnly bool, network *string) ([]domain.PaymentWallet, error)
	BatchReleaseWallets(ctx context.Context, walletIDs []uint64) error
}

type PaymentWalletUCase interface {
	CreatePaymentWallets(ctx context.Context, payloads []dto.PaymentWalletPayloadDTO) error
	IsRowExist(ctx context.Context) (bool, error)
	GetPaymentWalletByAddress(ctx context.Context, address string) (dto.PaymentWalletDTO, error)
	UpsertPaymentWalletBalances(
		ctx context.Context,
		walletIDs []uint64,
		newBalances []string,
		network constants.NetworkType,
		symbol string,
	) error
	GetPaymentWallets(ctx context.Context) ([]dto.PaymentWalletDTO, error)
	GetPaymentWalletsWithBalances(ctx context.Context, nonZeroOnly bool, network *constants.NetworkType) ([]dto.PaymentWalletBalanceDTO, error)
}
