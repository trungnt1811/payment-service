package domain

import (
	"time"

	"github.com/genefriendway/onchain-handler/internal/dto"
)

type TokenTransferHistory struct {
	ID              uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	RequestID       string    `json:"request_id"`
	Network         string    `json:"network"`
	TransactionHash string    `json:"transaction_hash"`
	FromAddress     string    `json:"from_address"`
	ToAddress       string    `json:"to_address"`
	TokenAmount     string    `json:"token_amount"`
	Fee             string    `json:"fee"`
	Symbol          string    `json:"symbol"`
	Status          bool      `json:"status"`
	Type            string    `json:"type"`
	ErrorMessage    string    `json:"error_message"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (m *TokenTransferHistory) TableName() string {
	return "onchain_token_transfer"
}

func (m *TokenTransferHistory) ToDto() dto.TokenTransferHistoryDTO {
	return dto.TokenTransferHistoryDTO{
		Network:         m.Network,
		TransactionHash: m.TransactionHash,
		FromAddress:     m.FromAddress,
		ToAddress:       m.ToAddress,
		TokenAmount:     m.TokenAmount,
		Fee:             m.Fee,
		Symbol:          m.Symbol,
		Status:          m.Status,
		Type:            m.Type,
		ErrorMessage:    m.ErrorMessage,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}
