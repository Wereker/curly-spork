// internal/handlers/attribute_value.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/catalog"
	"app/internal/response"
)

type AttributeValueHandler struct {
	db *gorm.DB
}

func NewAttributeValueHandler(db *gorm.DB) *AttributeValueHandler {
	return &AttributeValueHandler{db: db}
}

// GetAllAttributeValues godoc
// @Summary Получить список значений атрибутов
// @Description Возвращает список всех значений атрибутов с пагинацией
// @Tags AttributeValue
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param product_id query int false "Фильтр по товару"
// @Param attribute_id query int false "Фильтр по атрибуту"
// @Param search query string false "Поиск по значению"
// @Success 200 {object} response.AttributeValueListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /attribute-values [get]
func (h *AttributeValueHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	productID, _ := strconv.ParseUint(c.Query("product_id"), 10, 32)
	attributeID, _ := strconv.ParseUint(c.Query("attribute_id"), 10, 32)
	search := c.Query("search")
	offset := (page - 1) * limit

	query := h.db.Model(&models.AttributeValue{}).
		Preload("Product").
		Preload("Attribute")

	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}
	if attributeID > 0 {
		query = query.Where("attribute_id = ?", attributeID)
	}
	if search != "" {
		query = query.Where("value ILIKE ?", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество значений атрибутов"},
		)
	}

	var values []models.AttributeValue
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&values).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить значения атрибутов"},
		)
	}

	data := make([]response.AttributeValueResponse, 0, len(values))
	for _, value := range values {
		resp := response.AttributeValueResponse{
			ID:          value.ID,
			Value:       value.Value,
			ProductID:   value.ProductID,
			AttributeID: value.AttributeID,
			CreatedAt:   value.CreatedAt,
			UpdatedAt:   value.UpdatedAt,
		}

		if value.Product != nil {
			resp.Product = &response.ProductShortResponse{
				ID:    value.Product.ID,
				Slug:  value.Product.Slug,
				Title: value.Product.Title,
			}
		}

		if value.Attribute != nil {
			resp.Attribute = &response.AttributeShortResponse{
				ID:            value.Attribute.ID,
				Slug:          value.Attribute.Slug,
				Title:         value.Attribute.Title,
				AttributeType: value.Attribute.AttributeType,
			}
		}

		data = append(data, resp)
	}

	return c.JSON(response.AttributeValueListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetAttributeValueByID godoc
// @Summary Получить значение атрибута по ID
// @Description Возвращает значение атрибута по ID
// @Tags AttributeValue
// @Param id path int true "AttributeValue ID"
// @Success 200 {object} response.AttributeValueResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /attribute-values/{id} [get]
func (h *AttributeValueHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var value models.AttributeValue
	if err := h.db.Preload("Product").
		Preload("Attribute").
		First(&value, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Значение атрибута не найдено"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить значение атрибута"},
		)
	}

	resp := response.AttributeValueResponse{
		ID:          value.ID,
		Value:       value.Value,
		ProductID:   value.ProductID,
		AttributeID: value.AttributeID,
		CreatedAt:   value.CreatedAt,
		UpdatedAt:   value.UpdatedAt,
	}

	if value.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    value.Product.ID,
			Slug:  value.Product.Slug,
			Title: value.Product.Title,
		}
	}

	if value.Attribute != nil {
		resp.Attribute = &response.AttributeShortResponse{
			ID:            value.Attribute.ID,
			Slug:          value.Attribute.Slug,
			Title:         value.Attribute.Title,
			AttributeType: value.Attribute.AttributeType,
		}
	}

	return c.JSON(resp)
}

// CreateAttributeValue godoc
// @Summary Создать значение атрибута
// @Description Создает новое значение атрибута
// @Tags AttributeValue
// @Accept json
// @Produce json
// @Param data body response.AttributeValueCreate true "Данные для создания значения атрибута"
// @Success 201 {object} response.AttributeValueResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /attribute-values [post]
func (h *AttributeValueHandler) Create(c *fiber.Ctx) error {
	var req response.AttributeValueCreate

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

	var attribute models.Attribute
	if err := h.db.First(&attribute, req.AttributeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Атрибут не найден"},
		)
	}

	var existing models.AttributeValue
	if err := h.db.Where("product_id = ? AND attribute_id = ?", req.ProductID, req.AttributeID).
		First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Значение атрибута для этого товара уже существует"},
		)
	}

	value := models.AttributeValue{
		Value:       req.Value,
		ProductID:   req.ProductID,
		AttributeID: req.AttributeID,
	}

	if err := h.db.Create(&value).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать значение атрибута"},
		)
	}

	h.db.Preload("Product").Preload("Attribute").First(&value, value.ID)

	resp := response.AttributeValueResponse{
		ID:          value.ID,
		Value:       value.Value,
		ProductID:   value.ProductID,
		AttributeID: value.AttributeID,
		CreatedAt:   value.CreatedAt,
		UpdatedAt:   value.UpdatedAt,
	}

	if value.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    value.Product.ID,
			Slug:  value.Product.Slug,
			Title: value.Product.Title,
		}
	}

	if value.Attribute != nil {
		resp.Attribute = &response.AttributeShortResponse{
			ID:            value.Attribute.ID,
			Slug:          value.Attribute.Slug,
			Title:         value.Attribute.Title,
			AttributeType: value.Attribute.AttributeType,
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateAttributeValue godoc
// @Summary Полностью обновить значение атрибута
// @Description Обновляет все поля значения атрибута по ID (PUT)
// @Tags AttributeValue
// @Accept json
// @Produce json
// @Param id path int true "AttributeValue ID"
// @Param data body response.AttributeValueUpdate true "Обновляемые поля значения атрибута"
// @Success 200 {object} response.AttributeValueResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /attribute-values/{id} [put]
func (h *AttributeValueHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var value models.AttributeValue
	if err := h.db.First(&value, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Значение атрибута не найдено"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить значение атрибута"},
		)
	}

	var req response.AttributeValueUpdate
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

	var attribute models.Attribute
	if err := h.db.First(&attribute, req.AttributeID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Атрибут не найден"},
		)
	}

	if value.ProductID != req.ProductID || value.AttributeID != req.AttributeID {
		var existing models.AttributeValue
		if err := h.db.Where("product_id = ? AND attribute_id = ? AND id != ?",
			req.ProductID, req.AttributeID, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Значение атрибута для этого товара уже существует"},
			)
		}
	}

	value.Value = req.Value
	value.ProductID = req.ProductID
	value.AttributeID = req.AttributeID

	if err := h.db.Save(&value).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить значение атрибута"},
		)
	}

	h.db.Preload("Product").Preload("Attribute").First(&value, value.ID)

	resp := response.AttributeValueResponse{
		ID:          value.ID,
		Value:       value.Value,
		ProductID:   value.ProductID,
		AttributeID: value.AttributeID,
		CreatedAt:   value.CreatedAt,
		UpdatedAt:   value.UpdatedAt,
	}

	if value.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    value.Product.ID,
			Slug:  value.Product.Slug,
			Title: value.Product.Title,
		}
	}

	if value.Attribute != nil {
		resp.Attribute = &response.AttributeShortResponse{
			ID:            value.Attribute.ID,
			Slug:          value.Attribute.Slug,
			Title:         value.Attribute.Title,
			AttributeType: value.Attribute.AttributeType,
		}
	}

	return c.JSON(resp)
}

// PatchAttributeValue godoc
// @Summary Частично обновить значение атрибута
// @Description Обновляет указанные поля значения атрибута по ID (PATCH)
// @Tags AttributeValue
// @Accept json
// @Produce json
// @Param id path int true "AttributeValue ID"
// @Param data body response.AttributeValuePatch true "Обновляемые поля значения атрибута"
// @Success 200 {object} response.AttributeValueResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /attribute-values/{id} [patch]
func (h *AttributeValueHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var value models.AttributeValue
	if err := h.db.First(&value, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Значение атрибута не найдено"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить значение атрибута"},
		)
	}

	var req response.AttributeValuePatch
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

	if req.Value != nil {
		value.Value = *req.Value
	}
	if req.ProductID != nil {
		var product models.Product
		if err := h.db.First(&product, *req.ProductID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Товар не найден"},
			)
		}

		if value.ProductID != *req.ProductID || (req.AttributeID == nil && value.AttributeID != 0) {
			var existing models.AttributeValue
			checkAttrID := value.AttributeID
			if req.AttributeID != nil {
				checkAttrID = *req.AttributeID
			}
			if err := h.db.Where("product_id = ? AND attribute_id = ? AND id != ?",
				*req.ProductID, checkAttrID, id).First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Значение атрибута для этого товара уже существует"},
				)
			}
		}
		value.ProductID = *req.ProductID
	}
	if req.AttributeID != nil {
		var attribute models.Attribute
		if err := h.db.First(&attribute, *req.AttributeID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Атрибут не найден"},
			)
		}

		if value.AttributeID != *req.AttributeID || (req.ProductID == nil && value.ProductID != 0) {
			var existing models.AttributeValue
			checkProductID := value.ProductID
			if req.ProductID != nil {
				checkProductID = *req.ProductID
			}
			if err := h.db.Where("product_id = ? AND attribute_id = ? AND id != ?",
				checkProductID, *req.AttributeID, id).First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Значение атрибута для этого товара уже существует"},
				)
			}
		}
		value.AttributeID = *req.AttributeID
	}

	if err := h.db.Save(&value).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить значение атрибута"},
		)
	}

	h.db.Preload("Product").Preload("Attribute").First(&value, value.ID)

	resp := response.AttributeValueResponse{
		ID:          value.ID,
		Value:       value.Value,
		ProductID:   value.ProductID,
		AttributeID: value.AttributeID,
		CreatedAt:   value.CreatedAt,
		UpdatedAt:   value.UpdatedAt,
	}

	if value.Product != nil {
		resp.Product = &response.ProductShortResponse{
			ID:    value.Product.ID,
			Slug:  value.Product.Slug,
			Title: value.Product.Title,
		}
	}

	if value.Attribute != nil {
		resp.Attribute = &response.AttributeShortResponse{
			ID:            value.Attribute.ID,
			Slug:          value.Attribute.Slug,
			Title:         value.Attribute.Title,
			AttributeType: value.Attribute.AttributeType,
		}
	}

	return c.JSON(resp)
}

// DeleteAttributeValue godoc
// @Summary Удалить значение атрибута
// @Description Удаляет значение атрибута из БД по ID
// @Tags AttributeValue
// @Param id path int true "AttributeValue ID"
// @Success 204 "Значение атрибута успешно удалено"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /attribute-values/{id} [delete]
func (h *AttributeValueHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var value models.AttributeValue
	if err := h.db.First(&value, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Значение атрибута не найдено"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить значение атрибута"},
		)
	}

	if err := h.db.Delete(&value).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить значение атрибута"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
