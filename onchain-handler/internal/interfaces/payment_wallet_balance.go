package interfaces

import "context"

type PaymentWalletBalanceRepository interface {
	UpsertPaymentWalletBalances(
		ctx context.Context,
		walletIDs []uint64,
		newBalances []string,
		network string,
		symbol string,
	) error
}
