package repositories

import (
	"context"

	"github.com/genefriendway/onchain-handler/infra/caching"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

const (
	keyPrefixNetworkMetadata = "network_metadata_"
)

type networkMetadataCache struct {
	networkMetadataRepository *NetworkMetadataRepository
	cache                     infrainterfaces.CacheRepository
}

func NewNetworkMetadataCacheRepository(repo *NetworkMetadataRepository,
	cache infrainterfaces.CacheRepository,
) interfaces.NetworkMetadataRepository {
	return &networkMetadataCache{
		networkMetadataRepository: repo,
		cache:                     cache,
	}
}

func (c *networkMetadataCache) GetNetworksMetadata(ctx context.Context) ([]domain.NetworkMetadata, error) {
	key := &caching.Keyer{Raw: keyPrefixNetworkMetadata + "GetNetworksMetadata"}
	var networkMetadatas []domain.NetworkMetadata

	// Try to retrieve data from cache
	err := c.cache.RetrieveItem(key, &networkMetadatas)
	if err == nil {
		// Cache hit
		return networkMetadatas, nil
	}

	// Cache miss, fetch from repository
	networkMetadatas, err = c.networkMetadataRepository.GetNetworksMetadata(ctx)
	if err != nil {
		return nil, err
	}

	// Save to cache with expiration
	// Duaration -1 means no expiration
	if cacheErr := c.cache.SaveItem(key, networkMetadatas, -1); cacheErr != nil {
		logger.GetLogger().Warnf("Failed to save metadata to cache: %v", cacheErr)
	}

	return networkMetadatas, nil
}
