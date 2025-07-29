package handlers

import (
	"app/internal/models"
	"app/internal/response"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// GetProduct godoc
// @Summary Получить список продуктов
// @Description Возвращает список продуктов
// @Tags Products
// @Success 200 {array} response.ProductResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products [get]
func GetProduct(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var products []models.Product
		db.Find(&products)
		return c.JSON(&products)
	}
}

// CreateProduct godoc
// @Summary Создать продукт
// @Description Создает продукт в БД
// @Tags Products
// @Accept json
// @Produce json
// @Param data body response.ProductCreate true "Данные продукта"
// @Success 201 {object} response.ProductResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products [post]
func CreateProduct(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var product models.Product

		if err := c.BodyParser(&product); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				response.ErrorResponse{Error: "Неверный формат тела запроса"},
			)
		}

		if result := db.Create(&product); result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				response.ErrorResponse{Error: "Не удалось создать запись"},
			)
		}

		return c.Status(fiber.StatusCreated).JSON(product)
	}
}

// GetProductByID godoc
// @Summary Получить продукт по ID
// @Description Возвращает продукт по ID
// @Tags Products
// @Param id path int true "Product ID"
// @Success 200 {object} response.ProductResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /products/{id} [get]
func GetProductByID(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var product models.Product
		result := db.First(&product, id)
		if result.Error != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Продукт не найдет"},
			)
		}

		return c.JSON(product)
	}
}

// UpdateProduct godoc
// @Summary Обновить продукт
// @Description Обновляет существующий продукт по ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param data body response.ProductUpdate true "Обновляемые поля продукта"
// @Success 200 {object} response.ProductResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products/{id} [put]
func UpdateProduct(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var product models.Product
		result := db.First(&product, id)
		if result.Error != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Запись не найдена"},
			)
		}

		var updateData map[string]interface{}
		if err := c.BodyParser(&updateData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				response.ErrorResponse{Error: "Неверный формат тела запроса"},
			)
		}

		if err := db.Model(&product).Updates(updateData).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				response.ErrorResponse{Error: "Ошибка при обновлении"},
			)
		}

		return c.JSON(product)
	}
}

// DeleteProduct godoc
// @Summary Удалить продукт
// @Description Удаляет продукт из БД по ID
// @Tags Products
// @Param id path int true "Product ID"
// @Success 204 "Продукт успешно удалён"
// @Failure 404 {object} response.ErrorResponse
// @Router /products/{id} [delete]
func DeleteProduct(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var product models.Product
		result := db.First(&product, id)
		if result.Error != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Запись не найдена"},
			)
		}

		db.Delete(&product)
		return c.SendStatus(fiber.StatusNoContent)
	}
}
