package model

import (
	"time"

	"github.com/genefriendway/onchain-handler/internal/dto"
)

type TokenTransferHistory struct {
	ID              uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	RequestID       string    `json:"request_id"`
	TransactionHash string    `json:"transaction_hash"`
	FromAddress     string    `json:"from_address"`
	ToAddress       string    `json:"to_address"`
	TokenAmount     uint64    `json:"token_amount"`
	Symbol          string    `json:"symbol"`
	Status          bool      `json:"status"`
	ErrorMessage    string    `json:"error_message"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (m *TokenTransferHistory) TableName() string {
	return "onchain_token_transfer"
}

func (m *TokenTransferHistory) ToDto() dto.TokenTransferHistoryDTO {
	return dto.TokenTransferHistoryDTO{
		ID:              m.ID,
		RequestID:       m.RequestID,
		TransactionHash: m.TransactionHash,
		FromAddress:     m.FromAddress,
		ToAddress:       m.ToAddress,
		TokenAmount:     m.TokenAmount,
		Symbol:          m.Symbol,
		Status:          m.Status,
		ErrorMessage:    m.ErrorMessage,
	}
}
