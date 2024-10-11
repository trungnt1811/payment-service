package dto

type TokenTransferPayloadDTO struct {
	FromAddress string `json:"from_name"`
	ToAddress   string `json:"to_address"`
	TokenAmount string `json:"token_amount"`
	RequestID   string `json:"request_id"`
}
