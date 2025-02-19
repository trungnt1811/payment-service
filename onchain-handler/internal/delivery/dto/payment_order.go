package dto

import "time"

type PaymentOrderDTO struct {
	ID                  uint64           `json:"id"`
	RequestID           string           `json:"request_id"`
	VendorID            string           `json:"vendor_id"`
	PaymentAddress      string           `json:"payment_address"`
	Wallet              PaymentWalletDTO `json:"wallet"`
	BlockHeight         uint64           `json:"block_height"`
	UpcomingBlockHeight uint64           `json:"upcoming_block_height"`
	Amount              string           `json:"amount"`
	Transferred         string           `json:"transferred"`
	Symbol              string           `json:"symbol"`
	Network             string           `json:"network"`
	Status              string           `json:"status"`
	WebhookURL          string           `json:"webhook_url"`
	SucceededAt         time.Time        `json:"succeeded_at,omitempty"`
	ExpiredTime         time.Time        `json:"expired_time"`
}

type CreatedPaymentOrderDTO struct {
	ID             uint64 `json:"id"`
	RequestID      string `json:"request_id"`
	PaymentAddress string `json:"payment_address"`
	Amount         string `json:"amount"`
	Symbol         string `json:"symbol"`
	Network        string `json:"network"`
	Expired        uint64 `json:"expired"`
	Signature      []byte `json:"signature"`
}
