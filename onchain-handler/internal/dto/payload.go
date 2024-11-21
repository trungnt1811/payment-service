package dto

type TokenTransferPayloadDTO struct {
	Network     string `json:"network"`
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
	Network   string `json:"network"`
}

type PaymentOrderNetworkPayloadDTO struct {
	ID      uint64 `json:"id"`
	Network string `json:"network"`
}

type PaymentEventPayloadDTO struct {
	PaymentOrderID  uint64 `json:"payment_order_id"`
	Network         string `json:"network"`
	TransactionHash string `json:"transaction_hash"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	ContractAddress string `json:"contract_address"`
	TokenSymbol     string `json:"token_symbol"`
	Amount          string `json:"amount"`
}

type PaymentWalletPayloadDTO struct {
	ID      uint64 `json:"id"`
	Address string `json:"address"`
	InUse   bool   `json:"in_use"`
}

type UserWalletPayloadDTO struct {
	UserID  uint64 `json:"user_id"`
	Address string `json:"address"`
}
