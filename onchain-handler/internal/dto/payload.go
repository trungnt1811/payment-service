package dto

type TokenTransferPayloadDTO struct {
	RecipientAddress string `json:"recipient_address"`
	TokenAmount      string `json:"token_amount"`
}
