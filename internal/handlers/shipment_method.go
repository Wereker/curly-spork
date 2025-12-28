// internal/handlers/shipment_method.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/order"
	"app/internal/response"
)

type ShipmentMethodHandler struct {
	db *gorm.DB
}

func NewShipmentMethodHandler(db *gorm.DB) *ShipmentMethodHandler {
	return &ShipmentMethodHandler{db: db}
}

// GetAllShipmentMethods godoc
// @Summary Получить список способов доставки
// @Description Возвращает список всех способов доставки
// @Tags ShipmentMethod
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по названию"
// @Success 200 {object} response.ShipmentMethodListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipment-methods [get]
func (h *ShipmentMethodHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	offset := (page - 1) * limit

	query := h.db.Model(&models.ShipmentMethod{})

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество способов доставки"},
		)
	}

	var methods []models.ShipmentMethod
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&methods).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить способы доставки"},
		)
	}

	data := make([]response.ShipmentMethodResponse, 0, len(methods))
	for _, method := range methods {
		data = append(data, response.ShipmentMethodResponse{
			ID:          method.ID,
			Title:       method.Title,
			Description: method.Description,
			CreatedAt:   method.CreatedAt,
			UpdatedAt:   method.UpdatedAt,
		})
	}

	return c.JSON(response.ShipmentMethodListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetShipmentMethodByID godoc
// @Summary Получить способ доставки по ID
// @Description Возвращает способ доставки по ID
// @Tags ShipmentMethod
// @Param id path int true "ShipmentMethod ID"
// @Success 200 {object} response.ShipmentMethodResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /shipment-methods/{id} [get]
func (h *ShipmentMethodHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var method models.ShipmentMethod
	if err := h.db.First(&method, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Способ доставки не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить способ доставки"},
		)
	}

	return c.JSON(response.ShipmentMethodResponse{
		ID:          method.ID,
		Title:       method.Title,
		Description: method.Description,
		CreatedAt:   method.CreatedAt,
		UpdatedAt:   method.UpdatedAt,
	})
}

// CreateShipmentMethod godoc
// @Summary Создать способ доставки
// @Description Создает новый способ доставки
// @Tags ShipmentMethod
// @Accept json
// @Produce json
// @Param data body response.ShipmentMethodCreate true "Данные для создания способа доставки"
// @Success 201 {object} response.ShipmentMethodResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipment-methods [post]
func (h *ShipmentMethodHandler) Create(c *fiber.Ctx) error {
	var req response.ShipmentMethodCreate

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат тела запроса"},
		)
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: err.Error()},
		)
	}

	var existing models.ShipmentMethod
	if err := h.db.Where("title = ?", req.Title).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Способ доставки с таким названием уже существует"},
		)
	}

	method := models.ShipmentMethod{
		Title:       req.Title,
		Description: req.Description,
	}

	if err := h.db.Create(&method).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать способ доставки"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(response.ShipmentMethodResponse{
		ID:          method.ID,
		Title:       method.Title,
		Description: method.Description,
		CreatedAt:   method.CreatedAt,
		UpdatedAt:   method.UpdatedAt,
	})
}

// UpdateShipmentMethod godoc
// @Summary Полностью обновить способ доставки
// @Description Обновляет все поля способа доставки по ID (PUT)
// @Tags ShipmentMethod
// @Accept json
// @Produce json
// @Param id path int true "ShipmentMethod ID"
// @Param data body response.ShipmentMethodUpdate true "Обновляемые поля способа доставки"
// @Success 200 {object} response.ShipmentMethodResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipment-methods/{id} [put]
func (h *ShipmentMethodHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var method models.ShipmentMethod
	if err := h.db.First(&method, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Способ доставки не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить способ доставки"},
		)
	}

	var req response.ShipmentMethodUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат тела запроса"},
		)
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: err.Error()},
		)
	}

	if method.Title != req.Title {
		var existing models.ShipmentMethod
		if err := h.db.Where("title = ? AND id != ?", req.Title, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Способ доставки с таким названием уже существует"},
			)
		}
	}

	method.Title = req.Title
	method.Description = req.Description

	if err := h.db.Save(&method).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить способ доставки"},
		)
	}

	return c.JSON(response.ShipmentMethodResponse{
		ID:          method.ID,
		Title:       method.Title,
		Description: method.Description,
		CreatedAt:   method.CreatedAt,
		UpdatedAt:   method.UpdatedAt,
	})
}

// PatchShipmentMethod godoc
// @Summary Частично обновить способ доставки
// @Description Обновляет указанные поля способа доставки по ID (PATCH)
// @Tags ShipmentMethod
// @Accept json
// @Produce json
// @Param id path int true "ShipmentMethod ID"
// @Param data body response.ShipmentMethodPatch true "Обновляемые поля способа доставки"
// @Success 200 {object} response.ShipmentMethodResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipment-methods/{id} [patch]
func (h *ShipmentMethodHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var method models.ShipmentMethod
	if err := h.db.First(&method, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Способ доставки не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить способ доставки"},
		)
	}

	var req response.ShipmentMethodPatch
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат тела запроса"},
		)
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: err.Error()},
		)
	}

	if req.Title != nil {
		if method.Title != *req.Title {
			var existing models.ShipmentMethod
			if err := h.db.Where("title = ? AND id != ?", *req.Title, id).First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Способ доставки с таким названием уже существует"},
				)
			}
		}
		method.Title = *req.Title
	}

	if req.Description != nil {
		method.Description = req.Description
	}

	if err := h.db.Save(&method).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить способ доставки"},
		)
	}

	return c.JSON(response.ShipmentMethodResponse{
		ID:          method.ID,
		Title:       method.Title,
		Description: method.Description,
		CreatedAt:   method.CreatedAt,
		UpdatedAt:   method.UpdatedAt,
	})
}

// DeleteShipmentMethod godoc
// @Summary Удалить способ доставки
// @Description Удаляет способ доставки из БД по ID
// @Tags ShipmentMethod
// @Param id path int true "ShipmentMethod ID"
// @Success 204 "Способ доставки успешно удалён"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipment-methods/{id} [delete]
func (h *ShipmentMethodHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var method models.ShipmentMethod
	if err := h.db.First(&method, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Способ доставки не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить способ доставки"},
		)
	}

	var shipmentsCount int64
	if err := h.db.Model(&models.Shipment{}).Where("shipment_method_id = ?", id).Count(&shipmentsCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные записи"},
		)
	}

	if shipmentsCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить способ доставки, так как с ним связаны доставки"},
		)
	}

	if err := h.db.Delete(&method).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить способ доставки"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
