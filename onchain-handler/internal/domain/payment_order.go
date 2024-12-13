package domain

import (
	"time"

	"github.com/genefriendway/onchain-handler/internal/dto"
)

type PaymentOrder struct {
	ID                    uint64                `json:"id" gorm:"primaryKey;autoIncrement"`
	RequestID             string                `json:"request_id"`
	VendorID              string                `json:"vendor_id"`
	WalletID              uint64                `json:"wallet_id"`
	Wallet                PaymentWallet         `gorm:"foreignKey:WalletID"`
	BlockHeight           uint64                `json:"block_height"`
	UpcomingBlockHeight   uint64                `json:"upcoming_block_height"`
	Amount                string                `json:"amount"`
	Transferred           string                `json:"transferred"`
	Symbol                string                `json:"symbol"`
	Network               string                `json:"network"`
	Status                string                `json:"status"`
	WebhookURL            string                `json:"webhook_url"`
	SucceededAt           time.Time             `json:"succeeded_at"`
	ExpiredTime           time.Time             `json:"expired_time"`
	CreatedAt             time.Time             `json:"created_at"`
	UpdatedAt             time.Time             `json:"updated_at"`
	PaymentEventHistories []PaymentEventHistory `json:"payment_event_histories" gorm:"foreignKey:PaymentOrderID"`
}

func (m *PaymentOrder) TableName() string {
	return "payment_order"
}

func (m *PaymentOrder) ToDto() dto.PaymentOrderDTO {
	return dto.PaymentOrderDTO{
		ID:                  m.ID,
		RequestID:           m.RequestID,
		PaymentAddress:      m.Wallet.Address,
		Wallet:              m.Wallet.ToDto(),
		BlockHeight:         m.BlockHeight,
		UpcomingBlockHeight: m.UpcomingBlockHeight,
		Amount:              m.Amount,
		Transferred:         m.Transferred,
		Symbol:              m.Symbol,
		Network:             m.Network,
		Status:              m.Status,
		WebhookURL:          m.WebhookURL,
		ExpiredTime:         m.ExpiredTime,
	}
}

func (m *PaymentOrder) ToCreatedPaymentOrderDTO() dto.CreatedPaymentOrderDTO {
	return dto.CreatedPaymentOrderDTO{
		ID:             m.ID,
		RequestID:      m.RequestID,
		PaymentAddress: m.Wallet.Address,
		Amount:         m.Amount,
		Symbol:         m.Symbol,
		Network:        m.Network,
	}
}
