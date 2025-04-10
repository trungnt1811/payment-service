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
		return types.ErrNotFound // standardized error
	}

	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.IsNil() {
		return fmt.Errorf("destination must be a non-nil pointer")
	}

	cachedVal := reflect.ValueOf(cachedValue)

	if cachedVal.Type().AssignableTo(destVal.Elem().Type()) {
		destVal.Elem().Set(cachedVal)
		return nil
	}

	return fmt.Errorf(
		"cached value type (%v) does not match destination type (%v)",
		cachedVal.Type(), destVal.Elem().Type(),
	)
}

// Del deletes an item from the cache
func (c *goCacheClient) Del(ctx context.Context, key string) error {
	c.cache.Delete(key)
	return nil
}

// GetAllMatching retrieves all items from the cache that match a given prefix
func (c *goCacheClient) GetAllMatching(ctx context.Context, prefix string, valFactory func() any) ([]any, error) {
	items := c.cache.Items()
	var result []any

	for key, item := range items {
		if prefix != "" && !startsWith(key, prefix) {
			continue
		}

		valPtr := valFactory()
		val := reflect.ValueOf(valPtr)

		if val.Kind() != reflect.Ptr || val.IsNil() {
			return nil, fmt.Errorf("valFactory must return a non-nil pointer")
		}

		cachedVal := reflect.ValueOf(item.Object)
		if cachedVal.Type().AssignableTo(val.Elem().Type()) {
			val.Elem().Set(cachedVal)
			result = append(result, valPtr)
		}
	}

	return result, nil
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
