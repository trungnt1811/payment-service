package dto

type PaymentWalletDTO struct {
	ID      uint64 `json:"id"`
	Address string `json:"address"`
	InUse   bool   `json:"in_use"`
}
