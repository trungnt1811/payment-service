package dto

type TokenTransferPayloadDTO struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	TokenAmount uint64 `json:"token_amount"`
	RequestID   string `json:"request_id"`
}
