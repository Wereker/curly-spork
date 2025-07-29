package handlers

import (
	"app/internal/models"
	"app/internal/response"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// GetMeasure godoc
// @Summary Получить список ед. измерения
// @Description Возвращает список ед. измерения
// @Tags Measure
// @Success 200 {array} response.MeasureResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /measures [get]
func GetMeasure(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var measures []models.Measure

		if result := db.Find(&measures); result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				response.ErrorResponse{Error: "Не удалось получить данные"},
			)
		}

		return c.Status(fiber.StatusOK).JSON(measures)
	}
}

// CreateMeasure godoc
// @Summary Создать ед. измерения
// @Description Создает ед. измерения в БД
// @Tags Measure
// @Accept json
// @Produce json
// @Param data body response.MeasureCreate true "Данные ед. измерения"
// @Success 201 {object} response.MeasureResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /measures [post]
func CreateMeasure(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var measure models.Measure

		if err := c.BodyParser(&measure); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				response.ErrorResponse{Error: "Неверный формат тела запроса"},
			)
		}

		if result := db.Create(&measure); result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				response.ErrorResponse{Error: "Не удалось создать запись"},
			)
		}

		return c.Status(fiber.StatusCreated).JSON(measure)
	}
}

// GetMeasureByID godoc
// @Summary Получить ед. измерения по ID
// @Description Возвращает ед. измерения по ID
// @Tags Measure
// @Param id path int true "Measure ID"
// @Success 200 {object} response.MeasureResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /measures/{id} [get]
func GetMeasureByID(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var measure models.Measure
		result := db.First(&measure, id)
		if result.Error != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Запись не найдена"},
			)
		}

		return c.JSON(measure)
	}
}

// UpdateMeasure godoc
// @Summary Обновить ед. измерения
// @Description Обновляет существующую ед. измерения по ID
// @Tags Measure
// @Accept json
// @Produce json
// @Param id path int true "Measure ID"
// @Param data body response.MeasureUpdate true "Обновляемые поля ед. измерения"
// @Success 200 {object} response.MeasureResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /measures/{id} [put]
func UpdateMeasure(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var measure models.Measure
		result := db.First(&measure, id)
		if result.Error != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Запись не найдена"},
			)
		}

		var updatedData models.Measure
		if err := c.BodyParser(&updatedData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				response.ErrorResponse{Error: "Неверный формат тела запроса"},
			)
		}

		measure.Name = updatedData.Name

		if err := db.Save(&measure).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				response.ErrorResponse{Error: "Ошибка при обновлении"},
			)
		}

		return c.JSON(measure)
	}
}

// DeleteMeasure godoc
// @Summary Удалить ед. измерения
// @Description Удаляет единицу измерения из БД по ID
// @Tags Measure
// @Param id path int true "Measure ID"
// @Success 204 "Ед. измерения успешна удалёна"
// @Failure 404 {object} response.ErrorResponse
// @Router /measures/{id} [delete]
func DeleteMeasure(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var measure models.Measure
		result := db.First(&measure, id)
		if result.Error != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Запись не найдена"},
			)
		}

		db.Delete(&measure)
		return c.SendStatus(fiber.StatusNoContent)
	}
}
