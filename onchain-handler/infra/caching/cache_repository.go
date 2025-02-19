package caching

import (
	"context"
	"fmt"
	"time"

	"github.com/genefriendway/onchain-handler/infra/caching/types"
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

// SaveItem saves an item to the cache with a specified expiration time
func (repo *cachingRepository) SaveItem(key fmt.Stringer, val interface{}, expire time.Duration) error {
	return repo.client.Set(repo.ctx, key.String(), val, expire)
}

// RetrieveItem retrieves an item from the cache
func (repo *cachingRepository) RetrieveItem(key fmt.Stringer, val interface{}) error {
	return repo.client.Get(repo.ctx, key.String(), val)
}

// RemoveItem removes an item from the cache
func (repo *cachingRepository) RemoveItem(key fmt.Stringer) error {
	return repo.client.Del(repo.ctx, key.String())
}
