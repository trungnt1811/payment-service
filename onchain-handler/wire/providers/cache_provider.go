package providers

import (
	"context"
	"sync"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/infra/caching"
	cachetypes "github.com/genefriendway/onchain-handler/infra/caching/types"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

var (
	cacheOnce sync.Once
	cacheRepo cachetypes.CacheRepository
)

// ProvideCacheRepository provides a singleton instance of CacheRepository.
func ProvideCacheRepository(ctx context.Context) cachetypes.CacheRepository {
	cacheOnce.Do(func() {
		cacheType := conf.GetCacheType()
		switch cacheType {
		case "redis":
			// Using Redis cache
			logger.GetLogger().Info("Using Redis cache")
			cacheClient := caching.NewRedisCacheClient()
			cacheRepo = caching.NewCachingRepository(ctx, cacheClient)
		default:
			// Using in-memory cache (default)
			logger.GetLogger().Info("Using in-memory cache (default)")
			cacheClient := caching.NewGoCacheClient()
			cacheRepo = caching.NewCachingRepository(ctx, cacheClient)
		}
	})
	return cacheRepo
}
