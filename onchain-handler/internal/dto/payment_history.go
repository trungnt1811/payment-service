package dto

import "time"

type PaymentHistoryDTO struct {
	TransactionHash string    `json:"transaction_hash"`
	FromAddress     string    `json:"from_address"`
	ToAddress       string    `json:"to_address"`
	Amount          string    `json:"amount"`
	TokenSymbol     string    `json:"token_symbol"`
	CreatedAt       time.Time `json:"created_at"`
}
