package cache

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/adapters/cache/types"
)

type goCacheClient struct {
	cache *cache.Cache
}

// NewGoCacheClient initializes a new cache client with default expiration and cleanup interval
func NewGoCacheClient() types.CacheClient {
	return &goCacheClient{
		cache: cache.New(constants.DefaultExpiration, constants.CleanupInterval),
	}
}

// Set adds an item to the cache with a specified expiration time
// If the duration is 0 (DefaultExpiration), the cache's default expiration time is used.
// If it is -1 (NoExpiration), the item never expires.
func (c *goCacheClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	c.cache.Set(key, value, expiration)
	return nil
}

// Get retrieves an item from the cache and assigns it to the destination using reflection
func (c *goCacheClient) Get(ctx context.Context, key string, dest any) error {
	cachedValue, found := c.cache.Get(key)
	if !found {
		return fmt.Errorf("item not found in cache")
	}

	// Use reflection to set the value to the destination
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.IsNil() {
		return fmt.Errorf("destination must be a non-nil pointer")
	}

	cachedVal := reflect.ValueOf(cachedValue)

	if cachedVal.Type().AssignableTo(destVal.Elem().Type()) {
		destVal.Elem().Set(cachedVal)
		return nil
	}

	return fmt.Errorf("cached value type (%v) does not match destination type (%v)", cachedVal.Type(), destVal.Elem().Type())
}

// Del deletes an item from the cache
func (c *goCacheClient) Del(ctx context.Context, key string) error {
	c.cache.Delete(key)
	return nil
}
