// internal/handlers/product_warehouse.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/catalog"
	"app/internal/response"
)

type ProductWarehouseHandler struct {
	db *gorm.DB
}

func NewProductWarehouseHandler(db *gorm.DB) *ProductWarehouseHandler {
	return &ProductWarehouseHandler{db: db}
}

// GetAllProductWarehouses godoc
// @Summary Получить список записей склада
// @Description Возвращает список всех записей склада с пагинацией
// @Tags ProductWarehouse
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param product_id query int false "Фильтр по товару"
// @Param min_count query int false "Минимальное количество на складе"
// @Param max_count query int false "Максимальное количество на складе"
// @Param low_stock query bool false "Только товары с низким запасом (< 10)" default(false)
// @Success 200 {object} response.ProductWarehouseListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /product-warehouses [get]
func (h *ProductWarehouseHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	productID, _ := strconv.ParseUint(c.Query("product_id"), 10, 32)
	minCount, _ := strconv.ParseUint(c.Query("min_count"), 10, 32)
	maxCount, _ := strconv.ParseUint(c.Query("max_count"), 10, 32)
	lowStock, _ := strconv.ParseBool(c.Query("low_stock", "false"))
	offset := (page - 1) * limit

	query := h.db.Model(&models.ProductWarehouse{}).Preload("Product")

	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}
	if minCount > 0 {
		query = query.Where("count >= ?", minCount)
	}
	if maxCount > 0 {
		query = query.Where("count <= ?", maxCount)
	}
	if lowStock {
		query = query.Where("count < 10")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество записей склада"},
		)
	}

	var warehouses []models.ProductWarehouse
	if err := query.Order("count ASC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&warehouses).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить записи склада"},
		)
	}

	data := make([]response.ProductWarehouseResponse, 0, len(warehouses))
	for _, warehouse := range warehouses {
		resp := response.ProductWarehouseResponse{
			ID:        warehouse.ID,
			Count:     warehouse.Count,
			ProductID: warehouse.ProductID,
			CreatedAt: warehouse.CreatedAt,
			UpdatedAt: warehouse.UpdatedAt,
		}

		if warehouse.Product != nil {
			resp.Product = &response.ProductShortResponse{
				ID:    warehouse.Product.ID,
				Slug:  warehouse.Product.Slug,
				Title: warehouse.Product.Title,
				Price: warehouse.Product.Price,
			}
		}

		data = append(data, resp)
	}

	return c.JSON(response.ProductWarehouseListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetProductWarehouseByID godoc
// @Summary Получить запись склада по ID
// @Description Возвращает запись склада по ID
// @Tags ProductWarehouse
// @Param id path int true "ProductWarehouse ID"
// @Success 200 {object} response.ProductWarehouseResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /product-warehouses/{id} [get]
func (h *ProductWarehouseHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var warehouse models.ProductWarehouse
	if err := h.db.Preload("Product").
		First(&warehouse, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Запись склада не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить запись склада"},
		)
	}

	resp := response.ProductWarehouseResponse{
		ID:        warehouse.ID,
		Count:     warehouse.Count,
		ProductID: warehouse.ProductID,
		CreatedAt: warehouse.CreatedAt,
		UpdatedAt: warehouse.UpdatedAt,
	}

	if warehouse.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    warehouse.Product.ID,
			Slug:  warehouse.Product.Slug,
			Title: warehouse.Product.Title,
			Price: warehouse.Product.Price,
		}
	}

	return c.JSON(resp)
}

// GetProductWarehouseByProductID godoc
// @Summary Получить запись склада по ID товара
// @Description Возвращает запись склада по ID товара
// @Tags ProductWarehouse
// @Param product_id path int true "Product ID"
// @Success 200 {object} response.ProductWarehouseResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /product-warehouses/product/{product_id} [get]
func (h *ProductWarehouseHandler) GetByProductID(c *fiber.Ctx) error {
	productID, err := strconv.ParseUint(c.Params("product_id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID товара"},
		)
	}

	var warehouse models.ProductWarehouse
	if err := h.db.Preload("Product").
		Where("product_id = ?", productID).
		First(&warehouse).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Запись склада для этого товара не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить запись склада"},
		)
	}

	resp := response.ProductWarehouseResponse{
		ID:        warehouse.ID,
		Count:     warehouse.Count,
		ProductID: warehouse.ProductID,
		CreatedAt: warehouse.CreatedAt,
		UpdatedAt: warehouse.UpdatedAt,
	}

	if warehouse.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    warehouse.Product.ID,
			Slug:  warehouse.Product.Slug,
			Title: warehouse.Product.Title,
			Price: warehouse.Product.Price,
		}
	}

	return c.JSON(resp)
}

// CreateProductWarehouse godoc
// @Summary Создать запись склада
// @Description Создает новую запись склада для товара
// @Tags ProductWarehouse
// @Accept json
// @Produce json
// @Param data body response.ProductWarehouseCreate true "Данные для создания записи склада"
// @Success 201 {object} response.ProductWarehouseResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /product-warehouses [post]
func (h *ProductWarehouseHandler) Create(c *fiber.Ctx) error {
	var req response.ProductWarehouseCreate

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

	var product models.Product
	if err := h.db.First(&product, req.ProductID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Товар не найден"},
		)
	}

	var existing models.ProductWarehouse
	if err := h.db.Where("product_id = ?", req.ProductID).
		First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Запись склада для этого товара уже существует"},
		)
	}

	warehouse := models.ProductWarehouse{
		Count:     req.Count,
		ProductID: req.ProductID,
	}

	if err := h.db.Create(&warehouse).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать запись склада"},
		)
	}

	h.db.Preload("Product").First(&warehouse, warehouse.ID)

	resp := response.ProductWarehouseResponse{
		ID:        warehouse.ID,
		Count:     warehouse.Count,
		ProductID: warehouse.ProductID,
		CreatedAt: warehouse.CreatedAt,
		UpdatedAt: warehouse.UpdatedAt,
	}

	if warehouse.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    warehouse.Product.ID,
			Slug:  warehouse.Product.Slug,
			Title: warehouse.Product.Title,
			Price: warehouse.Product.Price,
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateProductWarehouse godoc
// @Summary Полностью обновить запись склада
// @Description Обновляет все поля записи склада по ID (PUT)
// @Tags ProductWarehouse
// @Accept json
// @Produce json
// @Param id path int true "ProductWarehouse ID"
// @Param data body response.ProductWarehouseUpdate true "Обновляемые поля записи склада"
// @Success 200 {object} response.ProductWarehouseResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /product-warehouses/{id} [put]
func (h *ProductWarehouseHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var warehouse models.ProductWarehouse
	if err := h.db.First(&warehouse, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Запись склада не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить запись склада"},
		)
	}

	var req response.ProductWarehouseUpdate
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

	var product models.Product
	if err := h.db.First(&product, req.ProductID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Товар не найден"},
		)
	}

	if warehouse.ProductID != req.ProductID {
		var existing models.ProductWarehouse
		if err := h.db.Where("product_id = ? AND id != ?", req.ProductID, id).
			First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Запись склада для этого товара уже существует"},
			)
		}
	}

	warehouse.Count = req.Count
	warehouse.ProductID = req.ProductID

	if err := h.db.Save(&warehouse).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить запись склада"},
		)
	}

	h.db.Preload("Product").First(&warehouse, warehouse.ID)

	resp := response.ProductWarehouseResponse{
		ID:        warehouse.ID,
		Count:     warehouse.Count,
		ProductID: warehouse.ProductID,
		CreatedAt: warehouse.CreatedAt,
		UpdatedAt: warehouse.UpdatedAt,
	}

	if warehouse.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    warehouse.Product.ID,
			Slug:  warehouse.Product.Slug,
			Title: warehouse.Product.Title,
			Price: warehouse.Product.Price,
		}
	}

	return c.JSON(resp)
}

// PatchProductWarehouse godoc
// @Summary Частично обновить запись склада
// @Description Обновляет указанные поля записи склада по ID (PATCH)
// @Tags ProductWarehouse
// @Accept json
// @Produce json
// @Param id path int true "ProductWarehouse ID"
// @Param data body response.ProductWarehousePatch true "Обновляемые поля записи склада"
// @Success 200 {object} response.ProductWarehouseResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /product-warehouses/{id} [patch]
func (h *ProductWarehouseHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var warehouse models.ProductWarehouse
	if err := h.db.First(&warehouse, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Запись склада не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить запись склада"},
		)
	}

	var req response.ProductWarehousePatch
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

	if req.Count != nil {
		warehouse.Count = *req.Count
	}
	if req.ProductID != nil {
		var product models.Product
		if err := h.db.First(&product, *req.ProductID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Товар не найден"},
			)
		}

		if warehouse.ProductID != *req.ProductID {
			var existing models.ProductWarehouse
			if err := h.db.Where("product_id = ? AND id != ?", *req.ProductID, id).
				First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Запись склада для этого товара уже существует"},
				)
			}
		}
		warehouse.ProductID = *req.ProductID
	}

	if err := h.db.Save(&warehouse).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить запись склада"},
		)
	}

	h.db.Preload("Product").First(&warehouse, warehouse.ID)

	resp := response.ProductWarehouseResponse{
		ID:        warehouse.ID,
		Count:     warehouse.Count,
		ProductID: warehouse.ProductID,
		CreatedAt: warehouse.CreatedAt,
		UpdatedAt: warehouse.UpdatedAt,
	}

	if warehouse.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    warehouse.Product.ID,
			Slug:  warehouse.Product.Slug,
			Title: warehouse.Product.Title,
			Price: warehouse.Product.Price,
		}
	}

	return c.JSON(resp)
}

// RestockProductWarehouse godoc
// @Summary Пополнить склад
// @Description Добавляет указанное количество товара на склад
// @Tags ProductWarehouse
// @Accept json
// @Produce json
// @Param id path int true "ProductWarehouse ID"
// @Param quantity body int true "Количество для пополнения" minimum(1)
// @Success 200 {object} response.ProductWarehouseResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /product-warehouses/{id}/restock [post]
func (h *ProductWarehouseHandler) Restock(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var warehouse models.ProductWarehouse
	if err := h.db.First(&warehouse, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Запись склада не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить запись склада"},
		)
	}

	var body struct {
		Quantity uint `json:"quantity" validate:"required,min=1"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат тела запроса"},
		)
	}

	if err := validate.Struct(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: err.Error()},
		)
	}

	warehouse.Count += body.Quantity

	if err := h.db.Save(&warehouse).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось пополнить склад"},
		)
	}

	h.db.Preload("Product").First(&warehouse, warehouse.ID)

	resp := response.ProductWarehouseResponse{
		ID:        warehouse.ID,
		Count:     warehouse.Count,
		ProductID: warehouse.ProductID,
		CreatedAt: warehouse.CreatedAt,
		UpdatedAt: warehouse.UpdatedAt,
	}

	if warehouse.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    warehouse.Product.ID,
			Slug:  warehouse.Product.Slug,
			Title: warehouse.Product.Title,
			Price: warehouse.Product.Price,
		}
	}

	return c.JSON(resp)
}

// DeleteProductWarehouse godoc
// @Summary Удалить запись склада
// @Description Удаляет запись склада из БД по ID
// @Tags ProductWarehouse
// @Param id path int true "ProductWarehouse ID"
// @Success 204 "Запись склада успешно удалена"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /product-warehouses/{id} [delete]
func (h *ProductWarehouseHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var warehouse models.ProductWarehouse
	if err := h.db.First(&warehouse, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Запись склада не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить запись склада"},
		)
	}

	if err := h.db.Delete(&warehouse).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить запись склада"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
