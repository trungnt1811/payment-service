package repositories

import (
	"context"

	"gorm.io/gorm"

	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type networkMetadataRepository struct {
	db *gorm.DB
}

// NewNetworkMetadataRepository creates a new NetworkMetadataRepository
func NewNetworkMetadataRepository(db *gorm.DB) repotypes.NetworkMetadataRepository {
	return &networkMetadataRepository{
		db: db,
	}
}

func (r *networkMetadataRepository) GetNetworksMetadata(ctx context.Context) ([]entities.NetworkMetadata, error) {
	var networksMetadata []entities.NetworkMetadata
	if err := r.db.WithContext(ctx).Find(&networksMetadata).Error; err != nil {
		return nil, err
	}
	return networksMetadata, nil
}
