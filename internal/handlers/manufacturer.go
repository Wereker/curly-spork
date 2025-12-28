// internal/handlers/manufacturer.go
package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/catalog"
	"app/internal/response"
)

type ManufacturerHandler struct {
	db *gorm.DB
}

func NewManufacturerHandler(db *gorm.DB) *ManufacturerHandler {
	return &ManufacturerHandler{db: db}
}

// GetAllManufacturers godoc
// @Summary Получить список производителей
// @Description Возвращает список всех производителей с пагинацией
// @Tags Manufacturer
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по названию, slug или стране"
// @Success 200 {object} response.ManufacturerListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /manufacturers [get]
func (h *ManufacturerHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	offset := (page - 1) * limit

	query := h.db.Model(&models.Manufacturer{}).
		Select("manufacturers.*, (SELECT COUNT(*) FROM products WHERE products.manufacturer_id = manufacturers.id) as products_count")

	if search != "" {
		query = query.Where("title ILIKE ? OR slug ILIKE ? OR country ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество производителей"},
		)
	}

	var manufacturers []models.Manufacturer
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&manufacturers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить производителей"},
		)
	}

	data := make([]response.ManufacturerResponse, 0, len(manufacturers))
	for _, manufacturer := range manufacturers {
		data = append(data, response.ManufacturerResponse{
			ID:            manufacturer.ID,
			Slug:          manufacturer.Slug,
			Title:         manufacturer.Title,
			Country:       manufacturer.Country,
			Description:   manufacturer.Description,
			ProductsCount: 0,
			CreatedAt:     manufacturer.CreatedAt,
			UpdatedAt:     manufacturer.UpdatedAt,
		})
	}

	return c.JSON(response.ManufacturerListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetManufacturerByID godoc
// @Summary Получить производителя по ID
// @Description Возвращает производителя по ID
// @Tags Manufacturer
// @Param id path int true "Manufacturer ID"
// @Success 200 {object} response.ManufacturerResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /manufacturers/{id} [get]
func (h *ManufacturerHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var manufacturer models.Manufacturer
	if err := h.db.First(&manufacturer, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Производитель не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить производителя"},
		)
	}

	var productsCount int64
	h.db.Model(&models.Product{}).Where("manufacturer_id = ?", id).Count(&productsCount)

	return c.JSON(response.ManufacturerResponse{
		ID:            manufacturer.ID,
		Slug:          manufacturer.Slug,
		Title:         manufacturer.Title,
		Country:       manufacturer.Country,
		Description:   manufacturer.Description,
		ProductsCount: productsCount,
		CreatedAt:     manufacturer.CreatedAt,
		UpdatedAt:     manufacturer.UpdatedAt,
	})
}

// GetManufacturerBySlug godoc
// @Summary Получить производителя по slug
// @Description Возвращает производителя по slug
// @Tags Manufacturer
// @Param slug path string true "Manufacturer Slug"
// @Success 200 {object} response.ManufacturerResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /manufacturers/slug/{slug} [get]
func (h *ManufacturerHandler) GetBySlug(c *fiber.Ctx) error {
	slug := strings.TrimSpace(c.Params("slug"))
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Slug не может быть пустым"},
		)
	}

	var manufacturer models.Manufacturer
	if err := h.db.Where("slug = ?", slug).
		First(&manufacturer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Производитель не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить производителя"},
		)
	}

	var productsCount int64
	h.db.Model(&models.Product{}).Where("manufacturer_id = ?", manufacturer.ID).Count(&productsCount)

	return c.JSON(response.ManufacturerResponse{
		ID:            manufacturer.ID,
		Slug:          manufacturer.Slug,
		Title:         manufacturer.Title,
		Country:       manufacturer.Country,
		Description:   manufacturer.Description,
		ProductsCount: productsCount,
		CreatedAt:     manufacturer.CreatedAt,
		UpdatedAt:     manufacturer.UpdatedAt,
	})
}

// CreateManufacturer godoc
// @Summary Создать производителя
// @Description Создает нового производителя
// @Tags Manufacturer
// @Accept json
// @Produce json
// @Param data body response.ManufacturerCreate true "Данные для создания производителя"
// @Success 201 {object} response.ManufacturerResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /manufacturers [post]
func (h *ManufacturerHandler) Create(c *fiber.Ctx) error {
	var req response.ManufacturerCreate

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

	var existing models.Manufacturer
	if err := h.db.Where("slug = ?", req.Slug).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Производитель с таким slug уже существует"},
		)
	}

	manufacturer := models.Manufacturer{
		Slug:        req.Slug,
		Title:       req.Title,
		Country:     req.Country,
		Description: req.Description,
	}

	if err := h.db.Create(&manufacturer).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать производителя"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(response.ManufacturerResponse{
		ID:          manufacturer.ID,
		Slug:        manufacturer.Slug,
		Title:       manufacturer.Title,
		Country:     manufacturer.Country,
		Description: manufacturer.Description,
		CreatedAt:   manufacturer.CreatedAt,
		UpdatedAt:   manufacturer.UpdatedAt,
	})
}

// UpdateManufacturer godoc
// @Summary Полностью обновить производителя
// @Description Обновляет все поля производителя по ID (PUT)
// @Tags Manufacturer
// @Accept json
// @Produce json
// @Param id path int true "Manufacturer ID"
// @Param data body response.ManufacturerUpdate true "Обновляемые поля производителя"
// @Success 200 {object} response.ManufacturerResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /manufacturers/{id} [put]
func (h *ManufacturerHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var manufacturer models.Manufacturer
	if err := h.db.First(&manufacturer, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Производитель не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить производителя"},
		)
	}

	var req response.ManufacturerUpdate
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

	if manufacturer.Slug != req.Slug {
		var existing models.Manufacturer
		if err := h.db.Where("slug = ? AND id != ?", req.Slug, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Производитель с таким slug уже существует"},
			)
		}
	}

	manufacturer.Slug = req.Slug
	manufacturer.Title = req.Title
	manufacturer.Country = req.Country
	manufacturer.Description = req.Description

	if err := h.db.Save(&manufacturer).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить производителя"},
		)
	}

	return c.JSON(response.ManufacturerResponse{
		ID:          manufacturer.ID,
		Slug:        manufacturer.Slug,
		Title:       manufacturer.Title,
		Country:     manufacturer.Country,
		Description: manufacturer.Description,
		CreatedAt:   manufacturer.CreatedAt,
		UpdatedAt:   manufacturer.UpdatedAt,
	})
}

// PatchManufacturer godoc
// @Summary Частично обновить производителя
// @Description Обновляет указанные поля производителя по ID (PATCH)
// @Tags Manufacturer
// @Accept json
// @Produce json
// @Param id path int true "Manufacturer ID"
// @Param data body response.ManufacturerPatch true "Обновляемые поля производителя"
// @Success 200 {object} response.ManufacturerResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /manufacturers/{id} [patch]
func (h *ManufacturerHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var manufacturer models.Manufacturer
	if err := h.db.First(&manufacturer, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Производитель не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить производителя"},
		)
	}

	var req response.ManufacturerPatch
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
		if manufacturer.Slug != *req.Slug {
			var existing models.Manufacturer
			if err := h.db.Where("slug = ? AND id != ?", *req.Slug, id).First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Производитель с таким slug уже существует"},
				)
			}
		}
		manufacturer.Slug = *req.Slug
	}
	if req.Title != nil {
		manufacturer.Title = *req.Title
	}
	if req.Country != nil {
		manufacturer.Country = *req.Country
	}
	if req.Description != nil {
		manufacturer.Description = *req.Description
	}

	if err := h.db.Save(&manufacturer).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить производителя"},
		)
	}

	return c.JSON(response.ManufacturerResponse{
		ID:          manufacturer.ID,
		Slug:        manufacturer.Slug,
		Title:       manufacturer.Title,
		Country:     manufacturer.Country,
		Description: manufacturer.Description,
		CreatedAt:   manufacturer.CreatedAt,
		UpdatedAt:   manufacturer.UpdatedAt,
	})
}

// DeleteManufacturer godoc
// @Summary Удалить производителя
// @Description Удаляет производителя из БД по ID
// @Tags Manufacturer
// @Param id path int true "Manufacturer ID"
// @Success 204 "Производитель успешно удален"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /manufacturers/{id} [delete]
func (h *ManufacturerHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var manufacturer models.Manufacturer
	if err := h.db.First(&manufacturer, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Производитель не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить производителя"},
		)
	}

	var productsCount int64
	if err := h.db.Model(&models.Product{}).Where("manufacturer_id = ?", id).Count(&productsCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные товары"},
		)
	}

	if productsCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить производителя, так как с ним связаны товары"},
		)
	}

	if err := h.db.Delete(&manufacturer).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить производителя"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
