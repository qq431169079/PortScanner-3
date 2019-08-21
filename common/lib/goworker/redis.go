package goworker

import (
	"github.com/go-redis/redis"
)

func newRedisClient(redisUrl string) (*redis.Client, error) {
	option, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}
	option.PoolSize = workerSettings.Connections
	option.MaxRetries = workerSettings.MaxRetries
	option.DialTimeout = workerSettings.DialTimeout
	option.ReadTimeout = workerSettings.ReadTimeout
	option.WriteTimeout = workerSettings.WriteTimeout
	return redis.NewClient(option), nil
}
