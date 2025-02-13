package entities

import "time"

type PaymentEventHistory struct {
	ID              uint64       `json:"id" gorm:"primaryKey;autoIncrement"`
	PaymentOrderID  uint64       `json:"payment_order_id"`
	PaymentOrder    PaymentOrder `gorm:"foreignKey:PaymentOrderID"`
	TransactionHash string       `json:"transaction_hash"`
	FromAddress     string       `json:"from_address"`
	ToAddress       string       `json:"to_address"`
	ContractAddress string       `json:"contract_address"`
	TokenSymbol     string       `json:"token_symbol"`
	Network         string       `json:"network"`
	Amount          string       `json:"amount"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

func (m *PaymentEventHistory) TableName() string {
	return "payment_event_history"
}
