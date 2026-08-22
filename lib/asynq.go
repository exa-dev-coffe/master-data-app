package lib

import (
	"eka-dev.cloud/master-data/config"
	"github.com/gofiber/fiber/v2/log"
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
		Password: config.Config.RedisPassword,
	}
	AsynqClient = asynq.NewClient(redisOpt)
	log.Info("Asynq Client initialized successfully in master-data at ", redisUrl)
}
