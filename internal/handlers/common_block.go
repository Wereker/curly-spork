// internal/handlers/common_block.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	catalog "app/internal/models/catalog"
	models "app/internal/models/sitecontent"
	"app/internal/response"
)

type CommonBlockHandler struct {
	db *gorm.DB
}

func NewCommonBlockHandler(db *gorm.DB) *CommonBlockHandler {
	return &CommonBlockHandler{db: db}
}

// GetAllCommonBlocks godoc
// @Summary Получить список общих блоков
// @Description Возвращает список всех общих блоков с пагинацией
// @Tags CommonBlock
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по названию"
// @Param is_show query bool false "Фильтр по отображению" default(true)
// @Param category_id query int false "Фильтр по категории"
// @Param manufacturer_id query int false "Фильтр по производителю"
// @Param attribute_id query int false "Фильтр по атрибуту"
// @Success 200 {object} response.CommonBlockListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /common-blocks [get]
func (h *CommonBlockHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	isShow, _ := strconv.ParseBool(c.Query("is_show", "true"))
	categoryID, _ := strconv.ParseUint(c.Query("category_id"), 10, 32)
	manufacturerID, _ := strconv.ParseUint(c.Query("manufacturer_id"), 10, 32)
	attributeID, _ := strconv.ParseUint(c.Query("attribute_id"), 10, 32)
	offset := (page - 1) * limit

	query := h.db.Model(&models.CommonBlock{}).
		Preload("Category").
		Preload("Manufacturer").
		Preload("Attribute")

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}
	if c.Query("is_show") != "" {
		query = query.Where("is_show = ?", isShow)
	}
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if manufacturerID > 0 {
		query = query.Where("manufacturer_id = ?", manufacturerID)
	}
	if attributeID > 0 {
		query = query.Where("attribute_id = ?", attributeID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество общих блоков"},
		)
	}

	var blocks []models.CommonBlock
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&blocks).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить общие блоки"},
		)
	}

	data := make([]response.CommonBlockResponse, 0, len(blocks))
	for _, block := range blocks {
		resp := response.CommonBlockResponse{
			ID:             block.ID,
			Title:          block.Title,
			CoverURL:       block.CoverURL,
			IsShow:         block.IsShow,
			CategoryID:     block.CategoryID,
			ManufacturerID: block.ManufacturerID,
			AttributeID:    block.AttributeID,
			CreatedAt:      block.CreatedAt,
			UpdatedAt:      block.UpdatedAt,
		}

		if block.Category != nil {
			resp.Category = &response.CategoryShortResponse{
				ID:    block.Category.ID,
				Slug:  block.Category.Slug,
				Title: block.Category.Title,
			}
		}

		if block.Manufacturer != nil {
			resp.Manufacturer = &response.ManufacturerShortResponse{
				ID:    block.Manufacturer.ID,
				Slug:  block.Manufacturer.Slug,
				Title: block.Manufacturer.Title,
			}
		}

		if block.Attribute != nil {
			resp.Attribute = &response.AttributeShortResponse{
				ID:            block.Attribute.ID,
				Slug:          block.Attribute.Slug,
				Title:         block.Attribute.Title,
				AttributeType: block.Attribute.AttributeType,
			}
		}

		data = append(data, resp)
	}

	return c.JSON(response.CommonBlockListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetCommonBlockByID godoc
// @Summary Получить общий блок по ID
// @Description Возвращает общий блок по ID
// @Tags CommonBlock
// @Param id path int true "CommonBlock ID"
// @Success 200 {object} response.CommonBlockResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /common-blocks/{id} [get]
func (h *CommonBlockHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var block models.CommonBlock
	if err := h.db.Preload("Category").
		Preload("Manufacturer").
		Preload("Attribute").
		First(&block, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Общий блок не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить общий блок"},
		)
	}

	resp := response.CommonBlockResponse{
		ID:             block.ID,
		Title:          block.Title,
		CoverURL:       block.CoverURL,
		IsShow:         block.IsShow,
		CategoryID:     block.CategoryID,
		ManufacturerID: block.ManufacturerID,
		AttributeID:    block.AttributeID,
		CreatedAt:      block.CreatedAt,
		UpdatedAt:      block.UpdatedAt,
	}

	if block.Category != nil {
		resp.Category = &response.CategoryShortResponse{
			ID:    block.Category.ID,
			Slug:  block.Category.Slug,
			Title: block.Category.Title,
		}
	}

	if block.Manufacturer != nil {
		resp.Manufacturer = &response.ManufacturerShortResponse{
			ID:    block.Manufacturer.ID,
			Slug:  block.Manufacturer.Slug,
			Title: block.Manufacturer.Title,
		}
	}

	if block.Attribute != nil {
		resp.Attribute = &response.AttributeShortResponse{
			ID:            block.Attribute.ID,
			Slug:          block.Attribute.Slug,
			Title:         block.Attribute.Title,
			AttributeType: block.Attribute.AttributeType,
		}
	}

	return c.JSON(resp)
}

// CreateCommonBlock godoc
// @Summary Создать общий блок
// @Description Создает новый общий блок
// @Tags CommonBlock
// @Accept json
// @Produce json
// @Param data body response.CommonBlockCreate true "Данные для создания общего блока"
// @Success 201 {object} response.CommonBlockResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /common-blocks [post]
func (h *CommonBlockHandler) Create(c *fiber.Ctx) error {
	var req response.CommonBlockCreate

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

	// Проверка связанных сущностей
	if req.CategoryID != nil && *req.CategoryID > 0 {
		var category catalog.Category
		if err := h.db.First(&category, *req.CategoryID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Категория не найдена"},
			)
		}
	}

	if req.ManufacturerID != nil && *req.ManufacturerID > 0 {
		var manufacturer catalog.Manufacturer
		if err := h.db.First(&manufacturer, *req.ManufacturerID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Производитель не найден"},
			)
		}
	}

	if req.AttributeID != nil && *req.AttributeID > 0 {
		var attribute catalog.Attribute
		if err := h.db.First(&attribute, *req.AttributeID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Атрибут не найден"},
			)
		}
	}

	block := models.CommonBlock{
		Title:          req.Title,
		CoverURL:       req.CoverURL,
		IsShow:         req.IsShow,
		CategoryID:     req.CategoryID,
		ManufacturerID: req.ManufacturerID,
		AttributeID:    req.AttributeID,
	}

	if err := h.db.Create(&block).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать общий блок"},
		)
	}

	h.db.Preload("Category").
		Preload("Manufacturer").
		Preload("Attribute").
		First(&block, block.ID)

	resp := response.CommonBlockResponse{
		ID:             block.ID,
		Title:          block.Title,
		CoverURL:       block.CoverURL,
		IsShow:         block.IsShow,
		CategoryID:     block.CategoryID,
		ManufacturerID: block.ManufacturerID,
		AttributeID:    block.AttributeID,
		CreatedAt:      block.CreatedAt,
		UpdatedAt:      block.UpdatedAt,
	}

	if block.Category != nil {
		resp.Category = &response.CategoryShortResponse{
			ID:    block.Category.ID,
			Slug:  block.Category.Slug,
			Title: block.Category.Title,
		}
	}

	if block.Manufacturer != nil {
		resp.Manufacturer = &response.ManufacturerShortResponse{
			ID:    block.Manufacturer.ID,
			Slug:  block.Manufacturer.Slug,
			Title: block.Manufacturer.Title,
		}
	}

	if block.Attribute != nil {
		resp.Attribute = &response.AttributeShortResponse{
			ID:            block.Attribute.ID,
			Slug:          block.Attribute.Slug,
			Title:         block.Attribute.Title,
			AttributeType: block.Attribute.AttributeType,
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateCommonBlock godoc
// @Summary Полностью обновить общий блок
// @Description Обновляет все поля общего блока по ID (PUT)
// @Tags CommonBlock
// @Accept json
// @Produce json
// @Param id path int true "CommonBlock ID"
// @Param data body response.CommonBlockUpdate true "Обновляемые поля общего блока"
// @Success 200 {object} response.CommonBlockResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /common-blocks/{id} [put]
func (h *CommonBlockHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var block models.CommonBlock
	if err := h.db.First(&block, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Общий блок не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить общий блок"},
		)
	}

	var req response.CommonBlockUpdate
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

	// Проверка связанных сущностей
	if req.CategoryID != nil && *req.CategoryID > 0 {
		var category catalog.Category
		if err := h.db.First(&category, *req.CategoryID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Категория не найдена"},
			)
		}
	}

	if req.ManufacturerID != nil && *req.ManufacturerID > 0 {
		var manufacturer catalog.Manufacturer
		if err := h.db.First(&manufacturer, *req.ManufacturerID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Производитель не найден"},
			)
		}
	}

	if req.AttributeID != nil && *req.AttributeID > 0 {
		var attribute catalog.Attribute
		if err := h.db.First(&attribute, *req.AttributeID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Атрибут не найден"},
			)
		}
	}

	block.Title = req.Title
	block.CoverURL = req.CoverURL
	block.IsShow = req.IsShow
	block.CategoryID = req.CategoryID
	block.ManufacturerID = req.ManufacturerID
	block.AttributeID = req.AttributeID

	if err := h.db.Save(&block).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить общий блок"},
		)
	}

	h.db.Preload("Category").
		Preload("Manufacturer").
		Preload("Attribute").
		First(&block, block.ID)

	resp := response.CommonBlockResponse{
		ID:             block.ID,
		Title:          block.Title,
		CoverURL:       block.CoverURL,
		IsShow:         block.IsShow,
		CategoryID:     block.CategoryID,
		ManufacturerID: block.ManufacturerID,
		AttributeID:    block.AttributeID,
		CreatedAt:      block.CreatedAt,
		UpdatedAt:      block.UpdatedAt,
	}

	if block.Category != nil {
		resp.Category = &response.CategoryShortResponse{
			ID:    block.Category.ID,
			Slug:  block.Category.Slug,
			Title: block.Category.Title,
		}
	}

	if block.Manufacturer != nil {
		resp.Manufacturer = &response.ManufacturerShortResponse{
			ID:    block.Manufacturer.ID,
			Slug:  block.Manufacturer.Slug,
			Title: block.Manufacturer.Title,
		}
	}

	if block.Attribute != nil {
		resp.Attribute = &response.AttributeShortResponse{
			ID:            block.Attribute.ID,
			Slug:          block.Attribute.Slug,
			Title:         block.Attribute.Title,
			AttributeType: block.Attribute.AttributeType,
		}
	}

	return c.JSON(resp)
}

// PatchCommonBlock godoc
// @Summary Частично обновить общий блок
// @Description Обновляет указанные поля общего блока по ID (PATCH)
// @Tags CommonBlock
// @Accept json
// @Produce json
// @Param id path int true "CommonBlock ID"
// @Param data body response.CommonBlockPatch true "Обновляемые поля общего блока"
// @Success 200 {object} response.CommonBlockResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /common-blocks/{id} [patch]
func (h *CommonBlockHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var block models.CommonBlock
	if err := h.db.First(&block, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Общий блок не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить общий блок"},
		)
	}

	var req response.CommonBlockPatch
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
		block.Title = *req.Title
	}
	if req.CoverURL != nil {
		block.CoverURL = *req.CoverURL
	}
	if req.IsShow != nil {
		block.IsShow = *req.IsShow
	}
	if req.CategoryID != nil {
		if *req.CategoryID == 0 {
			block.CategoryID = nil
		} else {
			var category catalog.Category
			if err := h.db.First(&category, *req.CategoryID).Error; err != nil {
				return c.Status(fiber.StatusNotFound).JSON(
					response.ErrorResponse{Error: "Категория не найдена"},
				)
			}
			block.CategoryID = req.CategoryID
		}
	}
	if req.ManufacturerID != nil {
		if *req.ManufacturerID == 0 {
			block.ManufacturerID = nil
		} else {
			var manufacturer catalog.Manufacturer
			if err := h.db.First(&manufacturer, *req.ManufacturerID).Error; err != nil {
				return c.Status(fiber.StatusNotFound).JSON(
					response.ErrorResponse{Error: "Производитель не найден"},
				)
			}
			block.ManufacturerID = req.ManufacturerID
		}
	}
	if req.AttributeID != nil {
		if *req.AttributeID == 0 {
			block.AttributeID = nil
		} else {
			var attribute catalog.Attribute
			if err := h.db.First(&attribute, *req.AttributeID).Error; err != nil {
				return c.Status(fiber.StatusNotFound).JSON(
					response.ErrorResponse{Error: "Атрибут не найден"},
				)
			}
			block.AttributeID = req.AttributeID
		}
	}

	if err := h.db.Save(&block).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить общий блок"},
		)
	}

	h.db.Preload("Category").
		Preload("Manufacturer").
		Preload("Attribute").
		First(&block, block.ID)

	resp := response.CommonBlockResponse{
		ID:             block.ID,
		Title:          block.Title,
		CoverURL:       block.CoverURL,
		IsShow:         block.IsShow,
		CategoryID:     block.CategoryID,
		ManufacturerID: block.ManufacturerID,
		AttributeID:    block.AttributeID,
		CreatedAt:      block.CreatedAt,
		UpdatedAt:      block.UpdatedAt,
	}

	if block.Category != nil {
		resp.Category = &response.CategoryShortResponse{
			ID:    block.Category.ID,
			Slug:  block.Category.Slug,
			Title: block.Category.Title,
		}
	}

	if block.Manufacturer != nil {
		resp.Manufacturer = &response.ManufacturerShortResponse{
			ID:    block.Manufacturer.ID,
			Slug:  block.Manufacturer.Slug,
			Title: block.Manufacturer.Title,
		}
	}

	if block.Attribute != nil {
		resp.Attribute = &response.AttributeShortResponse{
			ID:            block.Attribute.ID,
			Slug:          block.Attribute.Slug,
			Title:         block.Attribute.Title,
			AttributeType: block.Attribute.AttributeType,
		}
	}

	return c.JSON(resp)
}

// ToggleCommonBlockVisibility godoc
// @Summary Переключить видимость общего блока
// @Description Переключает видимость общего блока (is_show)
// @Tags CommonBlock
// @Param id path int true "CommonBlock ID"
// @Success 200 {object} response.CommonBlockResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /common-blocks/{id}/toggle-visibility [patch]
func (h *CommonBlockHandler) ToggleVisibility(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var block models.CommonBlock
	if err := h.db.First(&block, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Общий блок не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить общий блок"},
		)
	}

	block.IsShow = !block.IsShow

	if err := h.db.Save(&block).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить видимость общего блока"},
		)
	}

	h.db.Preload("Category").
		Preload("Manufacturer").
		Preload("Attribute").
		First(&block, block.ID)

	resp := response.CommonBlockResponse{
		ID:             block.ID,
		Title:          block.Title,
		CoverURL:       block.CoverURL,
		IsShow:         block.IsShow,
		CategoryID:     block.CategoryID,
		ManufacturerID: block.ManufacturerID,
		AttributeID:    block.AttributeID,
		CreatedAt:      block.CreatedAt,
		UpdatedAt:      block.UpdatedAt,
	}

	if block.Category != nil {
		resp.Category = &response.CategoryShortResponse{
			ID:    block.Category.ID,
			Slug:  block.Category.Slug,
			Title: block.Category.Title,
		}
	}

	if block.Manufacturer != nil {
		resp.Manufacturer = &response.ManufacturerShortResponse{
			ID:    block.Manufacturer.ID,
			Slug:  block.Manufacturer.Slug,
			Title: block.Manufacturer.Title,
		}
	}

	if block.Attribute != nil {
		resp.Attribute = &response.AttributeShortResponse{
			ID:            block.Attribute.ID,
			Slug:          block.Attribute.Slug,
			Title:         block.Attribute.Title,
			AttributeType: block.Attribute.AttributeType,
		}
	}

	return c.JSON(resp)
}

// DeleteCommonBlock godoc
// @Summary Удалить общий блок
// @Description Удаляет общий блок из БД по ID
// @Tags CommonBlock
// @Param id path int true "CommonBlock ID"
// @Success 204 "Общий блок успешно удален"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /common-blocks/{id} [delete]
func (h *CommonBlockHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var block models.CommonBlock
	if err := h.db.First(&block, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Общий блок не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить общий блок"},
		)
	}

	if err := h.db.Delete(&block).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить общий блок"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
