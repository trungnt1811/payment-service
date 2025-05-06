package repositories

import (
	"context"

	"gorm.io/gorm"

	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type tokenMetadataRepository struct {
	db *gorm.DB
}

// NewTokenMetadataRepository creates a new TokenMetadataRepository
func NewTokenMetadataRepository(db *gorm.DB) repotypes.TokenMetadataRepository {
	return &tokenMetadataRepository{
		db: db,
	}
}

// GetTokensMetadata retrieves all token metadata records
func (r *tokenMetadataRepository) GetTokensMetadata(ctx context.Context) ([]entities.TokenMetadata, error) {
	var tokensMetadata []entities.TokenMetadata
	if err := r.db.WithContext(ctx).Find(&tokensMetadata).Error; err != nil {
		return nil, err
	}
	return tokensMetadata, nil
}
