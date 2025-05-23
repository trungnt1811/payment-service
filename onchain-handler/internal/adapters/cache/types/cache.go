package types

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("cache: item not found")

type Keyer struct {
	Raw string
}

type Value struct {
	Raw string
}

func (k *Keyer) String() string {
	return k.Raw
}

type CacheClient interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Get(ctx context.Context, key string, dest any) error
	Del(ctx context.Context, key string) error
	GetAllMatching(ctx context.Context, prefix string, valFactory func() any) ([]any, error)
}

type CacheRepository interface {
	SaveItem(key fmt.Stringer, val any, expire time.Duration) error
	RetrieveItem(key fmt.Stringer, val any) error
	RemoveItem(key fmt.Stringer) error
	GetAllMatching(prefix fmt.Stringer, valFactory func() any) ([]any, error)
}
