// internal/handlers/request.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/order"
	user "app/internal/models/user"
	"app/internal/response"
)

type RequestHandler struct {
	db *gorm.DB
}

func NewRequestHandler(db *gorm.DB) *RequestHandler {
	return &RequestHandler{db: db}
}

// GetAllRequests godoc
// @Summary Получить список заявок
// @Description Возвращает список всех заявок
// @Tags Request
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param number query string false "Поиск по номеру"
// @Param order_id query int false "Фильтр по заказу"
// @Param user_id query int false "Фильтр по пользователю"
// @Param request_status_id query int false "Фильтр по статусу заявки"
// @Success 200 {object} response.RequestListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /requests [get]
func (h *RequestHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	number := c.Query("number")
	orderID, _ := strconv.ParseUint(c.Query("order_id"), 10, 32)
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	requestStatusID, _ := strconv.ParseUint(c.Query("request_status_id"), 10, 32)
	offset := (page - 1) * limit

	query := h.db.Model(&models.Request{}).
		Preload("Order").
		Preload("User").
		Preload("RequestStatus")

	if number != "" {
		query = query.Where("number ILIKE ?", "%"+number+"%")
	}
	if orderID > 0 {
		query = query.Where("order_id = ?", orderID)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if requestStatusID > 0 {
		query = query.Where("request_status_id = ?", requestStatusID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество заявок"},
		)
	}

	var requests []models.Request
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&requests).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить заявки"},
		)
	}

	data := make([]response.RequestResponse, 0, len(requests))
	for _, req := range requests {
		resp := response.RequestResponse{
			ID:              req.ID,
			Number:          req.Number,
			UserComment:     req.UserComment,
			ManagerComment:  req.ManagerComment,
			OrderID:         req.OrderID,
			UserID:          req.UserID,
			RequestStatusID: req.RequestStatusID,
			CreatedAt:       req.CreatedAt,
			UpdatedAt:       req.UpdatedAt,
		}

		if req.Order != nil {
			resp.Order = &response.OrderShortResponse{
				ID:     req.Order.ID,
				Number: req.Order.Number,
				Price:  req.Order.Price,
			}
		}

		if req.User != nil {
			resp.User = &response.UserShortResponse{
				ID:       req.User.ID,
				Username: req.User.Username,
				Email:    req.User.Email,
			}
		}

		if req.RequestStatus != nil {
			resp.RequestStatus = &response.RequestStatusResponse{
				ID:              req.RequestStatus.ID,
				Title:           req.RequestStatus.Title,
				BackgroundColor: req.RequestStatus.BackgroundColor,
			}
		}

		data = append(data, resp)
	}

	return c.JSON(response.RequestListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetRequestByID godoc
// @Summary Получить заявку по ID
// @Description Возвращает заявку по ID
// @Tags Request
// @Param id path int true "Request ID"
// @Success 200 {object} response.RequestResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /requests/{id} [get]
func (h *RequestHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var req models.Request
	if err := h.db.Preload("Order").
		Preload("User").
		Preload("RequestStatus").
		First(&req, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Заявка не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить заявку"},
		)
	}

	resp := response.RequestResponse{
		ID:              req.ID,
		Number:          req.Number,
		UserComment:     req.UserComment,
		ManagerComment:  req.ManagerComment,
		OrderID:         req.OrderID,
		UserID:          req.UserID,
		RequestStatusID: req.RequestStatusID,
		CreatedAt:       req.CreatedAt,
		UpdatedAt:       req.UpdatedAt,
	}

	if req.Order != nil {
		resp.Order = &response.OrderShortResponse{
			ID:     req.Order.ID,
			Number: req.Order.Number,
			Price:  req.Order.Price,
		}
	}

	if req.User != nil {
		resp.User = &response.UserShortResponse{
			ID:       req.User.ID,
			Username: req.User.Username,
			Email:    req.User.Email,
		}
	}

	if req.RequestStatus != nil {
		resp.RequestStatus = &response.RequestStatusResponse{
			ID:              req.RequestStatus.ID,
			Title:           req.RequestStatus.Title,
			BackgroundColor: req.RequestStatus.BackgroundColor,
		}
	}

	return c.JSON(resp)
}

// CreateRequest godoc
// @Summary Создать заявку
// @Description Создает новую заявку
// @Tags Request
// @Accept json
// @Produce json
// @Param data body response.RequestCreate true "Данные для создания заявки"
// @Success 201 {object} response.RequestResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /requests [post]
func (h *RequestHandler) Create(c *fiber.Ctx) error {
	var createReq response.RequestCreate

	if err := c.BodyParser(&createReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат тела запроса"},
		)
	}

	if err := validate.Struct(createReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: err.Error()},
		)
	}

	var order models.Order
	if err := h.db.First(&order, createReq.OrderID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Заказ не найден"},
		)
	}

	var user user.User
	if err := h.db.First(&user, createReq.UserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Пользователь не найден"},
		)
	}

	var requestStatus models.RequestStatus
	if err := h.db.First(&requestStatus, createReq.RequestStatusID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Статус заявки не найден"},
		)
	}

	var existingRequest models.Request
	if err := h.db.Where("order_id = ?", createReq.OrderID).First(&existingRequest).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Заявка для этого заказа уже существует"},
		)
	}

	req := models.Request{
		UserComment:     createReq.UserComment,
		ManagerComment:  createReq.ManagerComment,
		OrderID:         createReq.OrderID,
		UserID:          createReq.UserID,
		RequestStatusID: createReq.RequestStatusID,
	}

	if err := h.db.Create(&req).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать заявку"},
		)
	}

	h.db.Preload("Order").
		Preload("User").
		Preload("RequestStatus").
		First(&req, req.ID)

	resp := response.RequestResponse{
		ID:              req.ID,
		Number:          req.Number,
		UserComment:     req.UserComment,
		ManagerComment:  req.ManagerComment,
		OrderID:         req.OrderID,
		UserID:          req.UserID,
		RequestStatusID: req.RequestStatusID,
		CreatedAt:       req.CreatedAt,
		UpdatedAt:       req.UpdatedAt,
	}

	if req.Order != nil {
		resp.Order = &response.OrderShortResponse{
			ID:     req.Order.ID,
			Number: req.Order.Number,
			Price:  req.Order.Price,
		}
	}

	if req.User != nil {
		resp.User = &response.UserShortResponse{
			ID:       req.User.ID,
			Username: req.User.Username,
			Email:    req.User.Email,
		}
	}

	if req.RequestStatus != nil {
		resp.RequestStatus = &response.RequestStatusResponse{
			ID:              req.RequestStatus.ID,
			Title:           req.RequestStatus.Title,
			BackgroundColor: req.RequestStatus.BackgroundColor,
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateRequest godoc
// @Summary Полностью обновить заявку
// @Description Обновляет все поля заявки по ID (PUT)
// @Tags Request
// @Accept json
// @Produce json
// @Param id path int true "Request ID"
// @Param data body response.RequestUpdate true "Обновляемые поля заявки"
// @Success 200 {object} response.RequestResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /requests/{id} [put]
func (h *RequestHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var req models.Request
	if err := h.db.First(&req, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Заявка не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить заявку"},
		)
	}

	var updateReq response.RequestUpdate
	if err := c.BodyParser(&updateReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат тела запроса"},
		)
	}

	if err := validate.Struct(updateReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: err.Error()},
		)
	}

	var order models.Order
	if err := h.db.First(&order, updateReq.OrderID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Заказ не найден"},
		)
	}

	var user user.User
	if err := h.db.First(&user, updateReq.UserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Пользователь не найден"},
		)
	}

	var requestStatus models.RequestStatus
	if err := h.db.First(&requestStatus, updateReq.RequestStatusID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Статус заявки не найден"},
		)
	}

	if req.OrderID != updateReq.OrderID {
		var existingRequest models.Request
		if err := h.db.Where("order_id = ? AND id != ?", updateReq.OrderID, id).First(&existingRequest).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Заявка для этого заказа уже существует"},
			)
		}
	}

	req.UserComment = updateReq.UserComment
	req.ManagerComment = updateReq.ManagerComment
	req.OrderID = updateReq.OrderID
	req.UserID = updateReq.UserID
	req.RequestStatusID = updateReq.RequestStatusID

	if err := h.db.Save(&req).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить заявку"},
		)
	}

	h.db.Preload("Order").
		Preload("User").
		Preload("RequestStatus").
		First(&req, req.ID)

	resp := response.RequestResponse{
		ID:              req.ID,
		Number:          req.Number,
		UserComment:     req.UserComment,
		ManagerComment:  req.ManagerComment,
		OrderID:         req.OrderID,
		UserID:          req.UserID,
		RequestStatusID: req.RequestStatusID,
		CreatedAt:       req.CreatedAt,
		UpdatedAt:       req.UpdatedAt,
	}

	if req.Order != nil {
		resp.Order = &response.OrderShortResponse{
			ID:     req.Order.ID,
			Number: req.Order.Number,
			Price:  req.Order.Price,
		}
	}

	if req.User != nil {
		resp.User = &response.UserShortResponse{
			ID:       req.User.ID,
			Username: req.User.Username,
			Email:    req.User.Email,
		}
	}

	if req.RequestStatus != nil {
		resp.RequestStatus = &response.RequestStatusResponse{
			ID:              req.RequestStatus.ID,
			Title:           req.RequestStatus.Title,
			BackgroundColor: req.RequestStatus.BackgroundColor,
		}
	}

	return c.JSON(resp)
}

// PatchRequest godoc
// @Summary Частично обновить заявку
// @Description Обновляет указанные поля заявки по ID (PATCH)
// @Tags Request
// @Accept json
// @Produce json
// @Param id path int true "Request ID"
// @Param data body response.RequestPatch true "Обновляемые поля заявки"
// @Success 200 {object} response.RequestResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /requests/{id} [patch]
func (h *RequestHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var req models.Request
	if err := h.db.First(&req, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Заявка не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить заявку"},
		)
	}

	var patchReq response.RequestPatch
	if err := c.BodyParser(&patchReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат тела запроса"},
		)
	}

	if err := validate.Struct(patchReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: err.Error()},
		)
	}

	if patchReq.UserComment != nil {
		req.UserComment = *patchReq.UserComment
	}
	if patchReq.ManagerComment != nil {
		req.ManagerComment = *patchReq.ManagerComment
	}
	if patchReq.OrderID != nil {
		var order models.Order
		if err := h.db.First(&order, *patchReq.OrderID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Заказ не найден"},
			)
		}
		if req.OrderID != *patchReq.OrderID {
			var existingRequest models.Request
			if err := h.db.Where("order_id = ? AND id != ?", *patchReq.OrderID, id).First(&existingRequest).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Заявка для этого заказа уже существует"},
				)
			}
		}
		req.OrderID = *patchReq.OrderID
	}
	if patchReq.UserID != nil {
		var user user.User
		if err := h.db.First(&user, *patchReq.UserID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Пользователь не найден"},
			)
		}
		req.UserID = *patchReq.UserID
	}
	if patchReq.RequestStatusID != nil {
		var requestStatus models.RequestStatus
		if err := h.db.First(&requestStatus, *patchReq.RequestStatusID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус заявки не найден"},
			)
		}
		req.RequestStatusID = *patchReq.RequestStatusID
	}

	if err := h.db.Save(&req).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить заявку"},
		)
	}

	h.db.Preload("Order").
		Preload("User").
		Preload("RequestStatus").
		First(&req, req.ID)

	resp := response.RequestResponse{
		ID:              req.ID,
		Number:          req.Number,
		UserComment:     req.UserComment,
		ManagerComment:  req.ManagerComment,
		OrderID:         req.OrderID,
		UserID:          req.UserID,
		RequestStatusID: req.RequestStatusID,
		CreatedAt:       req.CreatedAt,
		UpdatedAt:       req.UpdatedAt,
	}

	if req.Order != nil {
		resp.Order = &response.OrderShortResponse{
			ID:     req.Order.ID,
			Number: req.Order.Number,
			Price:  req.Order.Price,
		}
	}

	if req.User != nil {
		resp.User = &response.UserShortResponse{
			ID:       req.User.ID,
			Username: req.User.Username,
			Email:    req.User.Email,
		}
	}

	if req.RequestStatus != nil {
		resp.RequestStatus = &response.RequestStatusResponse{
			ID:              req.RequestStatus.ID,
			Title:           req.RequestStatus.Title,
			BackgroundColor: req.RequestStatus.BackgroundColor,
		}
	}

	return c.JSON(resp)
}

// DeleteRequest godoc
// @Summary Удалить заявку
// @Description Удаляет заявку из БД по ID
// @Tags Request
// @Param id path int true "Request ID"
// @Success 204 "Заявка успешно удалена"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /requests/{id} [delete]
func (h *RequestHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var req models.Request
	if err := h.db.First(&req, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Заявка не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить заявку"},
		)
	}

	if err := h.db.Delete(&req).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить заявку"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
