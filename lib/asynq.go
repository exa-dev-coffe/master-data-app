package lib

import (
	"log/slog"

	"eka-dev.cloud/master-data/config"
	"github.com/hibiken/asynq"
)

var AsynqClient *asynq.Client

func InitAsynq() {
	redisUrl := config.Config.RedisUrl
	if redisUrl == "" {
		redisUrl = "localhost:6379"
	}
	redisOpt := asynq.RedisClientOpt{
		Addr:     redisUrl,
		Username: config.Config.RedisUsername,
		Password: config.Config.RedisPassword,
	}
	AsynqClient = asynq.NewClient(redisOpt)
	slog.Info("Asynq Client initialized successfully in master-data", "redis_url", redisUrl)
}
