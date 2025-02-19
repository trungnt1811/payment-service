package entities

import (
	"time"

	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
)

type PaymentWallet struct {
	ID                    uint64                 `json:"id" gorm:"primaryKey;autoIncrement"`
	Address               string                 `json:"address"`
	InUse                 bool                   `json:"in_use"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
	PaymentWalletBalances []PaymentWalletBalance `json:"payment_wallet_balance" gorm:"foreignKey:WalletID"`
}

func (m *PaymentWallet) TableName() string {
	return "payment_wallet"
}

func (m *PaymentWallet) ToDto() dto.PaymentWalletDTO {
	return dto.PaymentWalletDTO{
		ID:      m.ID,
		Address: m.Address,
		InUse:   m.InUse,
	}
}
