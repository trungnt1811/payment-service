package dto

import "time"

type LockEventDTO struct {
	ID              uint64    `json:"id"`
	UserAddress     string    `json:"user_address"`
	LockID          uint64    `json:"lock_id"`
	TransactionHash string    `json:"transaction_hash"`
	DepositAmount   string    `json:"deposit_amount"`
	Amount          string    `json:"amount"`
	CurrentBalance  string    `json:"current_balance"`
	LockAction      string    `json:"lock_action"`
	Status          uint8     `json:"status"`
	LockTimestamp   uint64    `json:"lock_timestamp"`
	LockDuration    uint64    `json:"lock_duration"`
	CreatedAt       time.Time `json:"created_at"`
	EndDuration     time.Time `json:"end_duration"`
}
