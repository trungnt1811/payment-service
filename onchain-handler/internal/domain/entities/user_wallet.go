package entities

import (
	"time"

	"github.com/genefriendway/onchain-handler/internal/domain/dto"
)

type UserWallet struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint64    `json:"user_id"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m *UserWallet) TableName() string {
	return "user_wallet"
}

func (m *UserWallet) ToDto() dto.UserWalletDTO {
	return dto.UserWalletDTO{
		UserID:  m.UserID,
		Address: m.Address,
	}
}
