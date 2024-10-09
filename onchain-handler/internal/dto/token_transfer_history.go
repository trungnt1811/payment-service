package dto

type TokenTransferHistoryDTO struct {
	ID                       uint64 `json:"id"`
	TokenDistributionAddress string `json:"token_distribution_address"`
	RecipientAddress         string `json:"recipient_address"`
	TransactionHash          string `json:"transaction_hash"`
	TokenAmount              string `json:"token_amount"`
	Status                   int16  `json:"status"`
	TxType                   string `json:"tx_type"`
	ErrorMessage             string `json:"error_message"`
}
