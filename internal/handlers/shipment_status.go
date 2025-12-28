package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/order"
	"app/internal/response"
)

// ShipmentStatusHandler структура для работы со статусами доставки
type ShipmentStatusHandler struct {
	db *gorm.DB
}

// NewShipmentStatusHandler создает новый обработчик
func NewShipmentStatusHandler(db *gorm.DB) *ShipmentStatusHandler {
	return &ShipmentStatusHandler{db: db}
}

// GetAllShipmentStatuses godoc
// @Summary Получить список статусов доставки
// @Description Возвращает список всех статусов доставки с пагинацией
// @Tags ShipmentStatus
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по названию"
// @Success 200 {object} response.ShipmentStatusListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipment-statuses [get]
func (h *ShipmentStatusHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")

	offset := (page - 1) * limit

	query := h.db.Model(&models.ShipmentStatus{})

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество статусов"},
		)
	}

	var statuses []models.ShipmentStatus
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&statuses).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статусы доставки"},
		)
	}

	data := make([]response.ShipmentStatusResponse, 0, len(statuses))
	for _, status := range statuses {
		data = append(data, response.ShipmentStatusResponse{
			ID:              status.ID,
			Title:           status.Title,
			BackgroundColor: status.BackgroundColor,
			CreatedAt:       status.CreatedAt,
			UpdatedAt:       status.UpdatedAt,
		})
	}

	return c.JSON(response.ShipmentStatusListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetShipmentStatusByID godoc
// @Summary Получить статус доставки по ID
// @Description Возвращает статус доставки по ID
// @Tags ShipmentStatus
// @Param id path int true "ShipmentStatus ID"
// @Success 200 {object} response.ShipmentStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /shipment-statuses/{id} [get]
func (h *ShipmentStatusHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.ShipmentStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус доставки не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус доставки"},
		)
	}

	return c.JSON(response.ShipmentStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// CreateShipmentStatus godoc
// @Summary Создать статус доставки
// @Description Создает новый статус доставки в системе
// @Tags ShipmentStatus
// @Accept json
// @Produce json
// @Param data body response.ShipmentStatusCreate true "Данные для создания статуса доставки"
// @Success 201 {object} response.ShipmentStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipment-statuses [post]
func (h *ShipmentStatusHandler) Create(c *fiber.Ctx) error {
	var req response.ShipmentStatusCreate

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

	var existing models.ShipmentStatus
	if err := h.db.Where("title = ?", req.Title).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Статус доставки с таким названием уже существует"},
		)
	}

	status := models.ShipmentStatus{
		Title:           req.Title,
		BackgroundColor: req.BackgroundColor,
	}

	if err := h.db.Create(&status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать статус доставки"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(response.ShipmentStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// UpdateShipmentStatus godoc
// @Summary Полностью обновить статус доставки
// @Description Обновляет все поля статуса доставки по ID (PUT)
// @Tags ShipmentStatus
// @Accept json
// @Produce json
// @Param id path int true "ShipmentStatus ID"
// @Param data body response.ShipmentStatusUpdate true "Обновляемые поля статуса доставки"
// @Success 200 {object} response.ShipmentStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipment-statuses/{id} [put]
func (h *ShipmentStatusHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.ShipmentStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус доставки не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус доставки"},
		)
	}

	var req response.ShipmentStatusUpdate
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

	if status.Title != req.Title {
		var existing models.ShipmentStatus
		if err := h.db.Where("title = ? AND id != ?", req.Title, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Статус доставки с таким названием уже существует"},
			)
		}
	}

	status.Title = req.Title
	status.BackgroundColor = req.BackgroundColor

	if err := h.db.Save(&status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить статус доставки"},
		)
	}

	return c.JSON(response.ShipmentStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// PatchShipmentStatus godoc
// @Summary Частично обновить статус доставки
// @Description Обновляет указанные поля статуса доставки по ID (PATCH)
// @Tags ShipmentStatus
// @Accept json
// @Produce json
// @Param id path int true "ShipmentStatus ID"
// @Param data body response.ShipmentStatusPatch true "Обновляемые поля статуса доставки"
// @Success 200 {object} response.ShipmentStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipment-statuses/{id} [patch]
func (h *ShipmentStatusHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.ShipmentStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус доставки не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус доставки"},
		)
	}

	var req response.ShipmentStatusPatch
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
		if status.Title != *req.Title {
			var existing models.ShipmentStatus
			if err := h.db.Where("title = ? AND id != ?", *req.Title, id).First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Статус доставки с таким названием уже существует"},
				)
			}
		}
		status.Title = *req.Title
	}

	if req.BackgroundColor != nil {
		status.BackgroundColor = *req.BackgroundColor
	}

	if err := h.db.Save(&status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить статус доставки"},
		)
	}

	return c.JSON(response.ShipmentStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// DeleteShipmentStatus godoc
// @Summary Удалить статус доставки
// @Description Удаляет статус доставки из БД по ID
// @Tags ShipmentStatus
// @Param id path int true "ShipmentStatus ID"
// @Success 204 "Статус доставки успешно удалён"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /shipment-statuses/{id} [delete]
func (h *ShipmentStatusHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.ShipmentStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус доставки не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус доставки"},
		)
	}

	var shipmentsCount int64
	if err := h.db.Model(&models.Shipment{}).Where("shipment_status_id = ?", id).Count(&shipmentsCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные записи"},
		)
	}

	if shipmentsCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить статус доставки, так как с ним связаны доставки"},
		)
	}

	if err := h.db.Delete(&status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить статус доставки"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
