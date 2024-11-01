package dto

type UserWalletDTO struct {
	UserID  uint64 `json:"user_id"`
	Address string `json:"address"`
}
