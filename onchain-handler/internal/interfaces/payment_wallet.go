package interfaces

import (
	"context"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/dto"
)

type PaymentWalletRepository interface {
	CreateNewWallet(tx *gorm.DB, inUse bool) (*domain.PaymentWallet, error)
	IsRowExist(ctx context.Context) (bool, error)
	ClaimFirstAvailableWallet(tx *gorm.DB, ctx context.Context) (*domain.PaymentWallet, error)
	GetPaymentWalletByAddress(ctx context.Context, address string) (*domain.PaymentWallet, error)
	GetPaymentWallets(ctx context.Context) ([]domain.PaymentWallet, error)
	GetPaymentWalletsWithBalances(ctx context.Context, nonZeroOnly bool, network *string) ([]domain.PaymentWallet, error)
	ReleaseWalletsByIDs(ctx context.Context, walletIDs []uint64) error
}

type PaymentWalletUCase interface {
	CreateAndGenerateWallet(ctx context.Context, mnemonic, passphrase, salt string, inUse bool) error
	IsRowExist(ctx context.Context) (bool, error)
	GetPaymentWalletByAddress(ctx context.Context, address string) (dto.PaymentWalletDTO, error)
	UpsertPaymentWalletBalance(
		ctx context.Context,
		walletID uint64,
		newBalance string,
		network constants.NetworkType,
		symbol string,
	) error
	SubtractPaymentWalletBalance(
		ctx context.Context,
		walletID uint64,
		amountToSubtract string,
		network constants.NetworkType,
		symbol string,
	) error
	GetPaymentWallets(ctx context.Context) ([]dto.PaymentWalletDTO, error)
	GetPaymentWalletsWithBalances(ctx context.Context, nonZeroOnly bool, network *constants.NetworkType) ([]dto.PaymentWalletBalanceDTO, error)
	GetReceivingWalletAddress(
		ctx context.Context, mnemonic, passphrase, salt string,
	) (string, error)
}
