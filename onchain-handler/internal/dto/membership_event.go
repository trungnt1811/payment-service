package dto

import "time"

type MembershipEventDTO struct {
	ID              uint64    `json:"id"`
	UserAddress     string    `json:"user_address"`
	OrderID         uint64    `json:"order_id"`
	TransactionHash string    `json:"transaction_hash"`
	Amount          string    `json:"amount"`
	Status          uint8     `json:"status"`
	EndDuration     time.Time `json:"end_duration"`
}
