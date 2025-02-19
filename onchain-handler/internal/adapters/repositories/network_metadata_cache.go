package repositories

import (
	"context"

	cachetypes "github.com/genefriendway/onchain-handler/infra/caching/types"
	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

const (
	keyPrefixNetworkMetadata = "network_metadata_"
)

type networkMetadataCache struct {
	networkMetadataRepository repotypes.NetworkMetadataRepository
	cache                     cachetypes.CacheRepository
}

func NewNetworkMetadataCacheRepository(
	repo repotypes.NetworkMetadataRepository,
	cache cachetypes.CacheRepository,
) repotypes.NetworkMetadataRepository {
	return &networkMetadataCache{
		networkMetadataRepository: repo,
		cache:                     cache,
	}
}

func (c *networkMetadataCache) GetNetworksMetadata(ctx context.Context) ([]entities.NetworkMetadata, error) {
	key := &cachetypes.Keyer{Raw: keyPrefixNetworkMetadata + "GetNetworksMetadata"}
	var networkMetadatas []entities.NetworkMetadata

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
