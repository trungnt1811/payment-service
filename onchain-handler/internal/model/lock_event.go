package model

import "time"

type LockEvent struct {
	ID              uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserAddress     string    `json:"user_address"`
	LockID          uint64    `json:"lock_id"`
	TransactionHash string    `json:"transaction_hash"`
	Amount          string    `json:"amount"`
	CurrentBalance  string    `json:"current_balance"`
	LockAction      string    `json:"lock_action"`
	Status          uint8     `json:"status"`
	LockDuration    uint64    `json:"lock_duration"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	EndDuration     time.Time `json:"end_duration"`
}

func (m *LockEvent) TableName() string {
	return "lock_event"
}
