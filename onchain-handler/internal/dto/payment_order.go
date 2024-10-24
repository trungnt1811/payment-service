package dto

import "time"

type PaymentOrderDTO struct {
	ID             uint64    `json:"id"`
	RequestID      string    `json:"request_id"`
	PaymentAddress string    `json:"payment_address"`
	BlockHeight    uint64    `json:"block_height"`
	Amount         string    `json:"amount"`
	Transferred    string    `json:"transferred"`
	Symbol         string    `json:"symbol"`
	Status         string    `json:"status"`
	ExpiredTime    time.Time `json:"expired_time"`
}
