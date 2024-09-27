package model

import (
	"time"

	"github.com/genefriendway/onchain-handler/internal/dto"
)

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
	LockTimestamp   uint64    `json:"lock_timestamp"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	EndDuration     time.Time `json:"end_duration"`
}

func (m *LockEvent) TableName() string {
	return "lock_event"
}

func (m *LockEvent) ToDto() dto.LockEventDTO {
	return dto.LockEventDTO{
		ID:              m.ID,
		UserAddress:     m.UserAddress,
		LockID:          m.LockID,
		TransactionHash: m.TransactionHash,
		Amount:          m.Amount,
		CurrentBalance:  m.CurrentBalance,
		LockAction:      m.LockAction,
		Status:          m.Status,
		LockDuration:    m.LockDuration,
		LockTimestamp:   m.LockTimestamp,
		CreatedAt:       m.CreatedAt,
		EndDuration:     m.EndDuration,
	}
}
