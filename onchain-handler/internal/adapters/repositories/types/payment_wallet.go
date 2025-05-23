package types

import (
	"context"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type PaymentWalletRepository interface {
	CreateNewWallet(tx *gorm.DB, inUse bool) (*entities.PaymentWallet, error)
	IsRowExist(ctx context.Context) (bool, error)
	ClaimFirstAvailableWallet(tx *gorm.DB, ctx context.Context) (*entities.PaymentWallet, error)
	GetPaymentWalletByAddress(ctx context.Context, address string) (*entities.PaymentWallet, error)
	GetPaymentWallets(ctx context.Context) ([]entities.PaymentWallet, error)
	GetPaymentWalletsWithBalances(
		ctx context.Context,
		limit, offset int,
		network *string,
		symbols []string,
	) ([]entities.PaymentWallet, error)
	GetPaymentWalletWithBalancesByAddress(
		ctx context.Context, address *string,
	) (entities.PaymentWallet, error)
	GetTotalBalancePerNetwork(
		ctx context.Context,
		network *string,
		symbols []string,
	) (map[string]map[string]string, error)
	ReleaseWalletsByIDs(tx *gorm.DB, walletIDs []uint64) error
	GetWalletIDByAddress(ctx context.Context, address string) (uint64, error)
}
