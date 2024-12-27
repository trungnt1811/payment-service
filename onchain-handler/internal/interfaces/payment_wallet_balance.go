package interfaces

import "context"

type PaymentWalletBalanceRepository interface {
	UpsertPaymentWalletBalance(
		ctx context.Context,
		walletID uint64,
		newBalance string,
		network string,
		symbol string,
	) error
	SubtractPaymentWalletBalance(
		ctx context.Context,
		walletID uint64,
		amountToSubtract string,
		network string,
		symbol string,
	) error
}
