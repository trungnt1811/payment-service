package dto

import "github.com/genefriendway/onchain-handler/constants"

type TokenTransferPayloadDTO struct {
	Network     string `json:"network"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	TokenAmount string `json:"token_amount"`
	RequestID   string `json:"request_id"`
	Symbol      string `json:"symbol"`
}

type PaymentOrderPayloadDTO struct {
	RequestID  string `json:"request_id"`
	Amount     string `json:"amount"`
	Symbol     string `json:"symbol"`
	Network    string `json:"network"`
	WebhookURL string `json:"webhook_url"`
}

type PaymentOrderNetworkPayloadDTO struct {
	RequestID string `json:"request_id" binding:"required"`
	Network   string `json:"network" binding:"required"`
}

type PaymentOrderSymbolPayloadDTO struct {
	RequestID string `json:"request_id" binding:"required"`
	Symbol    string `json:"symbol" binding:"required"`
}

type PaymentEventPayloadDTO struct {
	PaymentOrderID  uint64 `json:"payment_order_id"`
	Network         string `json:"network"`
	TransactionHash string `json:"transaction_hash"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	ContractAddress string `json:"contract_address"`
	TokenSymbol     string `json:"token_symbol"`
	Amount          string `json:"amount"`
}

type PaymentWalletPayloadDTO struct {
	ID      uint64 `json:"id"`
	Address string `json:"address"`
	InUse   bool   `json:"in_use"`
}

type UserWalletPayloadDTO struct {
	UserID  uint64 `json:"user_id"`
	Address string `json:"address"`
}

type SyncWalletBalancePayloadDTO struct {
	WalletAddress string                `json:"wallet_address" binding:"required"`
	Network       constants.NetworkType `json:"network" binding:"required"`
}

type UpdatePaymentOrderPayloadDTO struct {
	Network string `json:"network,omitempty"`
	Symbol  string `json:"symbol,omitempty"`
}
