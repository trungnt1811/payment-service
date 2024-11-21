package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
)

type networkMetadataRepository struct {
	db *gorm.DB
}

// NewNetworkMetadataRepository creates a new NetworkMetadataRepository
func NewNetworkMetadataRepository(db *gorm.DB) interfaces.NetworkMetadataRepository {
	return &networkMetadataRepository{
		db: db,
	}
}

func (r *networkMetadataRepository) GetNetworksMetadata(ctx context.Context) ([]domain.NetworkMetadata, error) {
	var networksMetadata []domain.NetworkMetadata
	if err := r.db.WithContext(ctx).Find(&networksMetadata).Error; err != nil {
		return nil, err
	}
	return networksMetadata, nil
}
