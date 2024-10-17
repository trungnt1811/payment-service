package model

import (
	"time"

	"github.com/genefriendway/onchain-handler/internal/dto"
)

type PaymentOrder struct {
	ID          uint64        `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      uint64        `json:"user_id"`
	WalletID    uint64        `json:"wallet_id"`
	Wallet      PaymentWallet `gorm:"foreignKey:WalletID"`
	BlockHeight uint64        `json:"block_height"`
	Amount      string        `json:"amount"`
	Transferred string        `json:"transferred"`
	Status      string        `json:"status"`
	ExpiredTime time.Time     `json:"expired_time"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func (m *PaymentOrder) TableName() string {
	return "payment_order"
}

func (m *PaymentOrder) ToDto() dto.PaymentOrderDTO {
	return dto.PaymentOrderDTO{
		ID:             m.ID,
		UserID:         m.UserID,
		PaymentAddress: m.Wallet.Address,
		BlockHeight:    m.BlockHeight,
		Amount:         m.Amount,
		Transferred:    m.Transferred,
		Status:         m.Status,
		ExpiredTime:    m.ExpiredTime,
	}
}
