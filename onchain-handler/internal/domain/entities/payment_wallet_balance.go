package entities

import "time"

type PaymentWalletBalance struct {
	ID        uint64        `json:"id" gorm:"primaryKey;autoIncrement"`
	WalletID  uint64        `json:"wallet_id"`
	Wallet    PaymentWallet `gorm:"foreignKey:WalletID"`
	Network   string        `json:"network"`
	Symbol    string        `json:"symbol"`
	Balance   string        `json:"balance"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

func (m *PaymentWalletBalance) TableName() string {
	return "payment_wallet_balance"
}
