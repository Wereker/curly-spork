package routes

import (
	"app/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	app.Get("/swagger/*", swagger.HandlerDefault)

	productGroup := app.Group("/products")
	productGroup.Get("/", handlers.GetProduct(db))
	productGroup.Post("/", handlers.CreateProduct(db))
	productGroup.Get("/:id", handlers.GetProductByID(db))
	productGroup.Put("/:id", handlers.UpdateProduct(db))
	productGroup.Delete("/:id", handlers.DeleteProduct(db))

	measureGroup := app.Group("/measures")
	measureGroup.Get("/", handlers.GetMeasure(db))
	measureGroup.Post("/", handlers.CreateMeasure(db))
	measureGroup.Get("/:id", handlers.GetMeasureByID(db))
	measureGroup.Put("/:id", handlers.UpdateMeasure(db))
	measureGroup.Delete("/:id", handlers.DeleteMeasure(db))
}
