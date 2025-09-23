package db

import (
	"log"

	"eka-dev.com/master-data/config"
	"eka-dev.com/master-data/utils/constant"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func init() {
	log.Println("databases init")
	dsn := config.Config.DBUrl
	if dsn == "" {
		log.Fatalln("Database DSN is not set")
	}

	var err error
	DB, err = sqlx.Open(constant.DialectPostgres, dsn)
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}

	// Db configuration
	DB.SetMaxOpenConns(config.Config.DBMaxPoolSize)
	DB.SetMaxIdleConns(config.Config.DBMinPoolSize)
	DB.SetConnMaxIdleTime(config.Config.DBIdleTimeout)
	DB.SetConnMaxLifetime(config.Config.DBMaxConnLifetime)

	err = DB.Ping()
	if err != nil {
		log.Fatalln("Failed to ping database:", err)
	}

	log.Println("Database connection established")

	DB = DB
}
