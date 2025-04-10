package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/adapters/cache/types"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

// redisCacheClient implements CacheClient interface
type redisCacheClient struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCacheClient initializes Redis cache client with configuration
func NewRedisCacheClient() types.CacheClient {
	config := conf.GetRedisConfiguration()
	ttl, err := time.ParseDuration(config.RedisTTL)
	if err != nil {
		logger.GetLogger().Warnf("Invalid REDIS_TTL format (%s), using default 10m", config.RedisTTL)
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
func (r *redisCacheClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if expiration == 0 {
		expiration = r.ttl
	}

	data, err := json.Marshal(value)
	if err != nil {
		logger.GetLogger().Errorf("Failed to marshal cache value for key: %s", key)
		return err
	}

	err = r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		logger.GetLogger().Errorf("Failed to set cache in Redis for key: %s", key)
	}
	return err
}

// Get retrieves a value from Redis and assigns it to the destination
func (r *redisCacheClient) Get(ctx context.Context, key string, dest any) error {
	// Validate that dest is a pointer and is not nil
	if dest == nil || reflect.ValueOf(dest).Kind() != reflect.Ptr || reflect.ValueOf(dest).IsNil() {
		return fmt.Errorf("destination must be a non-nil pointer")
	}

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return types.ErrNotFound
		}
		logger.GetLogger().Errorf("Failed to get cache from Redis for key: %s", key)
		return err
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		logger.GetLogger().Errorf("Failed to unmarshal cache value for key: %s", key)
	}
	return err
}

// Del removes a key from Redis
func (r *redisCacheClient) Del(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		logger.GetLogger().Errorf("Failed to delete cache from Redis for key: %s", key)
	}
	return err
}

// GetAllMatching retrieves all keys matching a pattern and unmarshals them into a slice of values
func (r *redisCacheClient) GetAllMatching(ctx context.Context, pattern string, valFactory func() any) ([]any, error) {
	var (
		cursor uint64
		keys   []string
		result []any
	)

	// SCAN loop
	for {
		var err error
		var k []string
		k, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan Redis keys: %w", err)
		}
		keys = append(keys, k...)
		if cursor == 0 {
			break
		}
	}

	// Batch MGET to reduce round trips (optional)
	if len(keys) == 0 {
		return result, nil
	}

	values, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to MGET Redis keys: %w", err)
	}

	for i, raw := range values {
		if raw == nil {
			continue
		}
		strVal, ok := raw.(string)
		if !ok {
			logger.GetLogger().Warnf("Unexpected value type for key %s", keys[i])
			continue
		}

		dest := valFactory()
		if dest == nil || reflect.ValueOf(dest).Kind() != reflect.Ptr || reflect.ValueOf(dest).IsNil() {
			logger.GetLogger().Warnf("valFactory returned nil or non-pointer")
			continue
		}

		if err := json.Unmarshal([]byte(strVal), dest); err != nil {
			logger.GetLogger().Warnf("Failed to unmarshal key %s: %v", keys[i], err)
			continue
		}

		result = append(result, dest)
	}

	return result, nil
}
