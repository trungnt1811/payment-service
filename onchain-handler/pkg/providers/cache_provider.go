package providers

import (
	"context"
	"sync"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/infra/caching"
	"github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

var (
	once      sync.Once
	cacheRepo interfaces.CacheRepository
)

// ProvideCacheRepository provides a singleton instance of CacheRepository.
func ProvideCacheRepository(ctx context.Context) interfaces.CacheRepository {
	once.Do(func() {
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
