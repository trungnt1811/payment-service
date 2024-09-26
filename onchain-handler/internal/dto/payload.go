package dto

type TransferTokenPayloadDTO struct {
	RecipientAddress string `json:"recipient_address"`
	TokenAmount      string `json:"token_amount"`
	TxType           string `json:"tx_type"`
}
