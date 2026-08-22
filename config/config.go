package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type appConfig struct {
	SecretJwt         string
	DBUrl             string
	DBMaxPoolSize     int
	DBMinPoolSize     int
	DBIdleTimeout     time.Duration
	DBMaxConnLifetime time.Duration
	Port              string
	Env               string
	LogLevel          string
	MinioEndpoint     string
	MinioAccessKey    string
	MinioSecretKey    string
	MinioUseSSL       bool
	MinioBucketName   string
	MinioBaseURL      string
	Secret            string
	AllowedOrigins    string
	RabbitmqUrl       string
	RedisUrl          string
	RedisUsername     string
	RedisPassword     string
	ServiceMasterDataUrl string
}

var Config appConfig

func init() {
	// Load env
	log.Println("Loading .env file")
	viper.SetConfigFile(".env") // atau bisa juga pakai viper.SetConfigName("app") + viper.AddConfigPath(".")
	viper.AutomaticEnv()        // override dengan ENV OS kalau ada

	if err := viper.ReadInConfig(); err != nil {
		log.Println("No .env file found, fallback to system environment")
	}

	Config = appConfig{
		SecretJwt:            viper.GetString("SECRET_JWT"),
		DBUrl:                viper.GetString("DB_URL"),
		DBMaxPoolSize:        viper.GetInt("DB_MAX_POOL_SIZE"),
		DBMinPoolSize:        viper.GetInt("DB_MIN_POOL_SIZE"),
		DBIdleTimeout:        viper.GetDuration("DB_IDLE_TIMEOUT") * time.Second,
		DBMaxConnLifetime:    viper.GetDuration("DB_MAX_CONN_LIFETIME") * time.Second,
		Port:                 viper.GetString("APP_PORT"),
		Env:                  viper.GetString("APP_ENV"),
		LogLevel:             viper.GetString("APP_LOG_LEVEL"),
		MinioSecretKey:       viper.GetString("MINIO_SECRET_KEY"),
		MinioAccessKey:       viper.GetString("MINIO_ACCESS_KEY"),
		MinioEndpoint:        viper.GetString("MINIO_ENDPOINT"),
		MinioUseSSL:          viper.GetBool("MINIO_USE_SSL"),
		MinioBucketName:      viper.GetString("MINIO_BUCKET_NAME"),
		MinioBaseURL:         viper.GetString("MINIO_BASE_URL"),
		Secret:               viper.GetString("APP_SECRET"),
		AllowedOrigins:       viper.GetString("APP_ALLOWED_ORIGINS"),
		RabbitmqUrl:          viper.GetString("RABBITMQ_URL"),
		RedisUrl:             viper.GetString("REDIS_URL"),
		RedisUsername:        viper.GetString("REDIS_USERNAME"),
		RedisPassword:        viper.GetString("REDIS_PASSWORD"),
		ServiceMasterDataUrl: viper.GetString("SERVICE_MASTER_DATA_URL"),
	}
	if Config.ServiceMasterDataUrl == "" {
		Config.ServiceMasterDataUrl = "http://localhost:8080"
	}
}
