package caching

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/infra/interfaces"
	pkglogger "github.com/genefriendway/onchain-handler/pkg/logger"
)

// redisCacheClient implements CacheClient interface
type redisCacheClient struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCacheClient initializes Redis cache client with configuration
func NewRedisCacheClient() interfaces.CacheClient {
	config := conf.GetRedisConfiguration()
	ttl, err := time.ParseDuration(config.RedisTtl)
	if err != nil {
		pkglogger.GetLogger().Warnf("Invalid REDIS_TTL format (%s), using default 10m", config.RedisTtl)
		ttl = 10 * time.Minute
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisAddress,
	})

	return &redisCacheClient{
		client: redisClient,
		ttl:    ttl,
	}
}

// Set stores a key-value pair in Redis with expiration
func (r *redisCacheClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if expiration == 0 {
		expiration = r.ttl
	}

	data, err := json.Marshal(value)
	if err != nil {
		pkglogger.GetLogger().Errorf("Failed to marshal cache value for key: %s", key)
		return err
	}

	err = r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		pkglogger.GetLogger().Errorf("Failed to set cache in Redis for key: %s", key)
	}
	return err
}

// Get retrieves a value from Redis and assigns it to the destination
func (r *redisCacheClient) Get(ctx context.Context, key string, dest interface{}) error {
	// Validate that dest is a pointer and is not nil
	if dest == nil || reflect.ValueOf(dest).Kind() != reflect.Ptr || reflect.ValueOf(dest).IsNil() {
		return fmt.Errorf("destination must be a non-nil pointer")
	}

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache miss for key: %s", key)
		}
		pkglogger.GetLogger().Errorf("Failed to get cache from Redis for key: %s", key)
		return err
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		pkglogger.GetLogger().Errorf("Failed to unmarshal cache value for key: %s", key)
	}
	return err
}

// Del removes a key from Redis
func (r *redisCacheClient) Del(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		pkglogger.GetLogger().Errorf("Failed to delete cache from Redis for key: %s", key)
	}
	return err
}
