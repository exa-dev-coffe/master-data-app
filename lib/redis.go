package lib

import (
	"log/slog"

	"eka-dev.cloud/master-data/config"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() {
	redisUrl := config.Config.RedisUrl
	if redisUrl == "" {
		redisUrl = "localhost:6379"
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Username: config.Config.RedisUsername,
		Password: config.Config.RedisPassword,
		DB:       0,
	})
	slog.Info("Redis client initialized successfully in master-data", "redis_url", redisUrl)
}
