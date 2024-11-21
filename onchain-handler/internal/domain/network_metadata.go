package domain

import (
	"time"

	"github.com/genefriendway/onchain-handler/internal/dto"
)

type NetworkMetadata struct {
	ID         uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Alias      string    `json:"alias"`
	Name       string    `json:"name"`
	IconBase64 string    `json:"icon_base64"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (m *NetworkMetadata) TableName() string {
	return "blockchain_network_metadata"
}

func (m *NetworkMetadata) ToDto() dto.NetworkMetadataDTO {
	return dto.NetworkMetadataDTO{
		ID:         m.ID,
		Alias:      m.Alias,
		Name:       m.Name,
		IconBase64: m.IconBase64,
	}
}
