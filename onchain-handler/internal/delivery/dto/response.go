package dto

import "time"

type PaginationDTOResponse struct {
	NextPage               int               `json:"next_page"`
	Page                   int               `json:"page"`
	Size                   int               `json:"size"`
	Total                  uint64            `json:"total,omitempty"`
	TotalTokenAmount       float64           `json:"total_token_amount,omitempty"`
	TotalBalancePerNetwork map[string]string `json:"total_balance_per_network,omitempty"`
	Data                   []any             `json:"data"`
}

type TokenTransferResultDTOResponse struct {
	RequestID    string `json:"request_id"`
	Status       bool   `json:"status"`
	ErrorMessage string `json:"error_message"`
}

type PaymentOrderDTOResponse struct {
	ID                  uint64              `json:"id"`
	RequestID           string              `json:"request_id"`
	Network             string              `json:"network"`
	Amount              string              `json:"amount"`
	Transferred         string              `json:"transferred"`
	Status              string              `json:"status"`
	WebhookURL          string              `json:"webhook_url"`
	Symbol              string              `json:"symbol"`
	BlockHeight         uint64              `json:"block_height"`
	UpcomingBlockHeight uint64              `json:"upcoming_block_height,omitempty"`
	PaymentAddress      string              `json:"payment_address,omitempty"`
	SucceededAt         *time.Time          `json:"succeeded_at,omitempty"`
	CreatedAt           time.Time           `json:"created_at"`
	Expired             uint64              `json:"expired,omitempty"`
	EventHistories      []PaymentHistoryDTO `json:"event_histories,omitempty"`
}
