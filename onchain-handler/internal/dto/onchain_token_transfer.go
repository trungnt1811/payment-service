package dto

type TokenTransferHistoryDTO struct {
	ID              uint64 `json:"id"`
	RequestID       string `json:"request_id,omitempty"`
	Network         string `json:"network"`
	TransactionHash string `json:"transaction_hash"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	TokenAmount     string `json:"token_amount"`
	Fee             string `json:"fee"`
	Symbol          string `json:"symbol"`
	Status          bool   `json:"status"`
	Type            string `json:"type"`
	ErrorMessage    string `json:"error_message,omitempty"`
}
