package dto

type TokenStats struct {
	Symbol           string `json:"symbol"`
	TotalOrders      uint64 `json:"total_orders"`
	TotalAmount      string `json:"total_amount"`
	TotalTransferred string `json:"total_transferred"`
}

type PeriodStatistics struct {
	PeriodStart uint64       `json:"period_start"`
	TokenStats  []TokenStats `json:"token_stats"`
}
