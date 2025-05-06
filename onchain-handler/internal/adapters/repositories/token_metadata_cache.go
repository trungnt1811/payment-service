package repositories

import (
	"context"

	cachetypes "github.com/genefriendway/onchain-handler/internal/adapters/cache/types"
	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

var keyPrefixTokenMetadata = "token_metadata_"

type tokenMetadataCache struct {
	tokenMetadataRepository repotypes.TokenMetadataRepository
	cache                   cachetypes.CacheRepository
}

func NewTokenMetadataCacheRepository(
	repo repotypes.TokenMetadataRepository,
	cache cachetypes.CacheRepository,
) repotypes.TokenMetadataRepository {
	return &tokenMetadataCache{
		tokenMetadataRepository: repo,
		cache:                   cache,
	}
}

func (c *tokenMetadataCache) GetTokensMetadata(ctx context.Context) ([]entities.TokenMetadata, error) {
	key := &cachetypes.Keyer{Raw: keyPrefixTokenMetadata + "GetTokensMetadata"}
	var tokensMetadata []entities.TokenMetadata

	// Try to retrieve data from cache
	err := c.cache.RetrieveItem(key, &tokensMetadata)
	if err == nil {
		// Cache hit
		return tokensMetadata, nil
	}

	// Cache miss, fetch from repository
	tokensMetadata, err = c.tokenMetadataRepository.GetTokensMetadata(ctx)
	if err != nil {
		return nil, err
	}

	// Save to cache with no expiration
	if cacheErr := c.cache.SaveItem(key, tokensMetadata, -1); cacheErr != nil {
		logger.GetLogger().Warnf("Failed to save token metadata to cache: %v", cacheErr)
	}

	return tokensMetadata, nil
}
