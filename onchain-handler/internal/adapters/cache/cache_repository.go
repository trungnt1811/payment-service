package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/adapters/cache/types"
)

type cachingRepository struct {
	ctx    context.Context
	client types.CacheClient
}

// NewCachingRepository initializes a new caching repository
func NewCachingRepository(ctx context.Context, client types.CacheClient) types.CacheRepository {
	return &cachingRepository{
		ctx:    ctx,
		client: client,
	}
}

// prependAppPrefix ensures all cache keys have a consistent prefix
func (repo *cachingRepository) prependAppPrefix(key string) string {
	return fmt.Sprintf("%s_%s", conf.GetAppName(), key)
}

// SaveItem saves an item to the cache with a specified expiration time
func (repo *cachingRepository) SaveItem(key fmt.Stringer, val any, expire time.Duration) error {
	prefixedKey := repo.prependAppPrefix(key.String())
	return repo.client.Set(repo.ctx, prefixedKey, val, expire)
}

// RetrieveItem retrieves an item from the cache
func (repo *cachingRepository) RetrieveItem(key fmt.Stringer, val any) error {
	prefixedKey := repo.prependAppPrefix(key.String())
	return repo.client.Get(repo.ctx, prefixedKey, val)
}

// RemoveItem removes an item from the cache
func (repo *cachingRepository) RemoveItem(key fmt.Stringer) error {
	prefixedKey := repo.prependAppPrefix(key.String())
	return repo.client.Del(repo.ctx, prefixedKey)
}

// GetAllMatching retrieves all items whose keys match a given prefix
func (repo *cachingRepository) GetAllMatching(prefix fmt.Stringer, valFactory func() any) ([]any, error) {
	// Full prefix must also be namespaced with the app name
	fullPrefix := repo.prependAppPrefix(prefix.String())
	return repo.client.GetAllMatching(repo.ctx, fullPrefix, valFactory)
}
