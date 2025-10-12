package main

import (
	"log"

	"eka-dev.cloud/master-data/config"
	"eka-dev.cloud/master-data/db"
	_ "eka-dev.cloud/master-data/db"
	_ "eka-dev.cloud/master-data/lib"
	"eka-dev.cloud/master-data/middleware"
	"eka-dev.cloud/master-data/modules/category"
	"eka-dev.cloud/master-data/modules/internalModule"
	"eka-dev.cloud/master-data/modules/menu"
	"eka-dev.cloud/master-data/modules/table"
	"eka-dev.cloud/master-data/modules/upload"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jmoiron/sqlx"
)

func main() {
	// Load env
	initiator()

	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Println("Error closing database connection:", err)
		}
	}(db.DB)

}

func initiator() {
	// Initialize the fiber app
	fiberApp := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	fiberApp.Use(logger.New(logger.Config{
		Format:     "[${time}] ${ip} ${method} ${path} - ${status} (${latency})\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Asia/Jakarta",
	}))

	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins: config.Config.AllowedOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Timestamp, X-Signature",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

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

	fiberApp.All("*", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(response.NotFound("Route not found", nil))
	})

	err := fiberApp.Listen(config.Config.Port)
	if err != nil {
		log.Fatalln("Failed to start server:", err)
		return
	}
}
