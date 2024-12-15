package dto

type PaymentStatistics struct {
	ID               uint64 `json:"id"`
	Granularity      string `json:"granularity"`
	PeriodStart      uint64 `json:"period_start"`
	TotalOrders      uint64 `json:"total_orders"`
	TotalAmount      string `json:"total_amount"`
	TotalTransferred string `json:"total_transferred"`
	Symbol           string `json:"symbol"`
}
