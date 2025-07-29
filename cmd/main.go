package main

import (
	"app/internal/db"
	"app/internal/routes"

	_ "app/docs"

	"github.com/gofiber/fiber/v2"
)

// @title Test API
// @version 1.0
func main() {
	database := db.Connect()
	app := fiber.New()

	routes.Setup(app, database)

	app.Listen(":3000")
}
