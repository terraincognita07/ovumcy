package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName: "Lume v0.1.0",
	})

	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("🌙 Lume is running!")
	})

	log.Println("Starting Lume on http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}
