package dto

type TokenTransferHistoryDTO struct {
	ID              uint64  `json:"id"`
	RequestID       string  `json:"request_id"`
	TransactionHash string  `json:"transaction_hash"`
	FromAddress     string  `json:"from_address"`
	ToAddress       string  `json:"to_address"`
	TokenAmount     uint64  `json:"token_amount"`
	Fee             float64 `json:"fee"`
	Symbol          string  `json:"symbol"`
	Status          bool    `json:"status"`
	ErrorMessage    string  `json:"error_message"`
}
