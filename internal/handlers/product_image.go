// internal/handlers/product_image.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/catalog"
	"app/internal/response"
)

type ProductImageHandler struct {
	db *gorm.DB
}

func NewProductImageHandler(db *gorm.DB) *ProductImageHandler {
	return &ProductImageHandler{db: db}
}

// GetAllProductImages godoc
// @Summary Получить список изображений товаров
// @Description Возвращает список всех изображений товаров с пагинацией
// @Tags ProductImage
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param product_id query int false "Фильтр по товару"
// @Param is_main query bool false "Фильтр по главному изображению"
// @Success 200 {object} response.ProductImageListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /product-images [get]
func (h *ProductImageHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	productID, _ := strconv.ParseUint(c.Query("product_id"), 10, 32)
	isMain, _ := strconv.ParseBool(c.Query("is_main"))
	offset := (page - 1) * limit

	query := h.db.Model(&models.ProductImage{}).Preload("Product")

	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}
	if c.Query("is_main") != "" {
		query = query.Where("is_main = ?", isMain)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество изображений товаров"},
		)
	}

	var images []models.ProductImage
	if err := query.Order("is_main DESC, created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&images).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить изображения товаров"},
		)
	}

	data := make([]response.ProductImageResponse, 0, len(images))
	for _, image := range images {
		resp := response.ProductImageResponse{
			ID:        image.ID,
			ImageURL:  image.ImageURL,
			AltText:   image.AltText,
			IsMain:    image.IsMain,
			ProductID: image.ProductID,
			CreatedAt: image.CreatedAt,
			UpdatedAt: image.UpdatedAt,
		}

		if image.Product != nil {
			resp.Product = &response.ProductShortResponse{
				ID:    image.Product.ID,
				Slug:  image.Product.Slug,
				Title: image.Product.Title,
			}
		}

		data = append(data, resp)
	}

	return c.JSON(response.ProductImageListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetProductImageByID godoc
// @Summary Получить изображение товара по ID
// @Description Возвращает изображение товара по ID
// @Tags ProductImage
// @Param id path int true "ProductImage ID"
// @Success 200 {object} response.ProductImageResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /product-images/{id} [get]
func (h *ProductImageHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var image models.ProductImage
	if err := h.db.Preload("Product").
		First(&image, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Изображение товара не найдено"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить изображение товара"},
		)
	}

	resp := response.ProductImageResponse{
		ID:        image.ID,
		ImageURL:  image.ImageURL,
		AltText:   image.AltText,
		IsMain:    image.IsMain,
		ProductID: image.ProductID,
		CreatedAt: image.CreatedAt,
		UpdatedAt: image.UpdatedAt,
	}

	if image.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    image.Product.ID,
			Slug:  image.Product.Slug,
			Title: image.Product.Title,
		}
	}

	return c.JSON(resp)
}

// CreateProductImage godoc
// @Summary Создать изображение товара
// @Description Создает новое изображение товара
// @Tags ProductImage
// @Accept json
// @Produce json
// @Param data body response.ProductImageCreate true "Данные для создания изображения товара"
// @Success 201 {object} response.ProductImageResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /product-images [post]
func (h *ProductImageHandler) Create(c *fiber.Ctx) error {
	var req response.ProductImageCreate

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

	if req.IsMain {
		// Снимаем флаг главного изображения у других изображений этого товара
		if err := h.db.Model(&models.ProductImage{}).
			Where("product_id = ? AND is_main = ?", req.ProductID, true).
			Update("is_main", false).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				response.ErrorResponse{Error: "Не удалось обновить статус главного изображения"},
			)
		}
	}

	image := models.ProductImage{
		ImageURL:  req.ImageURL,
		AltText:   req.AltText,
		IsMain:    req.IsMain,
		ProductID: req.ProductID,
	}

	if err := h.db.Create(&image).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать изображение товара"},
		)
	}

	h.db.Preload("Product").First(&image, image.ID)

	resp := response.ProductImageResponse{
		ID:        image.ID,
		ImageURL:  image.ImageURL,
		AltText:   image.AltText,
		IsMain:    image.IsMain,
		ProductID: image.ProductID,
		CreatedAt: image.CreatedAt,
		UpdatedAt: image.UpdatedAt,
	}

	if image.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    image.Product.ID,
			Slug:  image.Product.Slug,
			Title: image.Product.Title,
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateProductImage godoc
// @Summary Полностью обновить изображение товара
// @Description Обновляет все поля изображения товара по ID (PUT)
// @Tags ProductImage
// @Accept json
// @Produce json
// @Param id path int true "ProductImage ID"
// @Param data body response.ProductImageUpdate true "Обновляемые поля изображения товара"
// @Success 200 {object} response.ProductImageResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /product-images/{id} [put]
func (h *ProductImageHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var image models.ProductImage
	if err := h.db.First(&image, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Изображение товара не найдено"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить изображение товара"},
		)
	}

	var req response.ProductImageUpdate
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

	if req.IsMain && !image.IsMain {
		// Снимаем флаг главного изображения у других изображений этого товара
		if err := h.db.Model(&models.ProductImage{}).
			Where("product_id = ? AND is_main = ? AND id != ?", req.ProductID, true, id).
			Update("is_main", false).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				response.ErrorResponse{Error: "Не удалось обновить статус главного изображения"},
			)
		}
	}

	image.ImageURL = req.ImageURL
	image.AltText = req.AltText
	image.IsMain = req.IsMain
	image.ProductID = req.ProductID

	if err := h.db.Save(&image).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить изображение товара"},
		)
	}

	h.db.Preload("Product").First(&image, image.ID)

	resp := response.ProductImageResponse{
		ID:        image.ID,
		ImageURL:  image.ImageURL,
		AltText:   image.AltText,
		IsMain:    image.IsMain,
		ProductID: image.ProductID,
		CreatedAt: image.CreatedAt,
		UpdatedAt: image.UpdatedAt,
	}

	if image.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    image.Product.ID,
			Slug:  image.Product.Slug,
			Title: image.Product.Title,
		}
	}

	return c.JSON(resp)
}

// PatchProductImage godoc
// @Summary Частично обновить изображение товара
// @Description Обновляет указанные поля изображения товара по ID (PATCH)
// @Tags ProductImage
// @Accept json
// @Produce json
// @Param id path int true "ProductImage ID"
// @Param data body response.ProductImagePatch true "Обновляемые поля изображения товара"
// @Success 200 {object} response.ProductImageResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /product-images/{id} [patch]
func (h *ProductImageHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var image models.ProductImage
	if err := h.db.First(&image, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Изображение товара не найдено"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить изображение товара"},
		)
	}

	var req response.ProductImagePatch
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

	if req.ImageURL != nil {
		image.ImageURL = *req.ImageURL
	}
	if req.AltText != nil {
		image.AltText = *req.AltText
	}
	if req.IsMain != nil {
		if *req.IsMain && !image.IsMain {
			// Снимаем флаг главного изображения у других изображений этого товара
			productID := image.ProductID
			if req.ProductID != nil {
				productID = *req.ProductID
			}
			if err := h.db.Model(&models.ProductImage{}).
				Where("product_id = ? AND is_main = ? AND id != ?", productID, true, id).
				Update("is_main", false).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(
					response.ErrorResponse{Error: "Не удалось обновить статус главного изображения"},
				)
			}
		}
		image.IsMain = *req.IsMain
	}
	if req.ProductID != nil {
		var product models.Product
		if err := h.db.First(&product, *req.ProductID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Товар не найден"},
			)
		}
		image.ProductID = *req.ProductID
	}

	if err := h.db.Save(&image).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить изображение товара"},
		)
	}

	h.db.Preload("Product").First(&image, image.ID)

	resp := response.ProductImageResponse{
		ID:        image.ID,
		ImageURL:  image.ImageURL,
		AltText:   image.AltText,
		IsMain:    image.IsMain,
		ProductID: image.ProductID,
		CreatedAt: image.CreatedAt,
		UpdatedAt: image.UpdatedAt,
	}

	if image.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    image.Product.ID,
			Slug:  image.Product.Slug,
			Title: image.Product.Title,
		}
	}

	return c.JSON(resp)
}

// DeleteProductImage godoc
// @Summary Удалить изображение товара
// @Description Удаляет изображение товара из БД по ID
// @Tags ProductImage
// @Param id path int true "ProductImage ID"
// @Success 204 "Изображение товара успешно удалено"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /product-images/{id} [delete]
func (h *ProductImageHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var image models.ProductImage
	if err := h.db.First(&image, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Изображение товара не найдено"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить изображение товара"},
		)
	}

	if image.IsMain {
		// Ищем другое изображение для этого товара, чтобы сделать его главным
		var otherImage models.ProductImage
		if err := h.db.Where("product_id = ? AND id != ?", image.ProductID, id).
			First(&otherImage).Error; err == nil {
			// Делаем первое найденное изображение главным
			otherImage.IsMain = true
			if err := h.db.Save(&otherImage).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(
					response.ErrorResponse{Error: "Не удалось назначить новое главное изображение"},
				)
			}
		}
	}

	if err := h.db.Delete(&image).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить изображение товара"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
