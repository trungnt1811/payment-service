package caching

import (
	"context"
	"fmt"
	"time"
)

type CacheClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Del(ctx context.Context, key string) error
}

type CacheRepository interface {
	SaveItem(key fmt.Stringer, val interface{}, expire time.Duration) error
	RetrieveItem(key fmt.Stringer, val interface{}) error
	RemoveItem(key fmt.Stringer) error
}
