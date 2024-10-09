package dto

type TokenTransferHistoryDTO struct {
	ID              uint64 `json:"id"`
	TransactionHash string `json:"transaction_hash"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	TokenAmount     string `json:"token_amount"`
	Symbol          string `json:"symbol"`
	Status          bool   `json:"status"`
	ErrorMessage    string `json:"error_message"`
}
