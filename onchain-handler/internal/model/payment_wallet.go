package model

import (
	"time"

	"github.com/genefriendway/onchain-handler/internal/dto"
)

type PaymentWallet struct {
	ID         uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Address    string    `json:"address"`
	PrivateKey string    `json:"private_key"`
	InUse      bool      `json:"in_use"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (m *PaymentWallet) TableName() string {
	return "payment_wallet"
}

func (m *PaymentWallet) ToDto() dto.PaymentWalletDTO {
	return dto.PaymentWalletDTO{
		ID:         m.ID,
		Address:    m.Address,
		PrivateKey: m.PrivateKey,
		InUse:      m.InUse,
	}
}
