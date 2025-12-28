package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	catalog "app/internal/models/catalog"
	models "app/internal/models/order"
	"app/internal/response"
)

type OrderItemHandler struct {
	db *gorm.DB
}

func NewOrderItemHandler(db *gorm.DB) *OrderItemHandler {
	return &OrderItemHandler{db: db}
}

// GetAllOrderItems godoc
// @Summary Получить список товаров в заказах
// @Description Возвращает список всех товаров в заказах
// @Tags OrderItem
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param order_id query int false "Фильтр по заказу"
// @Param product_id query int false "Фильтр по товару"
// @Success 200 {object} response.OrderItemListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /order-items [get]
func (h *OrderItemHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	orderID, _ := strconv.ParseUint(c.Query("order_id"), 10, 32)
	productID, _ := strconv.ParseUint(c.Query("product_id"), 10, 32)
	offset := (page - 1) * limit

	query := h.db.Model(&models.OrderItem{}).Preload("Product")

	if orderID > 0 {
		query = query.Where("order_id = ?", orderID)
	}
	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество товаров в заказах"},
		)
	}

	var items []models.OrderItem
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить товары в заказах"},
		)
	}

	data := make([]response.OrderItemResponse, 0, len(items))
	for _, item := range items {
		totalPrice := float64(item.Quantity) * item.UnitPrice

		resp := response.OrderItemResponse{
			ID:         item.ID,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: totalPrice,
			OrderID:    item.OrderID,
			ProductID:  item.ProductID,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		}

		if item.Product != nil {
			resp.Product = &response.ProductShortResponse{
				ID:    item.Product.ID,
				Title: item.Product.Title,
				Slug:  item.Product.Slug,
				Price: item.Product.Price,
			}
		}

		data = append(data, resp)
	}

	return c.JSON(response.OrderItemListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetOrderItemByID godoc
// @Summary Получить товар в заказе по ID
// @Description Возвращает товар в заказе по ID
// @Tags OrderItem
// @Param id path int true "OrderItem ID"
// @Success 200 {object} response.OrderItemResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /order-items/{id} [get]
func (h *OrderItemHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var item models.OrderItem
	if err := h.db.Preload("Product").First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Товар в заказе не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить товар в заказе"},
		)
	}

	totalPrice := float64(item.Quantity) * item.UnitPrice

	resp := response.OrderItemResponse{
		ID:         item.ID,
		Quantity:   item.Quantity,
		UnitPrice:  item.UnitPrice,
		TotalPrice: totalPrice,
		OrderID:    item.OrderID,
		ProductID:  item.ProductID,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}

	if item.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    item.Product.ID,
			Title: item.Product.Title,
			Slug:  item.Product.Slug,
			Price: item.Product.Price,
		}
	}

	return c.JSON(resp)
}

// CreateOrderItem godoc
// @Summary Создать товар в заказе
// @Description Создает новый товар в заказе
// @Tags OrderItem
// @Accept json
// @Produce json
// @Param data body response.OrderItemCreate true "Данные для создания товара в заказе"
// @Success 201 {object} response.OrderItemResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /order-items [post]
func (h *OrderItemHandler) Create(c *fiber.Ctx) error {
	var req response.OrderItemCreate

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

	var order models.Order
	if err := h.db.First(&order, req.OrderID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Заказ не найден"},
		)
	}

	var product catalog.Product
	if err := h.db.First(&product, req.ProductID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Товар не найден"},
		)
	}

	item := models.OrderItem{
		Quantity:  req.Quantity,
		UnitPrice: req.UnitPrice,
		OrderID:   req.OrderID,
		ProductID: req.ProductID,
	}

	if err := h.db.Create(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать товар в заказе"},
		)
	}

	h.db.Preload("Product").First(&item, item.ID)

	totalPrice := float64(item.Quantity) * item.UnitPrice

	resp := response.OrderItemResponse{
		ID:         item.ID,
		Quantity:   item.Quantity,
		UnitPrice:  item.UnitPrice,
		TotalPrice: totalPrice,
		OrderID:    item.OrderID,
		ProductID:  item.ProductID,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}

	if item.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    item.Product.ID,
			Title: item.Product.Title,
			Slug:  item.Product.Slug,
			Price: item.Product.Price,
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateOrderItem godoc
// @Summary Полностью обновить товар в заказе
// @Description Обновляет все поля товара в заказе по ID (PUT)
// @Tags OrderItem
// @Accept json
// @Produce json
// @Param id path int true "OrderItem ID"
// @Param data body response.OrderItemUpdate true "Обновляемые поля товара в заказе"
// @Success 200 {object} response.OrderItemResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /order-items/{id} [put]
func (h *OrderItemHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var item models.OrderItem
	if err := h.db.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Товар в заказе не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить товар в заказе"},
		)
	}

	var req response.OrderItemUpdate
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

	var order models.Order
	if err := h.db.First(&order, req.OrderID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Заказ не найден"},
		)
	}

	var product catalog.Product
	if err := h.db.First(&product, req.ProductID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Товар не найден"},
		)
	}

	item.Quantity = req.Quantity
	item.UnitPrice = req.UnitPrice
	item.OrderID = req.OrderID
	item.ProductID = req.ProductID

	if err := h.db.Save(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить товар в заказе"},
		)
	}

	h.db.Preload("Product").First(&item, item.ID)

	totalPrice := float64(item.Quantity) * item.UnitPrice

	resp := response.OrderItemResponse{
		ID:         item.ID,
		Quantity:   item.Quantity,
		UnitPrice:  item.UnitPrice,
		TotalPrice: totalPrice,
		OrderID:    item.OrderID,
		ProductID:  item.ProductID,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}

	if item.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    item.Product.ID,
			Title: item.Product.Title,
			Slug:  item.Product.Slug,
			Price: item.Product.Price,
		}
	}

	return c.JSON(resp)
}

// PatchOrderItem godoc
// @Summary Частично обновить товар в заказе
// @Description Обновляет указанные поля товара в заказе по ID (PATCH)
// @Tags OrderItem
// @Accept json
// @Produce json
// @Param id path int true "OrderItem ID"
// @Param data body response.OrderItemPatch true "Обновляемые поля товара в заказе"
// @Success 200 {object} response.OrderItemResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /order-items/{id} [patch]
func (h *OrderItemHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var item models.OrderItem
	if err := h.db.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Товар в заказе не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить товар в заказе"},
		)
	}

	var req response.OrderItemPatch
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

	if req.Quantity != nil {
		item.Quantity = *req.Quantity
	}
	if req.UnitPrice != nil {
		item.UnitPrice = *req.UnitPrice
	}
	if req.OrderID != nil {
		var order models.Order
		if err := h.db.First(&order, *req.OrderID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Заказ не найден"},
			)
		}
		item.OrderID = *req.OrderID
	}
	if req.ProductID != nil {
		var product catalog.Product
		if err := h.db.First(&product, *req.ProductID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Товар не найден"},
			)
		}
		item.ProductID = *req.ProductID
	}

	if err := h.db.Save(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить товар в заказе"},
		)
	}

	h.db.Preload("Product").First(&item, item.ID)

	totalPrice := float64(item.Quantity) * item.UnitPrice

	resp := response.OrderItemResponse{
		ID:         item.ID,
		Quantity:   item.Quantity,
		UnitPrice:  item.UnitPrice,
		TotalPrice: totalPrice,
		OrderID:    item.OrderID,
		ProductID:  item.ProductID,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}

	if item.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    item.Product.ID,
			Title: item.Product.Title,
			Slug:  item.Product.Slug,
			Price: item.Product.Price,
		}
	}

	return c.JSON(resp)
}

// DeleteOrderItem godoc
// @Summary Удалить товар в заказе
// @Description Удаляет товар в заказе из БД по ID
// @Tags OrderItem
// @Param id path int true "OrderItem ID"
// @Success 204 "Товар в заказе успешно удален"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /order-items/{id} [delete]
func (h *OrderItemHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var item models.OrderItem
	if err := h.db.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Товар в заказе не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить товар в заказе"},
		)
	}

	if err := h.db.Delete(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить товар в заказе"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
