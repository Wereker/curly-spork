package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/order"
	"app/internal/response"
)

type PaymentMethodHandler struct {
	db *gorm.DB
}

func NewPaymentMethodHandler(db *gorm.DB) *PaymentMethodHandler {
	return &PaymentMethodHandler{db: db}
}

// GetAllPaymentMethods godoc
// @Summary Получить список способов оплаты
// @Description Возвращает список всех способов оплаты
// @Tags PaymentMethod
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по названию"
// @Success 200 {object} response.PaymentMethodListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payment-methods [get]
func (h *PaymentMethodHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	offset := (page - 1) * limit

	query := h.db.Model(&models.PaymentMethod{})

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество способов оплаты"},
		)
	}

	var methods []models.PaymentMethod
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&methods).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить способы оплаты"},
		)
	}

	data := make([]response.PaymentMethodResponse, 0, len(methods))
	for _, method := range methods {
		data = append(data, response.PaymentMethodResponse{
			ID:          method.ID,
			Title:       method.Title,
			Description: method.Description,
			CreatedAt:   method.CreatedAt,
			UpdatedAt:   method.UpdatedAt,
		})
	}

	return c.JSON(response.PaymentMethodListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetPaymentMethodByID godoc
// @Summary Получить способ оплаты по ID
// @Description Возвращает способ оплаты по ID
// @Tags PaymentMethod
// @Param id path int true "PaymentMethod ID"
// @Success 200 {object} response.PaymentMethodResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /payment-methods/{id} [get]
func (h *PaymentMethodHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var method models.PaymentMethod
	if err := h.db.First(&method, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Способ оплаты не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить способ оплаты"},
		)
	}

	return c.JSON(response.PaymentMethodResponse{
		ID:          method.ID,
		Title:       method.Title,
		Description: method.Description,
		CreatedAt:   method.CreatedAt,
		UpdatedAt:   method.UpdatedAt,
	})
}

// CreatePaymentMethod godoc
// @Summary Создать способ оплаты
// @Description Создает новый способ оплаты
// @Tags PaymentMethod
// @Accept json
// @Produce json
// @Param data body response.PaymentMethodCreate true "Данные для создания способа оплаты"
// @Success 201 {object} response.PaymentMethodResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payment-methods [post]
func (h *PaymentMethodHandler) Create(c *fiber.Ctx) error {
	var req response.PaymentMethodCreate

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

	var existing models.PaymentMethod
	if err := h.db.Where("title = ?", req.Title).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Способ оплаты с таким названием уже существует"},
		)
	}

	method := models.PaymentMethod{
		Title:       req.Title,
		Description: req.Description,
	}

	if err := h.db.Create(&method).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать способ оплаты"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(response.PaymentMethodResponse{
		ID:          method.ID,
		Title:       method.Title,
		Description: method.Description,
		CreatedAt:   method.CreatedAt,
		UpdatedAt:   method.UpdatedAt,
	})
}

// UpdatePaymentMethod godoc
// @Summary Полностью обновить способ оплаты
// @Description Обновляет все поля способа оплаты по ID (PUT)
// @Tags PaymentMethod
// @Accept json
// @Produce json
// @Param id path int true "PaymentMethod ID"
// @Param data body response.PaymentMethodUpdate true "Обновляемые поля способа оплаты"
// @Success 200 {object} response.PaymentMethodResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payment-methods/{id} [put]
func (h *PaymentMethodHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var method models.PaymentMethod
	if err := h.db.First(&method, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Способ оплаты не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить способ оплаты"},
		)
	}

	var req response.PaymentMethodUpdate
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
		var existing models.PaymentMethod
		if err := h.db.Where("title = ? AND id != ?", req.Title, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Способ оплаты с таким названием уже существует"},
			)
		}
	}

	method.Title = req.Title
	method.Description = req.Description

	if err := h.db.Save(&method).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить способ оплаты"},
		)
	}

	return c.JSON(response.PaymentMethodResponse{
		ID:          method.ID,
		Title:       method.Title,
		Description: method.Description,
		CreatedAt:   method.CreatedAt,
		UpdatedAt:   method.UpdatedAt,
	})
}

// PatchPaymentMethod godoc
// @Summary Частично обновить способ оплаты
// @Description Обновляет указанные поля способа оплаты по ID (PATCH)
// @Tags PaymentMethod
// @Accept json
// @Produce json
// @Param id path int true "PaymentMethod ID"
// @Param data body response.PaymentMethodPatch true "Обновляемые поля способа оплаты"
// @Success 200 {object} response.PaymentMethodResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payment-methods/{id} [patch]
func (h *PaymentMethodHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var method models.PaymentMethod
	if err := h.db.First(&method, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Способ оплаты не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить способ оплаты"},
		)
	}

	var req response.PaymentMethodPatch
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
			var existing models.PaymentMethod
			if err := h.db.Where("title = ? AND id != ?", *req.Title, id).First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Способ оплаты с таким названием уже существует"},
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
			response.ErrorResponse{Error: "Не удалось обновить способ оплаты"},
		)
	}

	return c.JSON(response.PaymentMethodResponse{
		ID:          method.ID,
		Title:       method.Title,
		Description: method.Description,
		CreatedAt:   method.CreatedAt,
		UpdatedAt:   method.UpdatedAt,
	})
}

// DeletePaymentMethod godoc
// @Summary Удалить способ оплаты
// @Description Удаляет способ оплаты из БД по ID
// @Tags PaymentMethod
// @Param id path int true "PaymentMethod ID"
// @Success 204 "Способ оплаты успешно удалён"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /payment-methods/{id} [delete]
func (h *PaymentMethodHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var method models.PaymentMethod
	if err := h.db.First(&method, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Способ оплаты не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить способ оплаты"},
		)
	}

	var paymentsCount int64
	if err := h.db.Model(&models.Payment{}).Where("payment_method_id = ?", id).Count(&paymentsCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные оплаты"},
		)
	}

	if paymentsCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить способ оплаты, так как с ним связаны оплаты"},
		)
	}

	if err := h.db.Delete(&method).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить способ оплаты"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
