package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/order"
	"app/internal/response"
)

type PaymentStatusHandler struct {
	db *gorm.DB
}

func NewPaymentStatusHandler(db *gorm.DB) *PaymentStatusHandler {
	return &PaymentStatusHandler{db: db}
}

// GetAllPaymentStatuses godoc
// @Summary Получить список статусов оплаты
// @Description Возвращает список всех статусов оплаты
// @Tags PaymentStatus
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по названию"
// @Success 200 {object} response.PaymentStatusListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payment-statuses [get]
func (h *PaymentStatusHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	offset := (page - 1) * limit

	query := h.db.Model(&models.PaymentStatus{})

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество статусов оплаты"},
		)
	}

	var statuses []models.PaymentStatus
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&statuses).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статусы оплаты"},
		)
	}

	data := make([]response.PaymentStatusResponse, 0, len(statuses))
	for _, status := range statuses {
		data = append(data, response.PaymentStatusResponse{
			ID:              status.ID,
			Title:           status.Title,
			BackgroundColor: status.BackgroundColor,
			CreatedAt:       status.CreatedAt,
			UpdatedAt:       status.UpdatedAt,
		})
	}

	return c.JSON(response.PaymentStatusListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetPaymentStatusByID godoc
// @Summary Получить статус оплаты по ID
// @Description Возвращает статус оплаты по ID
// @Tags PaymentStatus
// @Param id path int true "PaymentStatus ID"
// @Success 200 {object} response.PaymentStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /payment-statuses/{id} [get]
func (h *PaymentStatusHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.PaymentStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус оплаты не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус оплаты"},
		)
	}

	return c.JSON(response.PaymentStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// CreatePaymentStatus godoc
// @Summary Создать статус оплаты
// @Description Создает новый статус оплаты
// @Tags PaymentStatus
// @Accept json
// @Produce json
// @Param data body response.PaymentStatusCreate true "Данные для создания статуса оплаты"
// @Success 201 {object} response.PaymentStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payment-statuses [post]
func (h *PaymentStatusHandler) Create(c *fiber.Ctx) error {
	var req response.PaymentStatusCreate

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

	var existing models.PaymentStatus
	if err := h.db.Where("title = ?", req.Title).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Статус оплаты с таким названием уже существует"},
		)
	}

	status := models.PaymentStatus{
		Title:           req.Title,
		BackgroundColor: req.BackgroundColor,
	}

	if err := h.db.Create(&status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать статус оплаты"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(response.PaymentStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// UpdatePaymentStatus godoc
// @Summary Полностью обновить статус оплаты
// @Description Обновляет все поля статуса оплаты по ID (PUT)
// @Tags PaymentStatus
// @Accept json
// @Produce json
// @Param id path int true "PaymentStatus ID"
// @Param data body response.PaymentStatusUpdate true "Обновляемые поля статуса оплаты"
// @Success 200 {object} response.PaymentStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payment-statuses/{id} [put]
func (h *PaymentStatusHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.PaymentStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус оплаты не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус оплаты"},
		)
	}

	var req response.PaymentStatusUpdate
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
		var existing models.PaymentStatus
		if err := h.db.Where("title = ? AND id != ?", req.Title, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Статус оплаты с таким названием уже существует"},
			)
		}
	}

	status.Title = req.Title
	status.BackgroundColor = req.BackgroundColor

	if err := h.db.Save(&status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить статус оплаты"},
		)
	}

	return c.JSON(response.PaymentStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// PatchPaymentStatus godoc
// @Summary Частично обновить статус оплаты
// @Description Обновляет указанные поля статуса оплаты по ID (PATCH)
// @Tags PaymentStatus
// @Accept json
// @Produce json
// @Param id path int true "PaymentStatus ID"
// @Param data body response.PaymentStatusPatch true "Обновляемые поля статуса оплаты"
// @Success 200 {object} response.PaymentStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payment-statuses/{id} [patch]
func (h *PaymentStatusHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.PaymentStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус оплаты не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус оплаты"},
		)
	}

	var req response.PaymentStatusPatch
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
			var existing models.PaymentStatus
			if err := h.db.Where("title = ? AND id != ?", *req.Title, id).First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Статус оплаты с таким названием уже существует"},
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
			response.ErrorResponse{Error: "Не удалось обновить статус оплаты"},
		)
	}

	return c.JSON(response.PaymentStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// DeletePaymentStatus godoc
// @Summary Удалить статус оплаты
// @Description Удаляет статус оплаты из БД по ID
// @Tags PaymentStatus
// @Param id path int true "PaymentStatus ID"
// @Success 204 "Статус оплаты успешно удалён"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payment-statuses/{id} [delete]
func (h *PaymentStatusHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.PaymentStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус оплаты не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус оплаты"},
		)
	}

	var paymentsCount int64
	if err := h.db.Model(&models.Payment{}).Where("payment_status_id = ?", id).Count(&paymentsCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные оплаты"},
		)
	}

	if paymentsCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить статус оплаты, так как с ним связаны оплаты"},
		)
	}

	if err := h.db.Delete(&status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить статус оплаты"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
