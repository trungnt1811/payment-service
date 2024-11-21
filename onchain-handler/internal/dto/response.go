package dto

import "time"

type PaginationDTOResponse struct {
	NextPage int           `json:"next_page"`
	Page     int           `json:"page"`
	Size     int           `json:"size"`
	Total    int64         `json:"total,omitempty"`
	Data     []interface{} `json:"data"`
}

type TokenTransferResultDTOResponse struct {
	RequestID    string `json:"request_id"`
	Status       bool   `json:"status"`
	ErrorMessage string `json:"error_message"`
}

type PaymentOrderDTOResponse struct {
	ID             uint64              `json:"id"`
	RequestID      string              `json:"request_id"`
	Network        string              `json:"network"`
	Amount         string              `json:"amount"`
	Transferred    string              `json:"transferred"`
	Status         string              `json:"status"`
	SucceededAt    *time.Time          `json:"succeeded_at,omitempty"`
	Expired        *uint64             `json:"expired,omitempty"`
	EventHistories []PaymentHistoryDTO `json:"event_histories,omitempty"`
}
