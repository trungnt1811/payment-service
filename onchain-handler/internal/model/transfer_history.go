package model

import (
	"time"

	"github.com/genefriendway/onchain-handler/internal/dto"
)

type TransferHistory struct {
	ID               uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	RewardAddress    string    `json:"reward_address"`
	RecipientAddress string    `json:"recipient_address"`
	TransactionHash  string    `json:"transaction_hash"`
	TokenAmount      string    `json:"token_amount"`
	Status           int16     `json:"status"`
	ErrorMessage     string    `json:"error_message"`
	TxType           string    `json:"tx_type"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (m *TransferHistory) TableName() string {
	return "onchain_transactions"
}

func (m *TransferHistory) ToDto() dto.TransferHistoryDTO {
	return dto.TransferHistoryDTO{
		ID:               m.ID,
		RewardAddress:    m.RewardAddress,
		RecipientAddress: m.RecipientAddress,
		TransactionHash:  m.TransactionHash,
		TokenAmount:      m.TokenAmount,
		Status:           m.Status,
		TxType:           m.TxType,
		ErrorMessage:     m.ErrorMessage,
	}
}
