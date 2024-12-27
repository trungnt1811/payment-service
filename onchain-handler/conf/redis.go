package conf

import (
	"github.com/redis/go-redis/v9"
)

func RedisConn() *redis.Client {
	address := GetRedisConnectionURL()
	rdb := redis.NewClient(&redis.Options{
		Addr: address,
		DB:   0,
	})

	return rdb
}
