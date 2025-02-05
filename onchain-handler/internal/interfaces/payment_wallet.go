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
	GetPaymentWalletsWithBalances(
		ctx context.Context,
		limit, offset int,
		network *string,
	) ([]domain.PaymentWallet, error)
	GetPaymentWalletWithBalancesByAddress(
		ctx context.Context,
		network, address *string,
	) (domain.PaymentWallet, error)
	GetTotalBalancePerNetwork(ctx context.Context, network *string) (map[string]string, error)
	ReleaseWalletsByIDs(ctx context.Context, walletIDs []uint64) error
	GetWalletIDByAddress(ctx context.Context, address string) (uint64, error)
}

type PaymentWalletUCase interface {
	CreateAndGenerateWallet(ctx context.Context, mnemonic, passphrase, salt string, inUse bool) error
	IsRowExist(ctx context.Context) (bool, error)
	GetPaymentWalletByAddress(ctx context.Context, network *constants.NetworkType, address string) (dto.PaymentWalletBalanceDTO, error)
	AddPaymentWalletBalance(
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
	GetPaymentWalletsWithBalances(
		ctx context.Context, network *constants.NetworkType,
	) ([]dto.PaymentWalletBalanceDTO, error)
	GetPaymentWalletsWithBalancesPagination(
		ctx context.Context, page, size int, network *constants.NetworkType,
	) (dto.PaginationDTOResponse, error)
	GetReceivingWalletAddress(
		ctx context.Context, mnemonic, passphrase, salt string,
	) (string, error)
	SyncWalletBalance(ctx context.Context, walletAddress string, network constants.NetworkType) (string, error)
}
