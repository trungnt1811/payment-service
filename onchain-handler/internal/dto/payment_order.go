package dto

import "time"

type PaymentOrderDTO struct {
	ID             uint64    `json:"id"`
	UserID         uint64    `json:"user_id"`
	PaymentAddress string    `json:"payment_address"`
	BlockHeight    uint64    `json:"block_height"`
	Amount         string    `json:"amount"`
	Transferred    string    `json:"transferred"`
	Status         string    `json:"status"`
	ExpiredTime    time.Time `json:"expired_time"`
}
