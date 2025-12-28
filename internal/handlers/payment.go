// internal/handlers/payment.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/order"
	"app/internal/response"
)

type PaymentHandler struct {
	db *gorm.DB
}

func NewPaymentHandler(db *gorm.DB) *PaymentHandler {
	return &PaymentHandler{db: db}
}

// GetAllPayments godoc
// @Summary Получить список оплат
// @Description Возвращает список всех оплат
// @Tags Payment
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param payment_method_id query int false "Фильтр по способу оплаты"
// @Param payment_status_id query int false "Фильтр по статусу оплаты"
// @Param start_date query string false "Дата начала (YYYY-MM-DD)"
// @Param end_date query string false "Дата окончания (YYYY-MM-DD)"
// @Success 200 {object} response.PaymentListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payments [get]
func (h *PaymentHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	paymentMethodID, _ := strconv.ParseUint(c.Query("payment_method_id"), 10, 32)
	paymentStatusID, _ := strconv.ParseUint(c.Query("payment_status_id"), 10, 32)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	offset := (page - 1) * limit

	query := h.db.Model(&models.Payment{}).Preload("PaymentMethod").Preload("PaymentStatus")

	if paymentMethodID > 0 {
		query = query.Where("payment_method_id = ?", paymentMethodID)
	}
	if paymentStatusID > 0 {
		query = query.Where("payment_status_id = ?", paymentStatusID)
	}
	if startDate != "" {
		query = query.Where("payment_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("payment_date <= ?", endDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество оплат"},
		)
	}

	var payments []models.Payment
	if err := query.Order("payment_date DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить оплаты"},
		)
	}

	data := make([]response.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		resp := response.PaymentResponse{
			ID:              payment.ID,
			PaymentDate:     payment.PaymentDate,
			PaymentMethodID: payment.PaymentMethodID,
			PaymentStatusID: payment.PaymentStatusID,
			CreatedAt:       payment.CreatedAt,
			UpdatedAt:       payment.UpdatedAt,
		}

		if payment.PaymentMethod != nil {
			resp.PaymentMethod = &response.PaymentMethodResponse{
				ID:    payment.PaymentMethod.ID,
				Title: payment.PaymentMethod.Title,
			}
		}

		if payment.PaymentStatus != nil {
			resp.PaymentStatus = &response.PaymentStatusResponse{
				ID:              payment.PaymentStatus.ID,
				Title:           payment.PaymentStatus.Title,
				BackgroundColor: payment.PaymentStatus.BackgroundColor,
			}
		}

		data = append(data, resp)
	}

	return c.JSON(response.PaymentListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetPaymentByID godoc
// @Summary Получить оплату по ID
// @Description Возвращает оплату по ID
// @Tags Payment
// @Param id path int true "Payment ID"
// @Success 200 {object} response.PaymentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /payments/{id} [get]
func (h *PaymentHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var payment models.Payment
	if err := h.db.Preload("PaymentMethod").Preload("PaymentStatus").
		First(&payment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Оплата не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить оплату"},
		)
	}

	resp := response.PaymentResponse{
		ID:              payment.ID,
		PaymentDate:     payment.PaymentDate,
		PaymentMethodID: payment.PaymentMethodID,
		PaymentStatusID: payment.PaymentStatusID,
		CreatedAt:       payment.CreatedAt,
		UpdatedAt:       payment.UpdatedAt,
	}

	if payment.PaymentMethod != nil {
		resp.PaymentMethod = &response.PaymentMethodResponse{
			ID:    payment.PaymentMethod.ID,
			Title: payment.PaymentMethod.Title,
		}
	}

	if payment.PaymentStatus != nil {
		resp.PaymentStatus = &response.PaymentStatusResponse{
			ID:              payment.PaymentStatus.ID,
			Title:           payment.PaymentStatus.Title,
			BackgroundColor: payment.PaymentStatus.BackgroundColor,
		}
	}

	return c.JSON(resp)
}

// CreatePayment godoc
// @Summary Создать оплату
// @Description Создает новую оплату
// @Tags Payment
// @Accept json
// @Produce json
// @Param data body response.PaymentCreate true "Данные для создания оплаты"
// @Success 201 {object} response.PaymentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payments [post]
func (h *PaymentHandler) Create(c *fiber.Ctx) error {
	var req response.PaymentCreate

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

	var paymentMethod models.PaymentMethod
	if err := h.db.First(&paymentMethod, req.PaymentMethodID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Способ оплаты не найден"},
		)
	}

	var paymentStatus models.PaymentStatus
	if err := h.db.First(&paymentStatus, req.PaymentStatusID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Статус оплаты не найден"},
		)
	}

	payment := models.Payment{
		PaymentMethodID: req.PaymentMethodID,
		PaymentStatusID: req.PaymentStatusID,
	}

	if err := h.db.Create(&payment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать оплату"},
		)
	}

	h.db.Preload("PaymentMethod").Preload("PaymentStatus").First(&payment, payment.ID)

	resp := response.PaymentResponse{
		ID:              payment.ID,
		PaymentDate:     payment.PaymentDate,
		PaymentMethodID: payment.PaymentMethodID,
		PaymentStatusID: payment.PaymentStatusID,
		CreatedAt:       payment.CreatedAt,
		UpdatedAt:       payment.UpdatedAt,
	}

	if payment.PaymentMethod != nil {
		resp.PaymentMethod = &response.PaymentMethodResponse{
			ID:    payment.PaymentMethod.ID,
			Title: payment.PaymentMethod.Title,
		}
	}

	if payment.PaymentStatus != nil {
		resp.PaymentStatus = &response.PaymentStatusResponse{
			ID:              payment.PaymentStatus.ID,
			Title:           payment.PaymentStatus.Title,
			BackgroundColor: payment.PaymentStatus.BackgroundColor,
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdatePayment godoc
// @Summary Полностью обновить оплату
// @Description Обновляет все поля оплаты по ID (PUT)
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param data body response.PaymentUpdate true "Обновляемые поля оплаты"
// @Success 200 {object} response.PaymentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payments/{id} [put]
func (h *PaymentHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var payment models.Payment
	if err := h.db.First(&payment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Оплата не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить оплату"},
		)
	}

	var req response.PaymentUpdate
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

	var paymentMethod models.PaymentMethod
	if err := h.db.First(&paymentMethod, req.PaymentMethodID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Способ оплаты не найден"},
		)
	}

	var paymentStatus models.PaymentStatus
	if err := h.db.First(&paymentStatus, req.PaymentStatusID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Статус оплаты не найден"},
		)
	}

	payment.PaymentMethodID = req.PaymentMethodID
	payment.PaymentStatusID = req.PaymentStatusID

	if err := h.db.Save(&payment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить оплату"},
		)
	}

	h.db.Preload("PaymentMethod").Preload("PaymentStatus").First(&payment, payment.ID)

	resp := response.PaymentResponse{
		ID:              payment.ID,
		PaymentDate:     payment.PaymentDate,
		PaymentMethodID: payment.PaymentMethodID,
		PaymentStatusID: payment.PaymentStatusID,
		CreatedAt:       payment.CreatedAt,
		UpdatedAt:       payment.UpdatedAt,
	}

	if payment.PaymentMethod != nil {
		resp.PaymentMethod = &response.PaymentMethodResponse{
			ID:    payment.PaymentMethod.ID,
			Title: payment.PaymentMethod.Title,
		}
	}

	if payment.PaymentStatus != nil {
		resp.PaymentStatus = &response.PaymentStatusResponse{
			ID:              payment.PaymentStatus.ID,
			Title:           payment.PaymentStatus.Title,
			BackgroundColor: payment.PaymentStatus.BackgroundColor,
		}
	}

	return c.JSON(resp)
}

// PatchPayment godoc
// @Summary Частично обновить оплату
// @Description Обновляет указанные поля оплаты по ID (PATCH)
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param data body response.PaymentPatch true "Обновляемые поля оплаты"
// @Success 200 {object} response.PaymentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payments/{id} [patch]
func (h *PaymentHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var payment models.Payment
	if err := h.db.First(&payment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Оплата не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить оплату"},
		)
	}

	var req response.PaymentPatch
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

	if req.PaymentMethodID != nil {
		var paymentMethod models.PaymentMethod
		if err := h.db.First(&paymentMethod, *req.PaymentMethodID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Способ оплаты не найден"},
			)
		}
		payment.PaymentMethodID = *req.PaymentMethodID
	}
	if req.PaymentStatusID != nil {
		var paymentStatus models.PaymentStatus
		if err := h.db.First(&paymentStatus, *req.PaymentStatusID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус оплаты не найден"},
			)
		}
		payment.PaymentStatusID = *req.PaymentStatusID
	}

	if err := h.db.Save(&payment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить оплату"},
		)
	}

	h.db.Preload("PaymentMethod").Preload("PaymentStatus").First(&payment, payment.ID)

	resp := response.PaymentResponse{
		ID:              payment.ID,
		PaymentDate:     payment.PaymentDate,
		PaymentMethodID: payment.PaymentMethodID,
		PaymentStatusID: payment.PaymentStatusID,
		CreatedAt:       payment.CreatedAt,
		UpdatedAt:       payment.UpdatedAt,
	}

	if payment.PaymentMethod != nil {
		resp.PaymentMethod = &response.PaymentMethodResponse{
			ID:    payment.PaymentMethod.ID,
			Title: payment.PaymentMethod.Title,
		}
	}

	if payment.PaymentStatus != nil {
		resp.PaymentStatus = &response.PaymentStatusResponse{
			ID:              payment.PaymentStatus.ID,
			Title:           payment.PaymentStatus.Title,
			BackgroundColor: payment.PaymentStatus.BackgroundColor,
		}
	}

	return c.JSON(resp)
}

// DeletePayment godoc
// @Summary Удалить оплату
// @Description Удаляет оплату из БД по ID
// @Tags Payment
// @Param id path int true "Payment ID"
// @Success 204 "Оплата успешно удалена"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payments/{id} [delete]
func (h *PaymentHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var payment models.Payment
	if err := h.db.First(&payment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Оплата не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить оплату"},
		)
	}

	var ordersCount int64
	if err := h.db.Model(&models.Order{}).Where("payment_id = ?", id).Count(&ordersCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные заказы"},
		)
	}

	if ordersCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить оплату, так как с ней связаны заказы"},
		)
	}

	if err := h.db.Delete(&payment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить оплату"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
