package dto

type PaymentWalletDTO struct {
	ID         uint64 `json:"id"`
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"`
	InUse      bool   `json:"in_use"`
}
