package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type NetworkMetadataRepository struct {
	db *gorm.DB
}

// NewNetworkMetadataRepository creates a new NetworkMetadataRepository
func NewNetworkMetadataRepository(db *gorm.DB) *NetworkMetadataRepository {
	return &NetworkMetadataRepository{
		db: db,
	}
}

func (r *NetworkMetadataRepository) GetNetworksMetadata(ctx context.Context) ([]entities.NetworkMetadata, error) {
	var networksMetadata []entities.NetworkMetadata
	if err := r.db.WithContext(ctx).Find(&networksMetadata).Error; err != nil {
		return nil, err
	}
	return networksMetadata, nil
}
