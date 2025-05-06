package entities

import (
	"time"

	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
)

type TokenMetadata struct {
	ID         uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Symbol     string    `json:"alias"`
	Name       string    `json:"name"`
	IconBase64 string    `json:"icon_base64"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (m *TokenMetadata) TableName() string {
	return "token_metadata"
}

func (m *TokenMetadata) ToDto() dto.TokenMetadataDTO {
	return dto.TokenMetadataDTO{
		ID:         m.ID,
		Symbol:     m.Symbol,
		Name:       m.Name,
		IconBase64: m.IconBase64,
	}
}
