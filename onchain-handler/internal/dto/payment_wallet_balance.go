package dto

type TokenBalanceDTO struct {
	Symbol string `json:"symbol"`
	Amount string `json:"amount"`
}

type NetworkBalanceDTO struct {
	Network       string            `json:"network"`
	TokenBalances []TokenBalanceDTO `json:"token_balances"`
}

type PaymentWalletBalanceDTO struct {
	ID              uint64              `json:"id"`
	Address         string              `json:"address"`
	NetworkBalances []NetworkBalanceDTO `json:"network_balances"`
}
