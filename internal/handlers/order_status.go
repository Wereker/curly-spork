// internal/handlers/order_status.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/order"
	"app/internal/response"
)

type OrderStatusHandler struct {
	db *gorm.DB
}

func NewOrderStatusHandler(db *gorm.DB) *OrderStatusHandler {
	return &OrderStatusHandler{db: db}
}

// GetAllOrderStatuses godoc
// @Summary Получить список статусов заказов
// @Description Возвращает список всех статусов заказов
// @Tags OrderStatus
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по названию"
// @Success 200 {object} response.OrderStatusListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /order-statuses [get]
func (h *OrderStatusHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	offset := (page - 1) * limit

	query := h.db.Model(&models.OrderStatus{})

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество статусов заказов"},
		)
	}

	var statuses []models.OrderStatus
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&statuses).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статусы заказов"},
		)
	}

	data := make([]response.OrderStatusResponse, 0, len(statuses))
	for _, status := range statuses {
		data = append(data, response.OrderStatusResponse{
			ID:              status.ID,
			Title:           status.Title,
			BackgroundColor: status.BackgroundColor,
			CreatedAt:       status.CreatedAt,
			UpdatedAt:       status.UpdatedAt,
		})
	}

	return c.JSON(response.OrderStatusListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetOrderStatusByID godoc
// @Summary Получить статус заказа по ID
// @Description Возвращает статус заказа по ID
// @Tags OrderStatus
// @Param id path int true "OrderStatus ID"
// @Success 200 {object} response.OrderStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /order-statuses/{id} [get]
func (h *OrderStatusHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.OrderStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус заказа не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус заказа"},
		)
	}

	return c.JSON(response.OrderStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// CreateOrderStatus godoc
// @Summary Создать статус заказа
// @Description Создает новый статус заказа
// @Tags OrderStatus
// @Accept json
// @Produce json
// @Param data body response.OrderStatusCreate true "Данные для создания статуса заказа"
// @Success 201 {object} response.OrderStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /order-statuses [post]
func (h *OrderStatusHandler) Create(c *fiber.Ctx) error {
	var req response.OrderStatusCreate

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

	var existing models.OrderStatus
	if err := h.db.Where("title = ?", req.Title).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Статус заказа с таким названием уже существует"},
		)
	}

	status := models.OrderStatus{
		Title:           req.Title,
		BackgroundColor: req.BackgroundColor,
	}

	if err := h.db.Create(&status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать статус заказа"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(response.OrderStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// UpdateOrderStatus godoc
// @Summary Полностью обновить статус заказа
// @Description Обновляет все поля статуса заказа по ID (PUT)
// @Tags OrderStatus
// @Accept json
// @Produce json
// @Param id path int true "OrderStatus ID"
// @Param data body response.OrderStatusUpdate true "Обновляемые поля статуса заказа"
// @Success 200 {object} response.OrderStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /order-statuses/{id} [put]
func (h *OrderStatusHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.OrderStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус заказа не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус заказа"},
		)
	}

	var req response.OrderStatusUpdate
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
		var existing models.OrderStatus
		if err := h.db.Where("title = ? AND id != ?", req.Title, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Статус заказа с таким названием уже существует"},
			)
		}
	}

	status.Title = req.Title
	status.BackgroundColor = req.BackgroundColor

	if err := h.db.Save(&status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить статус заказа"},
		)
	}

	return c.JSON(response.OrderStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// PatchOrderStatus godoc
// @Summary Частично обновить статус заказа
// @Description Обновляет указанные поля статуса заказа по ID (PATCH)
// @Tags OrderStatus
// @Accept json
// @Produce json
// @Param id path int true "OrderStatus ID"
// @Param data body response.OrderStatusPatch true "Обновляемые поля статуса заказа"
// @Success 200 {object} response.OrderStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /order-statuses/{id} [patch]
func (h *OrderStatusHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.OrderStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус заказа не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус заказа"},
		)
	}

	var req response.OrderStatusPatch
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
			var existing models.OrderStatus
			if err := h.db.Where("title = ? AND id != ?", *req.Title, id).First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Статус заказа с таким названием уже существует"},
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
			response.ErrorResponse{Error: "Не удалось обновить статус заказа"},
		)
	}

	return c.JSON(response.OrderStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// DeleteOrderStatus godoc
// @Summary Удалить статус заказа
// @Description Удаляет статус заказа из БД по ID
// @Tags OrderStatus
// @Param id path int true "OrderStatus ID"
// @Success 204 "Статус заказа успешно удалён"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /order-statuses/{id} [delete]
func (h *OrderStatusHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.OrderStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус заказа не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус заказа"},
		)
	}

	var ordersCount int64
	if err := h.db.Model(&models.Order{}).Where("order_status_id = ?", id).Count(&ordersCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные заказы"},
		)
	}

	if ordersCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить статус заказа, так как с ним связаны заказы"},
		)
	}

	if err := h.db.Delete(&status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить статус заказа"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
