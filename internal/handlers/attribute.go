// internal/handlers/attribute.go
package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/catalog"
	"app/internal/response"
)

type AttributeHandler struct {
	db *gorm.DB
}

func NewAttributeHandler(db *gorm.DB) *AttributeHandler {
	return &AttributeHandler{db: db}
}

// GetAllAttributes godoc
// @Summary Получить список атрибутов
// @Description Возвращает список всех атрибутов с пагинацией
// @Tags Attribute
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по названию или slug"
// @Param attribute_type query string false "Фильтр по типу атрибута"
// @Success 200 {object} response.AttributeListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /attributes [get]
func (h *AttributeHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	attributeType := models.AttributeType(c.Query("attribute_type"))
	offset := (page - 1) * limit

	query := h.db.Model(&models.Attribute{}).Preload("AttributeValues")

	if search != "" {
		query = query.Where("title ILIKE ? OR slug ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if attributeType != "" {
		query = query.Where("attribute_type = ?", attributeType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество атрибутов"},
		)
	}

	var attributes []models.Attribute
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&attributes).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить атрибуты"},
		)
	}

	data := make([]response.AttributeResponse, 0, len(attributes))
	for _, attribute := range attributes {
		resp := response.AttributeResponse{
			ID:            attribute.ID,
			Slug:          attribute.Slug,
			Title:         attribute.Title,
			AttributeType: attribute.AttributeType,
			CreatedAt:     attribute.CreatedAt,
			UpdatedAt:     attribute.UpdatedAt,
		}

		if len(attribute.AttributeValues) > 0 {
			resp.Values = make([]response.AttributeValueShortResponse, 0, len(attribute.AttributeValues))
			for _, value := range attribute.AttributeValues {
				resp.Values = append(resp.Values, response.AttributeValueShortResponse{
					ID:    value.ID,
					Value: value.Value,
				})
			}
		}

		data = append(data, resp)
	}

	return c.JSON(response.AttributeListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetAttributeByID godoc
// @Summary Получить атрибут по ID
// @Description Возвращает атрибут по ID
// @Tags Attribute
// @Param id path int true "Attribute ID"
// @Success 200 {object} response.AttributeResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /attributes/{id} [get]
func (h *AttributeHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var attribute models.Attribute
	if err := h.db.Preload("AttributeValues").
		First(&attribute, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Атрибут не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить атрибут"},
		)
	}

	resp := response.AttributeResponse{
		ID:            attribute.ID,
		Slug:          attribute.Slug,
		Title:         attribute.Title,
		AttributeType: attribute.AttributeType,
		CreatedAt:     attribute.CreatedAt,
		UpdatedAt:     attribute.UpdatedAt,
	}

	if len(attribute.AttributeValues) > 0 {
		resp.Values = make([]response.AttributeValueShortResponse, 0, len(attribute.AttributeValues))
		for _, value := range attribute.AttributeValues {
			resp.Values = append(resp.Values, response.AttributeValueShortResponse{
				ID:    value.ID,
				Value: value.Value,
			})
		}
	}

	return c.JSON(resp)
}

// GetAttributeBySlug godoc
// @Summary Получить атрибут по slug
// @Description Возвращает атрибут по slug
// @Tags Attribute
// @Param slug path string true "Attribute Slug"
// @Success 200 {object} response.AttributeResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /attributes/slug/{slug} [get]
func (h *AttributeHandler) GetBySlug(c *fiber.Ctx) error {
	slug := strings.TrimSpace(c.Params("slug"))
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Slug не может быть пустым"},
		)
	}

	var attribute models.Attribute
	if err := h.db.Preload("AttributeValues").
		Where("slug = ?", slug).
		First(&attribute).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Атрибут не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить атрибут"},
		)
	}

	resp := response.AttributeResponse{
		ID:            attribute.ID,
		Slug:          attribute.Slug,
		Title:         attribute.Title,
		AttributeType: attribute.AttributeType,
		CreatedAt:     attribute.CreatedAt,
		UpdatedAt:     attribute.UpdatedAt,
	}

	if len(attribute.AttributeValues) > 0 {
		resp.Values = make([]response.AttributeValueShortResponse, 0, len(attribute.AttributeValues))
		for _, value := range attribute.AttributeValues {
			resp.Values = append(resp.Values, response.AttributeValueShortResponse{
				ID:    value.ID,
				Value: value.Value,
			})
		}
	}

	return c.JSON(resp)
}

// CreateAttribute godoc
// @Summary Создать атрибут
// @Description Создает новый атрибут
// @Tags Attribute
// @Accept json
// @Produce json
// @Param data body response.AttributeCreate true "Данные для создания атрибута"
// @Success 201 {object} response.AttributeResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /attributes [post]
func (h *AttributeHandler) Create(c *fiber.Ctx) error {
	var req response.AttributeCreate

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

	var existing models.Attribute
	if err := h.db.Where("slug = ?", req.Slug).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Атрибут с таким slug уже существует"},
		)
	}

	attribute := models.Attribute{
		Slug:          req.Slug,
		Title:         req.Title,
		AttributeType: req.AttributeType,
	}

	if err := h.db.Create(&attribute).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать атрибут"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(response.AttributeResponse{
		ID:            attribute.ID,
		Slug:          attribute.Slug,
		Title:         attribute.Title,
		AttributeType: attribute.AttributeType,
		CreatedAt:     attribute.CreatedAt,
		UpdatedAt:     attribute.UpdatedAt,
	})
}

// UpdateAttribute godoc
// @Summary Полностью обновить атрибут
// @Description Обновляет все поля атрибута по ID (PUT)
// @Tags Attribute
// @Accept json
// @Produce json
// @Param id path int true "Attribute ID"
// @Param data body response.AttributeUpdate true "Обновляемые поля атрибута"
// @Success 200 {object} response.AttributeResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /attributes/{id} [put]
func (h *AttributeHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var attribute models.Attribute
	if err := h.db.First(&attribute, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Атрибут не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить атрибут"},
		)
	}

	var req response.AttributeUpdate
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

	if attribute.Slug != req.Slug {
		var existing models.Attribute
		if err := h.db.Where("slug = ? AND id != ?", req.Slug, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Атрибут с таким slug уже существует"},
			)
		}
	}

	attribute.Slug = req.Slug
	attribute.Title = req.Title
	attribute.AttributeType = req.AttributeType

	if err := h.db.Save(&attribute).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить атрибут"},
		)
	}

	return c.JSON(response.AttributeResponse{
		ID:            attribute.ID,
		Slug:          attribute.Slug,
		Title:         attribute.Title,
		AttributeType: attribute.AttributeType,
		CreatedAt:     attribute.CreatedAt,
		UpdatedAt:     attribute.UpdatedAt,
	})
}

// PatchAttribute godoc
// @Summary Частично обновить атрибут
// @Description Обновляет указанные поля атрибута по ID (PATCH)
// @Tags Attribute
// @Accept json
// @Produce json
// @Param id path int true "Attribute ID"
// @Param data body response.AttributePatch true "Обновляемые поля атрибута"
// @Success 200 {object} response.AttributeResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /attributes/{id} [patch]
func (h *AttributeHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var attribute models.Attribute
	if err := h.db.First(&attribute, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Атрибут не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить атрибут"},
		)
	}

	var req response.AttributePatch
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
		if attribute.Slug != *req.Slug {
			var existing models.Attribute
			if err := h.db.Where("slug = ? AND id != ?", *req.Slug, id).First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Атрибут с таким slug уже существует"},
				)
			}
		}
		attribute.Slug = *req.Slug
	}
	if req.Title != nil {
		attribute.Title = *req.Title
	}
	if req.AttributeType != nil {
		attribute.AttributeType = *req.AttributeType
	}

	if err := h.db.Save(&attribute).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить атрибут"},
		)
	}

	return c.JSON(response.AttributeResponse{
		ID:            attribute.ID,
		Slug:          attribute.Slug,
		Title:         attribute.Title,
		AttributeType: attribute.AttributeType,
		CreatedAt:     attribute.CreatedAt,
		UpdatedAt:     attribute.UpdatedAt,
	})
}

// DeleteAttribute godoc
// @Summary Удалить атрибут
// @Description Удаляет атрибут из БД по ID
// @Tags Attribute
// @Param id path int true "Attribute ID"
// @Success 204 "Атрибут успешно удален"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /attributes/{id} [delete]
func (h *AttributeHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var attribute models.Attribute
	if err := h.db.First(&attribute, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Атрибут не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить атрибут"},
		)
	}

	var valuesCount int64
	if err := h.db.Model(&models.AttributeValue{}).Where("attribute_id = ?", id).Count(&valuesCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные значения"},
		)
	}

	if valuesCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить атрибут, так как у него есть связанные значения"},
		)
	}

	var productsCount int64
	if err := h.db.Model(&models.AttributeValue{}).Where("attribute_id = ?", id).Count(&productsCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные товары"},
		)
	}

	if productsCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить атрибут, так как с ним связаны товары"},
		)
	}

	if err := h.db.Delete(&attribute).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить атрибут"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
