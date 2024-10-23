package redisqueue

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func InitClient(redisClient *redis.Client) (*asynq.Client, asynq.RedisClientOpt) {
	asynqRedisClientOpt := asynq.RedisClientOpt{
		Network:      redisClient.Options().Network,
		Addr:         redisClient.Options().Addr,
		Username:     redisClient.Options().Username,
		Password:     redisClient.Options().Password,
		DB:           redisClient.Options().DB,
		DialTimeout:  redisClient.Options().DialTimeout,
		ReadTimeout:  redisClient.Options().ReadTimeout,
		WriteTimeout: redisClient.Options().WriteTimeout,
		PoolSize:     redisClient.Options().PoolSize,
		TLSConfig:    redisClient.Options().TLSConfig,
	}

	return asynq.NewClient(asynqRedisClientOpt), asynqRedisClientOpt
}
