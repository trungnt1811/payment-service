package domain

import (
	"time"

	"github.com/genefriendway/onchain-handler/internal/dto"
)

// PaymentStatistics represents the payment statistics domain model.
type PaymentStatistics struct {
	ID               uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Granularity      string    `json:"granularity"`  // Enum: DAILY, WEEKLY, MONTHLY, YEARLY
	PeriodStart      time.Time `json:"period_start"` // Start of the period
	TotalOrders      uint64    `json:"total_orders"`
	TotalAmount      string    `json:"total_amount"`
	TotalTransferred string    `json:"total_transferred"`
	Symbol           string    `json:"symbol"`
	VendorID         string    `json:"vendor_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName overrides the default table name for GORM.
func (m *PaymentStatistics) TableName() string {
	return "payment_statistics"
}

func (m *PaymentStatistics) ToDto() dto.PaymentStatistics {
	return dto.PaymentStatistics{
		PeriodStart:      uint64(m.PeriodStart.UTC().Unix()),
		TotalOrders:      m.TotalOrders,
		TotalAmount:      m.TotalAmount,
		TotalTransferred: m.TotalTransferred,
		Symbol:           m.Symbol,
	}
}
