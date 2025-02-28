package types

import (
	"context"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
)

type PaymentWalletUCase interface {
	CreateAndGenerateWallet(ctx context.Context, mnemonic, passphrase, salt string, inUse bool) error
	IsRowExist(ctx context.Context) (bool, error)
	GetPaymentWalletByAddress(ctx context.Context, address string) (dto.PaymentWalletBalanceDTO, error)
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
	GetReceivingWalletAddressWithBalances(
		ctx context.Context, mnemonic, passphrase, salt string,
	) (string, map[constants.NetworkType]string, error)
	SyncWalletBalance(ctx context.Context, walletAddress string, network constants.NetworkType) (string, error)
}
