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
	SucceededAt    time.Time `json:"succeeded_at,omitempty"`
	ExpiredTime    time.Time `json:"expired_time"`
}

type CreatedPaymentOrderDTO struct {
	ID             uint64 `json:"id"`
	RequestID      string `json:"request_id"`
	PaymentAddress string `json:"payment_address"`
	Amount         string `json:"amount"`
	Symbol         string `json:"symbol"`
	Signature      []byte `json:"signature"`
}
