package dto

type TokenTransferPayloadDTO struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	TokenAmount string `json:"token_amount"`
	RequestID   string `json:"request_id"`
	Symbol      string `json:"symbol"`
}

type PaymentOrderPayloadDTO struct {
	RequestID string `json:"request_id"`
	Amount    string `json:"amount"`
	Symbol    string `json:"symbol"`
}
