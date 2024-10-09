package model

import (
	"time"

	"github.com/genefriendway/onchain-handler/internal/dto"
)

type TokenTransferHistory struct {
	ID                       uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	TokenDistributionAddress string    `json:"token_distribution_address"`
	RecipientAddress         string    `json:"recipient_address"`
	TransactionHash          string    `json:"transaction_hash"`
	TokenAmount              string    `json:"token_amount"`
	Status                   int16     `json:"status"`
	ErrorMessage             string    `json:"error_message"`
	TxType                   string    `json:"tx_type"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

func (m *TokenTransferHistory) TableName() string {
	return "onchain_transaction"
}

func (m *TokenTransferHistory) ToDto() dto.TokenTransferHistoryDTO {
	return dto.TokenTransferHistoryDTO{
		ID:                       m.ID,
		TokenDistributionAddress: m.TokenDistributionAddress,
		RecipientAddress:         m.RecipientAddress,
		TransactionHash:          m.TransactionHash,
		TokenAmount:              m.TokenAmount,
		Status:                   m.Status,
		TxType:                   m.TxType,
		ErrorMessage:             m.ErrorMessage,
	}
}
