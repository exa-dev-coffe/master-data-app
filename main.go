package main

import (
	"context"
	"log/slog"
	"os"

	"eka-dev.cloud/master-data/config"
	"eka-dev.cloud/master-data/db"
	_ "eka-dev.cloud/master-data/db"
	"eka-dev.cloud/master-data/lib"
	_ "eka-dev.cloud/master-data/lib"
	"eka-dev.cloud/master-data/middleware"
	"eka-dev.cloud/master-data/modules/category"
	"eka-dev.cloud/master-data/modules/internalModule"
	"eka-dev.cloud/master-data/modules/menu"
	"eka-dev.cloud/master-data/modules/promotion"
	"eka-dev.cloud/master-data/modules/table"
	"eka-dev.cloud/master-data/modules/upload"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	middleware.InitLogger("master-data")
	
	shutdown, err := lib.InitTracer("master-data")
	if err == nil && shutdown != nil {
		defer shutdown(context.Background())
	}

	// Load env
	initiator()

	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			slog.Error("Error closing database connection", "error", err)
		}
	}(db.DB)

}

func initiator() {
	// Initialize Asynq Client for background task scheduling
	lib.InitAsynq()

	// Initialize the fiber app
	fiberApp := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	fiberApp.Use(requestid.New())
	fiberApp.Use(middleware.TraceMiddleware())
	fiberApp.Use(middleware.RequestLogger())

	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		err := db.DB.Ping()
		if err != nil {
			slog.Error("Database ping failed", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(response.InternalServerError("Database connection error", nil))
		}
		err = lib.HealthCheck()
		if err != nil {
			slog.Error("RabbitMQ connection failed", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(response.InternalServerError("RabbitMQ connection error", nil))
		}
		return c.Status(fiber.StatusOK).JSON(response.Success("OK", nil))
	})

	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins: config.Config.AllowedOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Timestamp, X-Signature",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	ch, err := lib.GetChannel()
	if err != nil {
		slog.Error("Failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}

	defer func(ch *amqp.Channel) {
		err := ch.Close()
		if err != nil {
			slog.Error("Error closing RabbitMQ channel", "error", err)
		}
	}(ch)

	// Initialize RabbitMQ consumers
	// Menu
	menu.NewListener(ch, db.DB)

	// Initialize routes
	// Categories
	category.NewHandler(fiberApp, db.DB)
	// Menus
	menu.NewHandler(fiberApp, db.DB)
	// Uploads
	upload.NewHandler(fiberApp)
	// Tables
	table.NewHandler(fiberApp, db.DB)
	// Internal
	internalModule.NewHandler(fiberApp, db.DB)
	// Promotions
	promotion.NewHandler(fiberApp, db.DB)

	fiberApp.All("*", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(response.NotFound("Route not found", nil))
	})

	err = fiberApp.Listen(config.Config.Port)

	if err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
