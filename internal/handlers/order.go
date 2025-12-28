// internal/handlers/order.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/order"
	user "app/internal/models/user"
	"app/internal/response"
)

type OrderHandler struct {
	db *gorm.DB
}

func NewOrderHandler(db *gorm.DB) *OrderHandler {
	return &OrderHandler{db: db}
}

// GetAllOrders godoc
// @Summary Получить список заказов
// @Description Возвращает список всех заказов
// @Tags Order
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param number query string false "Поиск по номеру"
// @Param user_id query int false "Фильтр по пользователю"
// @Param order_status_id query int false "Фильтр по статусу заказа"
// @Param is_urgent query bool false "Фильтр по срочности"
// @Param start_date query string false "Дата начала (YYYY-MM-DD)"
// @Param end_date query string false "Дата окончания (YYYY-MM-DD)"
// @Success 200 {object} response.OrderListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /orders [get]
func (h *OrderHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	number := c.Query("number")
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	orderStatusID, _ := strconv.ParseUint(c.Query("order_status_id"), 10, 32)
	isUrgent, _ := strconv.ParseBool(c.Query("is_urgent"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	offset := (page - 1) * limit

	query := h.db.Model(&models.Order{}).
		Preload("User").
		Preload("Shipment").
		Preload("Payment").
		Preload("OrderStatus").
		Preload("OrderItems")

	if number != "" {
		query = query.Where("number ILIKE ?", "%"+number+"%")
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if orderStatusID > 0 {
		query = query.Where("order_status_id = ?", orderStatusID)
	}
	if c.Query("is_urgent") != "" {
		query = query.Where("is_urgent = ?", isUrgent)
	}
	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество заказов"},
		)
	}

	var orders []models.Order
	if err := query.Order("date DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить заказы"},
		)
	}

	data := make([]response.OrderResponse, 0, len(orders))
	for _, order := range orders {
		totalAmount := order.Price - order.Discount
		if order.Shipment != nil {
			totalAmount += order.Shipment.Price
		}

		resp := response.OrderResponse{
			ID:            order.ID,
			Number:        order.Number,
			Date:          order.Date,
			Price:         order.Price,
			Discount:      order.Discount,
			IsUrgent:      order.IsUrgent,
			TotalAmount:   totalAmount,
			UserID:        order.UserID,
			ShipmentID:    order.ShipmentID,
			PaymentID:     order.PaymentID,
			OrderStatusID: order.OrderStatusID,
			CreatedAt:     order.CreatedAt,
			UpdatedAt:     order.UpdatedAt,
		}

		if order.User != nil {
			resp.User = &response.UserResponse{
				ID:        order.User.ID,
				Username:  order.User.Username,
				Email:     order.User.Email,
				FirstName: order.User.FirstName,
				LastName:  order.User.LastName,
			}
		}

		if order.Shipment != nil {
			resp.Shipment = &response.ShipmentResponse{
				ID:     order.Shipment.ID,
				Number: order.Shipment.Number,
				Price:  order.Shipment.Price,
			}
		}

		if order.Payment != nil {
			resp.Payment = &response.PaymentResponse{
				ID:          order.Payment.ID,
				PaymentDate: order.Payment.PaymentDate,
			}
		}

		if order.OrderStatus != nil {
			resp.OrderStatus = &response.OrderStatusResponse{
				ID:              order.OrderStatus.ID,
				Title:           order.OrderStatus.Title,
				BackgroundColor: order.OrderStatus.BackgroundColor,
			}
		}

		if len(order.OrderItems) > 0 {
			resp.OrderItems = make([]response.OrderItemResponse, 0, len(order.OrderItems))
			for _, item := range order.OrderItems {
				resp.OrderItems = append(resp.OrderItems, response.OrderItemResponse{
					ID:        item.ID,
					Quantity:  item.Quantity,
					UnitPrice: item.UnitPrice,
					ProductID: item.ProductID,
				})
			}
		}

		data = append(data, resp)
	}

	return c.JSON(response.OrderListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetOrderByID godoc
// @Summary Получить заказ по ID
// @Description Возвращает заказ по ID
// @Tags Order
// @Param id path int true "Order ID"
// @Success 200 {object} response.OrderResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /orders/{id} [get]
func (h *OrderHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var order models.Order
	if err := h.db.Preload("User").
		Preload("Shipment").
		Preload("Payment").
		Preload("OrderStatus").
		Preload("OrderItems").
		First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Заказ не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить заказ"},
		)
	}

	totalAmount := order.Price - order.Discount
	if order.Shipment != nil {
		totalAmount += order.Shipment.Price
	}

	resp := response.OrderResponse{
		ID:            order.ID,
		Number:        order.Number,
		Date:          order.Date,
		Price:         order.Price,
		Discount:      order.Discount,
		IsUrgent:      order.IsUrgent,
		TotalAmount:   totalAmount,
		UserID:        order.UserID,
		ShipmentID:    order.ShipmentID,
		PaymentID:     order.PaymentID,
		OrderStatusID: order.OrderStatusID,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}

	if order.User != nil {
		resp.User = &response.UserResponse{
			ID:        order.User.ID,
			Username:  order.User.Username,
			Email:     order.User.Email,
			FirstName: order.User.FirstName,
			LastName:  order.User.LastName,
		}
	}

	if order.Shipment != nil {
		resp.Shipment = &response.ShipmentResponse{
			ID:     order.Shipment.ID,
			Number: order.Shipment.Number,
			Price:  order.Shipment.Price,
		}
	}

	if order.Payment != nil {
		resp.Payment = &response.PaymentResponse{
			ID:          order.Payment.ID,
			PaymentDate: order.Payment.PaymentDate,
		}
	}

	if order.OrderStatus != nil {
		resp.OrderStatus = &response.OrderStatusResponse{
			ID:              order.OrderStatus.ID,
			Title:           order.OrderStatus.Title,
			BackgroundColor: order.OrderStatus.BackgroundColor,
		}
	}

	if len(order.OrderItems) > 0 {
		resp.OrderItems = make([]response.OrderItemResponse, 0, len(order.OrderItems))
		for _, item := range order.OrderItems {
			resp.OrderItems = append(resp.OrderItems, response.OrderItemResponse{
				ID:        item.ID,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
				ProductID: item.ProductID,
			})
		}
	}

	return c.JSON(resp)
}

// CreateOrder godoc
// @Summary Создать заказ
// @Description Создает новый заказ
// @Tags Order
// @Accept json
// @Produce json
// @Param data body response.OrderCreate true "Данные для создания заказа"
// @Success 201 {object} response.OrderResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /orders [post]
func (h *OrderHandler) Create(c *fiber.Ctx) error {
	var req response.OrderCreate

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

	var user user.User
	if err := h.db.First(&user, req.UserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Пользователь не найден"},
		)
	}

	var shipment models.Shipment
	if err := h.db.First(&shipment, req.ShipmentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Доставка не найдена"},
		)
	}

	var payment models.Payment
	if err := h.db.First(&payment, req.PaymentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Оплата не найдена"},
		)
	}

	var orderStatus models.OrderStatus
	if err := h.db.First(&orderStatus, req.OrderStatusID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Статус заказа не найден"},
		)
	}

	order := models.Order{
		Price:         req.Price,
		Discount:      req.Discount,
		IsUrgent:      req.IsUrgent,
		UserID:        req.UserID,
		ShipmentID:    req.ShipmentID,
		PaymentID:     req.PaymentID,
		OrderStatusID: req.OrderStatusID,
	}

	if err := h.db.Create(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать заказ"},
		)
	}

	h.db.Preload("User").
		Preload("Shipment").
		Preload("Payment").
		Preload("OrderStatus").
		First(&order, order.ID)

	totalAmount := order.Price - order.Discount
	if order.Shipment != nil {
		totalAmount += order.Shipment.Price
	}

	resp := response.OrderResponse{
		ID:            order.ID,
		Number:        order.Number,
		Date:          order.Date,
		Price:         order.Price,
		Discount:      order.Discount,
		IsUrgent:      order.IsUrgent,
		TotalAmount:   totalAmount,
		UserID:        order.UserID,
		ShipmentID:    order.ShipmentID,
		PaymentID:     order.PaymentID,
		OrderStatusID: order.OrderStatusID,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}

	if order.User != nil {
		resp.User = &response.UserResponse{
			ID:       order.User.ID,
			Username: order.User.Username,
			Email:    order.User.Email,
		}
	}

	if order.Shipment != nil {
		resp.Shipment = &response.ShipmentResponse{
			ID:     order.Shipment.ID,
			Number: order.Shipment.Number,
			Price:  order.Shipment.Price,
		}
	}

	if order.Payment != nil {
		resp.Payment = &response.PaymentResponse{
			ID:          order.Payment.ID,
			PaymentDate: order.Payment.PaymentDate,
		}
	}

	if order.OrderStatus != nil {
		resp.OrderStatus = &response.OrderStatusResponse{
			ID:              order.OrderStatus.ID,
			Title:           order.OrderStatus.Title,
			BackgroundColor: order.OrderStatus.BackgroundColor,
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateOrder godoc
// @Summary Полностью обновить заказ
// @Description Обновляет все поля заказа по ID (PUT)
// @Tags Order
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param data body response.OrderUpdate true "Обновляемые поля заказа"
// @Success 200 {object} response.OrderResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /orders/{id} [put]
func (h *OrderHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var order models.Order
	if err := h.db.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Заказ не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить заказ"},
		)
	}

	var req response.OrderUpdate
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

	var user user.User
	if err := h.db.First(&user, req.UserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Пользователь не найден"},
		)
	}

	var shipment models.Shipment
	if err := h.db.First(&shipment, req.ShipmentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Доставка не найдена"},
		)
	}

	var payment models.Payment
	if err := h.db.First(&payment, req.PaymentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Оплата не найдена"},
		)
	}

	var orderStatus models.OrderStatus
	if err := h.db.First(&orderStatus, req.OrderStatusID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Статус заказа не найден"},
		)
	}

	order.Price = req.Price
	order.Discount = req.Discount
	order.IsUrgent = req.IsUrgent
	order.UserID = req.UserID
	order.ShipmentID = req.ShipmentID
	order.PaymentID = req.PaymentID
	order.OrderStatusID = req.OrderStatusID

	if err := h.db.Save(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить заказ"},
		)
	}

	h.db.Preload("User").
		Preload("Shipment").
		Preload("Payment").
		Preload("OrderStatus").
		First(&order, order.ID)

	totalAmount := order.Price - order.Discount
	if order.Shipment != nil {
		totalAmount += order.Shipment.Price
	}

	resp := response.OrderResponse{
		ID:            order.ID,
		Number:        order.Number,
		Date:          order.Date,
		Price:         order.Price,
		Discount:      order.Discount,
		IsUrgent:      order.IsUrgent,
		TotalAmount:   totalAmount,
		UserID:        order.UserID,
		ShipmentID:    order.ShipmentID,
		PaymentID:     order.PaymentID,
		OrderStatusID: order.OrderStatusID,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}

	if order.User != nil {
		resp.User = &response.UserResponse{
			ID:       order.User.ID,
			Username: order.User.Username,
			Email:    order.User.Email,
		}
	}

	if order.Shipment != nil {
		resp.Shipment = &response.ShipmentResponse{
			ID:     order.Shipment.ID,
			Number: order.Shipment.Number,
			Price:  order.Shipment.Price,
		}
	}

	if order.Payment != nil {
		resp.Payment = &response.PaymentResponse{
			ID:          order.Payment.ID,
			PaymentDate: order.Payment.PaymentDate,
		}
	}

	if order.OrderStatus != nil {
		resp.OrderStatus = &response.OrderStatusResponse{
			ID:              order.OrderStatus.ID,
			Title:           order.OrderStatus.Title,
			BackgroundColor: order.OrderStatus.BackgroundColor,
		}
	}

	return c.JSON(resp)
}

// PatchOrder godoc
// @Summary Частично обновить заказ
// @Description Обновляет указанные поля заказа по ID (PATCH)
// @Tags Order
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param data body response.OrderPatch true "Обновляемые поля заказа"
// @Success 200 {object} response.OrderResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /orders/{id} [patch]
func (h *OrderHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var order models.Order
	if err := h.db.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Заказ не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить заказ"},
		)
	}

	var req response.OrderPatch
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
		order.Price = *req.Price
	}
	if req.Discount != nil {
		order.Discount = *req.Discount
	}
	if req.IsUrgent != nil {
		order.IsUrgent = *req.IsUrgent
	}
	if req.UserID != nil {
		var user user.User
		if err := h.db.First(&user, *req.UserID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Пользователь не найден"},
			)
		}
		order.UserID = *req.UserID
	}
	if req.ShipmentID != nil {
		var shipment models.Shipment
		if err := h.db.First(&shipment, *req.ShipmentID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Доставка не найдена"},
			)
		}
		order.ShipmentID = *req.ShipmentID
	}
	if req.PaymentID != nil {
		var payment models.Payment
		if err := h.db.First(&payment, *req.PaymentID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Оплата не найдена"},
			)
		}
		order.PaymentID = *req.PaymentID
	}
	if req.OrderStatusID != nil {
		var orderStatus models.OrderStatus
		if err := h.db.First(&orderStatus, *req.OrderStatusID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус заказа не найден"},
			)
		}
		order.OrderStatusID = *req.OrderStatusID
	}

	if err := h.db.Save(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить заказ"},
		)
	}

	h.db.Preload("User").
		Preload("Shipment").
		Preload("Payment").
		Preload("OrderStatus").
		First(&order, order.ID)

	totalAmount := order.Price - order.Discount
	if order.Shipment != nil {
		totalAmount += order.Shipment.Price
	}

	resp := response.OrderResponse{
		ID:            order.ID,
		Number:        order.Number,
		Date:          order.Date,
		Price:         order.Price,
		Discount:      order.Discount,
		IsUrgent:      order.IsUrgent,
		TotalAmount:   totalAmount,
		UserID:        order.UserID,
		ShipmentID:    order.ShipmentID,
		PaymentID:     order.PaymentID,
		OrderStatusID: order.OrderStatusID,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}

	if order.User != nil {
		resp.User = &response.UserResponse{
			ID:       order.User.ID,
			Username: order.User.Username,
			Email:    order.User.Email,
		}
	}

	if order.Shipment != nil {
		resp.Shipment = &response.ShipmentResponse{
			ID:     order.Shipment.ID,
			Number: order.Shipment.Number,
			Price:  order.Shipment.Price,
		}
	}

	if order.Payment != nil {
		resp.Payment = &response.PaymentResponse{
			ID:          order.Payment.ID,
			PaymentDate: order.Payment.PaymentDate,
		}
	}

	if order.OrderStatus != nil {
		resp.OrderStatus = &response.OrderStatusResponse{
			ID:              order.OrderStatus.ID,
			Title:           order.OrderStatus.Title,
			BackgroundColor: order.OrderStatus.BackgroundColor,
		}
	}

	return c.JSON(resp)
}

// DeleteOrder godoc
// @Summary Удалить заказ
// @Description Удаляет заказ из БД по ID
// @Tags Order
// @Param id path int true "Order ID"
// @Success 204 "Заказ успешно удален"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /orders/{id} [delete]
func (h *OrderHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var order models.Order
	if err := h.db.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Заказ не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить заказ"},
		)
	}

	var orderItemsCount int64
	if err := h.db.Model(&models.OrderItem{}).Where("order_id = ?", id).Count(&orderItemsCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные товары"},
		)
	}

	if orderItemsCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить заказ, так как в нем есть товары"},
		)
	}

	var requestCount int64
	if err := h.db.Model(&models.Request{}).Where("order_id = ?", id).Count(&requestCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные заявки"},
		)
	}

	if requestCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить заказ, так как с ним связаны заявки"},
		)
	}

	if err := h.db.Delete(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить заказ"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
