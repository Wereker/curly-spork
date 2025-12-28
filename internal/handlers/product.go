// internal/handlers/product.go
package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/catalog"
	order "app/internal/models/order"
	"app/internal/response"
)

type ProductHandler struct {
	db *gorm.DB
}

func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{db: db}
}

// GetAllProducts godoc
// @Summary Получить список товаров
// @Description Возвращает список всех товаров с пагинацией и фильтрацией
// @Tags Product
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по названию или slug"
// @Param category_id query int false "Фильтр по категории"
// @Param manufacturer_id query int false "Фильтр по производителю"
// @Param min_price query number false "Минимальная цена"
// @Param max_price query number false "Максимальная цена"
// @Param with_discount query bool false "Только со скидкой" default(false)
// @Success 200 {object} response.ProductListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products [get]
func (h *ProductHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	categoryID, _ := strconv.ParseUint(c.Query("category_id"), 10, 32)
	manufacturerID, _ := strconv.ParseUint(c.Query("manufacturer_id"), 10, 32)
	minPrice, _ := strconv.ParseFloat(c.Query("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("max_price"), 64)
	withDiscount, _ := strconv.ParseBool(c.Query("with_discount", "false"))
	offset := (page - 1) * limit

	query := h.db.Model(&models.Product{}).
		Preload("Category").
		Preload("Manufacturer").
		Preload("ProductImages").
		Preload("ProductWarehouse")

	if search != "" {
		query = query.Where("title ILIKE ? OR slug ILIKE ? OR short_description ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if manufacturerID > 0 {
		query = query.Where("manufacturer_id = ?", manufacturerID)
	}
	if minPrice > 0 {
		query = query.Where("price >= ?", minPrice)
	}
	if maxPrice > 0 {
		query = query.Where("price <= ?", maxPrice)
	}
	if withDiscount {
		query = query.Where("discount > 0")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество товаров"},
		)
	}

	var products []models.Product
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить товары"},
		)
	}

	data := make([]response.ProductResponse, 0, len(products))
	for _, product := range products {
		discountedPrice := product.Price
		if product.Discount != nil && *product.Discount > 0 {
			discountedPrice = product.Price * (100 - float64(*product.Discount)) / 100
		}

		resp := response.ProductResponse{
			ID:               product.ID,
			Slug:             product.Slug,
			Title:            product.Title,
			ShortDescription: product.ShortDescription,
			Description:      product.Description,
			Price:            product.Price,
			Discount:         product.Discount,
			DiscountedPrice:  discountedPrice,
			CategoryID:       product.CategoryID,
			ManufacturerID:   product.ManufacturerID,
			CreatedAt:        product.CreatedAt,
			UpdatedAt:        product.UpdatedAt,
		}

		if product.Category != nil {
			resp.Category = &response.CategoryShortResponse{
				ID:    product.Category.ID,
				Slug:  product.Category.Slug,
				Title: product.Category.Title,
			}
		}

		if product.Manufacturer != nil {
			resp.Manufacturer = &response.ManufacturerShortResponse{
				ID:    product.Manufacturer.ID,
				Slug:  product.Manufacturer.Slug,
				Title: product.Manufacturer.Title,
			}
		}

		if len(product.ProductImages) > 0 {
			resp.Images = make([]response.ProductImageResponse, 0, len(product.ProductImages))
			for _, image := range product.ProductImages {
				resp.Images = append(resp.Images, response.ProductImageResponse{
					ID:       image.ID,
					ImageURL: image.ImageURL,
					IsMain:   image.IsMain,
				})
			}
		}

		if product.ProductWarehouse != nil {
			resp.Warehouse = &response.ProductWarehouseResponse{}
		}

		data = append(data, resp)
	}

	return c.JSON(response.ProductListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetProductByID godoc
// @Summary Получить товар по ID
// @Description Возвращает товар по ID со всей информацией
// @Tags Product
// @Param id path int true "Product ID"
// @Success 200 {object} response.ProductResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /products/{id} [get]
func (h *ProductHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var product models.Product
	if err := h.db.Preload("Category").
		Preload("Manufacturer").
		Preload("ProductImages").
		Preload("ProductWarehouse").
		Preload("AttributeValues.Attribute").
		First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Товар не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить товар"},
		)
	}

	discountedPrice := product.Price
	if product.Discount != nil && *product.Discount > 0 {
		discountedPrice = product.Price * (100 - float64(*product.Discount)) / 100
	}

	resp := response.ProductResponse{
		ID:               product.ID,
		Slug:             product.Slug,
		Title:            product.Title,
		ShortDescription: product.ShortDescription,
		Description:      product.Description,
		Price:            product.Price,
		Discount:         product.Discount,
		DiscountedPrice:  discountedPrice,
		CategoryID:       product.CategoryID,
		ManufacturerID:   product.ManufacturerID,
		CreatedAt:        product.CreatedAt,
		UpdatedAt:        product.UpdatedAt,
	}

	if product.Category != nil {
		resp.Category = &response.CategoryShortResponse{
			ID:    product.Category.ID,
			Slug:  product.Category.Slug,
			Title: product.Category.Title,
		}
	}

	if product.Manufacturer != nil {
		resp.Manufacturer = &response.ManufacturerShortResponse{
			ID:    product.Manufacturer.ID,
			Slug:  product.Manufacturer.Slug,
			Title: product.Manufacturer.Title,
		}
	}

	if len(product.ProductImages) > 0 {
		resp.Images = make([]response.ProductImageResponse, 0, len(product.ProductImages))
		for _, image := range product.ProductImages {
			resp.Images = append(resp.Images, response.ProductImageResponse{
				ID:       image.ID,
				ImageURL: image.ImageURL,
				IsMain:   image.IsMain,
			})
		}
	}

	if product.ProductWarehouse != nil {
		resp.Warehouse = &response.ProductWarehouseResponse{}
	}

	if len(product.AttributeValues) > 0 {
		resp.Attributes = make([]response.ProductAttributeResponse, 0, len(product.AttributeValues))
		for _, av := range product.AttributeValues {
			if av.Attribute != nil {
				resp.Attributes = append(resp.Attributes, response.ProductAttributeResponse{
					AttributeID:    av.AttributeID,
					AttributeSlug:  av.Attribute.Slug,
					AttributeTitle: av.Attribute.Title,
					AttributeType:  av.Attribute.AttributeType,
					Value:          av.Value,
				})
			}
		}
	}

	return c.JSON(resp)
}

// GetProductBySlug godoc
// @Summary Получить товар по slug
// @Description Возвращает товар по slug со всей информацией
// @Tags Product
// @Param slug path string true "Product Slug"
// @Success 200 {object} response.ProductResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /products/slug/{slug} [get]
func (h *ProductHandler) GetBySlug(c *fiber.Ctx) error {
	slug := strings.TrimSpace(c.Params("slug"))
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Slug не может быть пустым"},
		)
	}

	var product models.Product
	if err := h.db.Preload("Category").
		Preload("Manufacturer").
		Preload("ProductImages").
		Preload("ProductWarehouse").
		Preload("AttributeValues.Attribute").
		Where("slug = ?", slug).
		First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Товар не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить товар"},
		)
	}

	discountedPrice := product.Price
	if product.Discount != nil && *product.Discount > 0 {
		discountedPrice = product.Price * (100 - float64(*product.Discount)) / 100
	}

	resp := response.ProductResponse{
		ID:               product.ID,
		Slug:             product.Slug,
		Title:            product.Title,
		ShortDescription: product.ShortDescription,
		Description:      product.Description,
		Price:            product.Price,
		Discount:         product.Discount,
		DiscountedPrice:  discountedPrice,
		CategoryID:       product.CategoryID,
		ManufacturerID:   product.ManufacturerID,
		CreatedAt:        product.CreatedAt,
		UpdatedAt:        product.UpdatedAt,
	}

	if product.Category != nil {
		resp.Category = &response.CategoryShortResponse{
			ID:    product.Category.ID,
			Slug:  product.Category.Slug,
			Title: product.Category.Title,
		}
	}

	if product.Manufacturer != nil {
		resp.Manufacturer = &response.ManufacturerShortResponse{
			ID:    product.Manufacturer.ID,
			Slug:  product.Manufacturer.Slug,
			Title: product.Manufacturer.Title,
		}
	}

	if len(product.ProductImages) > 0 {
		resp.Images = make([]response.ProductImageResponse, 0, len(product.ProductImages))
		for _, image := range product.ProductImages {
			resp.Images = append(resp.Images, response.ProductImageResponse{
				ID:       image.ID,
				ImageURL: image.ImageURL,
				IsMain:   image.IsMain,
			})
		}
	}

	if product.ProductWarehouse != nil {
		resp.Warehouse = &response.ProductWarehouseResponse{}
	}

	if len(product.AttributeValues) > 0 {
		resp.Attributes = make([]response.ProductAttributeResponse, 0, len(product.AttributeValues))
		for _, av := range product.AttributeValues {
			if av.Attribute != nil {
				resp.Attributes = append(resp.Attributes, response.ProductAttributeResponse{
					AttributeID:    av.AttributeID,
					AttributeSlug:  av.Attribute.Slug,
					AttributeTitle: av.Attribute.Title,
					AttributeType:  av.Attribute.AttributeType,
					Value:          av.Value,
				})
			}
		}
	}

	return c.JSON(resp)
}

// CreateProduct godoc
// @Summary Создать товар
// @Description Создает новый товар
// @Tags Product
// @Accept json
// @Produce json
// @Param data body response.ProductCreate true "Данные для создания товара"
// @Success 201 {object} response.ProductResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products [post]
func (h *ProductHandler) Create(c *fiber.Ctx) error {
	var req response.ProductCreate

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

	var existing models.Product
	if err := h.db.Where("slug = ?", req.Slug).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Товар с таким slug уже существует"},
		)
	}

	var category models.Category
	if err := h.db.First(&category, req.CategoryID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Категория не найдена"},
		)
	}

	var manufacturer models.Manufacturer
	if err := h.db.First(&manufacturer, req.ManufacturerID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Производитель не найден"},
		)
	}

	product := models.Product{
		Slug:             req.Slug,
		Title:            req.Title,
		ShortDescription: req.ShortDescription,
		Description:      req.Description,
		Price:            req.Price,
		Discount:         req.Discount,
		CategoryID:       req.CategoryID,
		ManufacturerID:   req.ManufacturerID,
	}

	if err := h.db.Create(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать товар"},
		)
	}

	h.db.Preload("Category").
		Preload("Manufacturer").
		First(&product, product.ID)

	discountedPrice := product.Price
	if product.Discount != nil && *product.Discount > 0 {
		discountedPrice = product.Price * (100 - float64(*product.Discount)) / 100
	}

	resp := response.ProductResponse{
		ID:               product.ID,
		Slug:             product.Slug,
		Title:            product.Title,
		ShortDescription: product.ShortDescription,
		Description:      product.Description,
		Price:            product.Price,
		Discount:         product.Discount,
		DiscountedPrice:  discountedPrice,
		CategoryID:       product.CategoryID,
		ManufacturerID:   product.ManufacturerID,
		CreatedAt:        product.CreatedAt,
		UpdatedAt:        product.UpdatedAt,
	}

	if product.Category != nil {
		resp.Category = &response.CategoryShortResponse{
			ID:    product.Category.ID,
			Slug:  product.Category.Slug,
			Title: product.Category.Title,
		}
	}

	if product.Manufacturer != nil {
		resp.Manufacturer = &response.ManufacturerShortResponse{
			ID:    product.Manufacturer.ID,
			Slug:  product.Manufacturer.Slug,
			Title: product.Manufacturer.Title,
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateProduct godoc
// @Summary Полностью обновить товар
// @Description Обновляет все поля товара по ID (PUT)
// @Tags Product
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param data body response.ProductUpdate true "Обновляемые поля товара"
// @Success 200 {object} response.ProductResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products/{id} [put]
func (h *ProductHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var product models.Product
	if err := h.db.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Товар не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить товар"},
		)
	}

	var req response.ProductUpdate
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

	if product.Slug != req.Slug {
		var existing models.Product
		if err := h.db.Where("slug = ? AND id != ?", req.Slug, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Товар с таким slug уже существует"},
			)
		}
	}

	var category models.Category
	if err := h.db.First(&category, req.CategoryID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Категория не найдена"},
		)
	}

	var manufacturer models.Manufacturer
	if err := h.db.First(&manufacturer, req.ManufacturerID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Производитель не найден"},
		)
	}

	product.Slug = req.Slug
	product.Title = req.Title
	product.ShortDescription = req.ShortDescription
	product.Description = req.Description
	product.Price = req.Price
	product.Discount = req.Discount
	product.CategoryID = req.CategoryID
	product.ManufacturerID = req.ManufacturerID

	if err := h.db.Save(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить товар"},
		)
	}

	h.db.Preload("Category").
		Preload("Manufacturer").
		First(&product, product.ID)

	discountedPrice := product.Price
	if product.Discount != nil && *product.Discount > 0 {
		discountedPrice = product.Price * (100 - float64(*product.Discount)) / 100
	}

	resp := response.ProductResponse{
		ID:               product.ID,
		Slug:             product.Slug,
		Title:            product.Title,
		ShortDescription: product.ShortDescription,
		Description:      product.Description,
		Price:            product.Price,
		Discount:         product.Discount,
		DiscountedPrice:  discountedPrice,
		CategoryID:       product.CategoryID,
		ManufacturerID:   product.ManufacturerID,
		CreatedAt:        product.CreatedAt,
		UpdatedAt:        product.UpdatedAt,
	}

	if product.Category != nil {
		resp.Category = &response.CategoryShortResponse{
			ID:    product.Category.ID,
			Slug:  product.Category.Slug,
			Title: product.Category.Title,
		}
	}

	if product.Manufacturer != nil {
		resp.Manufacturer = &response.ManufacturerShortResponse{
			ID:    product.Manufacturer.ID,
			Slug:  product.Manufacturer.Slug,
			Title: product.Manufacturer.Title,
		}
	}

	return c.JSON(resp)
}

// PatchProduct godoc
// @Summary Частично обновить товар
// @Description Обновляет указанные поля товара по ID (PATCH)
// @Tags Product
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param data body response.ProductPatch true "Обновляемые поля товара"
// @Success 200 {object} response.ProductResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products/{id} [patch]
func (h *ProductHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var product models.Product
	if err := h.db.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Товар не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить товар"},
		)
	}

	var req response.ProductPatch
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

	if req.Slug != nil {
		if product.Slug != *req.Slug {
			var existing models.Product
			if err := h.db.Where("slug = ? AND id != ?", *req.Slug, id).First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Товар с таким slug уже существует"},
				)
			}
		}
		product.Slug = *req.Slug
	}
	if req.Title != nil {
		product.Title = *req.Title
	}
	if req.ShortDescription != nil {
		product.ShortDescription = *req.ShortDescription
	}
	if req.Description != nil {
		product.Description = req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Discount != nil {
		product.Discount = req.Discount
	}
	if req.CategoryID != nil {
		var category models.Category
		if err := h.db.First(&category, *req.CategoryID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Категория не найдена"},
			)
		}
		product.CategoryID = *req.CategoryID
	}
	if req.ManufacturerID != nil {
		var manufacturer models.Manufacturer
		if err := h.db.First(&manufacturer, *req.ManufacturerID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Производитель не найден"},
			)
		}
		product.ManufacturerID = *req.ManufacturerID
	}

	if err := h.db.Save(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить товар"},
		)
	}

	h.db.Preload("Category").
		Preload("Manufacturer").
		First(&product, product.ID)

	discountedPrice := product.Price
	if product.Discount != nil && *product.Discount > 0 {
		discountedPrice = product.Price * (100 - float64(*product.Discount)) / 100
	}

	resp := response.ProductResponse{
		ID:               product.ID,
		Slug:             product.Slug,
		Title:            product.Title,
		ShortDescription: product.ShortDescription,
		Description:      product.Description,
		Price:            product.Price,
		Discount:         product.Discount,
		DiscountedPrice:  discountedPrice,
		CategoryID:       product.CategoryID,
		ManufacturerID:   product.ManufacturerID,
		CreatedAt:        product.CreatedAt,
		UpdatedAt:        product.UpdatedAt,
	}

	if product.Category != nil {
		resp.Category = &response.CategoryShortResponse{
			ID:    product.Category.ID,
			Slug:  product.Category.Slug,
			Title: product.Category.Title,
		}
	}

	if product.Manufacturer != nil {
		resp.Manufacturer = &response.ManufacturerShortResponse{
			ID:    product.Manufacturer.ID,
			Slug:  product.Manufacturer.Slug,
			Title: product.Manufacturer.Title,
		}
	}

	return c.JSON(resp)
}

// DeleteProduct godoc
// @Summary Удалить товар
// @Description Удаляет товар из БД по ID
// @Tags Product
// @Param id path int true "Product ID"
// @Success 204 "Товар успешно удален"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products/{id} [delete]
func (h *ProductHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var product models.Product
	if err := h.db.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Товар не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить товар"},
		)
	}

	var orderItemsCount int64
	if err := h.db.Model(&order.OrderItem{}).Where("product_id = ?", id).Count(&orderItemsCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные заказы"},
		)
	}

	if orderItemsCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить товар, так как с ним связаны заказы"},
		)
	}

	if err := h.db.Delete(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить товар"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
