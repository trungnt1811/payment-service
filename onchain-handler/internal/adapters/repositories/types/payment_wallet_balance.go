package interfaces

import "context"

type PaymentWalletBalanceRepository interface {
	AddPaymentWalletBalance(
		ctx context.Context,
		walletID uint64,
		amountToAdd string,
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
	UpsertPaymentWalletBalance(
		ctx context.Context,
		walletID uint64,
		newBalance string,
		network string,
		symbol string,
	) error
}
