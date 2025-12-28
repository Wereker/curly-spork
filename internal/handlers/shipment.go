// internal/handlers/shipment.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/order"
	"app/internal/response"
)

type ShipmentHandler struct {
	db *gorm.DB
}

func NewShipmentHandler(db *gorm.DB) *ShipmentHandler {
	return &ShipmentHandler{db: db}
}

// GetAllShipments godoc
// @Summary Получить список доставок
// @Description Возвращает список всех доставок
// @Tags Shipment
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param number query string false "Поиск по номеру"
// @Param shipment_method_id query int false "Фильтр по способу доставки"
// @Param shipment_status_id query int false "Фильтр по статусу доставки"
// @Success 200 {object} response.ShipmentListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipments [get]
func (h *ShipmentHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	number := c.Query("number")
	shipmentMethodID, _ := strconv.ParseUint(c.Query("shipment_method_id"), 10, 32)
	shipmentStatusID, _ := strconv.ParseUint(c.Query("shipment_status_id"), 10, 32)
	offset := (page - 1) * limit

	query := h.db.Model(&models.Shipment{}).Preload("ShipmentMethod").Preload("ShipmentStatus")

	if number != "" {
		query = query.Where("number ILIKE ?", "%"+number+"%")
	}
	if shipmentMethodID > 0 {
		query = query.Where("shipment_method_id = ?", shipmentMethodID)
	}
	if shipmentStatusID > 0 {
		query = query.Where("shipment_status_id = ?", shipmentStatusID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество доставок"},
		)
	}

	var shipments []models.Shipment
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&shipments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить доставки"},
		)
	}

	data := make([]response.ShipmentResponse, 0, len(shipments))
	for _, shipment := range shipments {
		resp := response.ShipmentResponse{
			ID:               shipment.ID,
			Number:           shipment.Number,
			Price:            shipment.Price,
			Address:          shipment.Address,
			ShipmentMethodID: shipment.ShipmentMethodID,
			ShipmentStatusID: shipment.ShipmentStatusID,
			CreatedAt:        shipment.CreatedAt,
			UpdatedAt:        shipment.UpdatedAt,
		}

		if shipment.ShipmentMethod != nil {
			resp.ShipmentMethod = &response.ShipmentMethodResponse{
				ID:    shipment.ShipmentMethod.ID,
				Title: shipment.ShipmentMethod.Title,
			}
		}

		if shipment.ShipmentStatus != nil {
			resp.ShipmentStatus = &response.ShipmentStatusResponse{
				ID:              shipment.ShipmentStatus.ID,
				Title:           shipment.ShipmentStatus.Title,
				BackgroundColor: shipment.ShipmentStatus.BackgroundColor,
			}
		}

		data = append(data, resp)
	}

	return c.JSON(response.ShipmentListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetShipmentByID godoc
// @Summary Получить доставку по ID
// @Description Возвращает доставку по ID
// @Tags Shipment
// @Param id path int true "Shipment ID"
// @Success 200 {object} response.ShipmentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /shipments/{id} [get]
func (h *ShipmentHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var shipment models.Shipment
	if err := h.db.Preload("ShipmentMethod").Preload("ShipmentStatus").
		First(&shipment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Доставка не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить доставку"},
		)
	}

	resp := response.ShipmentResponse{
		ID:               shipment.ID,
		Number:           shipment.Number,
		Price:            shipment.Price,
		Address:          shipment.Address,
		ShipmentMethodID: shipment.ShipmentMethodID,
		ShipmentStatusID: shipment.ShipmentStatusID,
		CreatedAt:        shipment.CreatedAt,
		UpdatedAt:        shipment.UpdatedAt,
	}

	if shipment.ShipmentMethod != nil {
		resp.ShipmentMethod = &response.ShipmentMethodResponse{
			ID:    shipment.ShipmentMethod.ID,
			Title: shipment.ShipmentMethod.Title,
		}
	}

	if shipment.ShipmentStatus != nil {
		resp.ShipmentStatus = &response.ShipmentStatusResponse{
			ID:              shipment.ShipmentStatus.ID,
			Title:           shipment.ShipmentStatus.Title,
			BackgroundColor: shipment.ShipmentStatus.BackgroundColor,
		}
	}

	return c.JSON(resp)
}

// CreateShipment godoc
// @Summary Создать доставку
// @Description Создает новую доставку
// @Tags Shipment
// @Accept json
// @Produce json
// @Param data body response.ShipmentCreate true "Данные для создания доставки"
// @Success 201 {object} response.ShipmentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipments [post]
func (h *ShipmentHandler) Create(c *fiber.Ctx) error {
	var req response.ShipmentCreate

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

	var shipmentMethod models.ShipmentMethod
	if err := h.db.First(&shipmentMethod, req.ShipmentMethodID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Способ доставки не найден"},
		)
	}

	var shipmentStatus models.ShipmentStatus
	if err := h.db.First(&shipmentStatus, req.ShipmentStatusID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Статус доставки не найден"},
		)
	}

	shipment := models.Shipment{
		Price:            req.Price,
		Address:          req.Address,
		ShipmentMethodID: req.ShipmentMethodID,
		ShipmentStatusID: req.ShipmentStatusID,
	}

	if err := h.db.Create(&shipment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать доставку"},
		)
	}

	h.db.Preload("ShipmentMethod").Preload("ShipmentStatus").First(&shipment, shipment.ID)

	resp := response.ShipmentResponse{
		ID:               shipment.ID,
		Number:           shipment.Number,
		Price:            shipment.Price,
		Address:          shipment.Address,
		ShipmentMethodID: shipment.ShipmentMethodID,
		ShipmentStatusID: shipment.ShipmentStatusID,
		CreatedAt:        shipment.CreatedAt,
		UpdatedAt:        shipment.UpdatedAt,
	}

	if shipment.ShipmentMethod != nil {
		resp.ShipmentMethod = &response.ShipmentMethodResponse{
			ID:    shipment.ShipmentMethod.ID,
			Title: shipment.ShipmentMethod.Title,
		}
	}

	if shipment.ShipmentStatus != nil {
		resp.ShipmentStatus = &response.ShipmentStatusResponse{
			ID:              shipment.ShipmentStatus.ID,
			Title:           shipment.ShipmentStatus.Title,
			BackgroundColor: shipment.ShipmentStatus.BackgroundColor,
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateShipment godoc
// @Summary Полностью обновить доставку
// @Description Обновляет все поля доставки по ID (PUT)
// @Tags Shipment
// @Accept json
// @Produce json
// @Param id path int true "Shipment ID"
// @Param data body response.ShipmentUpdate true "Обновляемые поля доставки"
// @Success 200 {object} response.ShipmentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipments/{id} [put]
func (h *ShipmentHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var shipment models.Shipment
	if err := h.db.First(&shipment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Доставка не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить доставку"},
		)
	}

	var req response.ShipmentUpdate
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

	var shipmentMethod models.ShipmentMethod
	if err := h.db.First(&shipmentMethod, req.ShipmentMethodID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Способ доставки не найден"},
		)
	}

	var shipmentStatus models.ShipmentStatus
	if err := h.db.First(&shipmentStatus, req.ShipmentStatusID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Статус доставки не найден"},
		)
	}

	shipment.Price = req.Price
	shipment.Address = req.Address
	shipment.ShipmentMethodID = req.ShipmentMethodID
	shipment.ShipmentStatusID = req.ShipmentStatusID

	if err := h.db.Save(&shipment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить доставку"},
		)
	}

	h.db.Preload("ShipmentMethod").Preload("ShipmentStatus").First(&shipment, shipment.ID)

	resp := response.ShipmentResponse{
		ID:               shipment.ID,
		Number:           shipment.Number,
		Price:            shipment.Price,
		Address:          shipment.Address,
		ShipmentMethodID: shipment.ShipmentMethodID,
		ShipmentStatusID: shipment.ShipmentStatusID,
		CreatedAt:        shipment.CreatedAt,
		UpdatedAt:        shipment.UpdatedAt,
	}

	if shipment.ShipmentMethod != nil {
		resp.ShipmentMethod = &response.ShipmentMethodResponse{
			ID:    shipment.ShipmentMethod.ID,
			Title: shipment.ShipmentMethod.Title,
		}
	}

	if shipment.ShipmentStatus != nil {
		resp.ShipmentStatus = &response.ShipmentStatusResponse{
			ID:              shipment.ShipmentStatus.ID,
			Title:           shipment.ShipmentStatus.Title,
			BackgroundColor: shipment.ShipmentStatus.BackgroundColor,
		}
	}

	return c.JSON(resp)
}

// PatchShipment godoc
// @Summary Частично обновить доставку
// @Description Обновляет указанные поля доставки по ID (PATCH)
// @Tags Shipment
// @Accept json
// @Produce json
// @Param id path int true "Shipment ID"
// @Param data body response.ShipmentPatch true "Обновляемые поля доставки"
// @Success 200 {object} response.ShipmentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipments/{id} [patch]
func (h *ShipmentHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var shipment models.Shipment
	if err := h.db.First(&shipment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Доставка не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить доставку"},
		)
	}

	var req response.ShipmentPatch
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

	if req.Price != nil {
		shipment.Price = *req.Price
	}
	if req.Address != nil {
		shipment.Address = *req.Address
	}
	if req.ShipmentMethodID != nil {
		var shipmentMethod models.ShipmentMethod
		if err := h.db.First(&shipmentMethod, *req.ShipmentMethodID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Способ доставки не найден"},
			)
		}
		shipment.ShipmentMethodID = *req.ShipmentMethodID
	}
	if req.ShipmentStatusID != nil {
		var shipmentStatus models.ShipmentStatus
		if err := h.db.First(&shipmentStatus, *req.ShipmentStatusID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус доставки не найден"},
			)
		}
		shipment.ShipmentStatusID = *req.ShipmentStatusID
	}

	if err := h.db.Save(&shipment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить доставку"},
		)
	}

	h.db.Preload("ShipmentMethod").Preload("ShipmentStatus").First(&shipment, shipment.ID)

	resp := response.ShipmentResponse{
		ID:               shipment.ID,
		Number:           shipment.Number,
		Price:            shipment.Price,
		Address:          shipment.Address,
		ShipmentMethodID: shipment.ShipmentMethodID,
		ShipmentStatusID: shipment.ShipmentStatusID,
		CreatedAt:        shipment.CreatedAt,
		UpdatedAt:        shipment.UpdatedAt,
	}

	if shipment.ShipmentMethod != nil {
		resp.ShipmentMethod = &response.ShipmentMethodResponse{
			ID:    shipment.ShipmentMethod.ID,
			Title: shipment.ShipmentMethod.Title,
		}
	}

	if shipment.ShipmentStatus != nil {
		resp.ShipmentStatus = &response.ShipmentStatusResponse{
			ID:              shipment.ShipmentStatus.ID,
			Title:           shipment.ShipmentStatus.Title,
			BackgroundColor: shipment.ShipmentStatus.BackgroundColor,
		}
	}

	return c.JSON(resp)
}

// DeleteShipment godoc
// @Summary Удалить доставку
// @Description Удаляет доставку из БД по ID
// @Tags Shipment
// @Param id path int true "Shipment ID"
// @Success 204 "Доставка успешно удалена"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipments/{id} [delete]
func (h *ShipmentHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var shipment models.Shipment
	if err := h.db.First(&shipment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Доставка не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить доставку"},
		)
	}

	var ordersCount int64
	if err := h.db.Model(&models.Order{}).Where("shipment_id = ?", id).Count(&ordersCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные заказы"},
		)
	}

	if ordersCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить доставку, так как с ней связаны заказы"},
		)
	}

	if err := h.db.Delete(&shipment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить доставку"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
