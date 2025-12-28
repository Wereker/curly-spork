// internal/handlers/new.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/sitecontent"
	"app/internal/response"
)

type NewHandler struct {
	db *gorm.DB
}

func NewNewHandler(db *gorm.DB) *NewHandler {
	return &NewHandler{db: db}
}

// GetAllNews godoc
// @Summary Получить список новостей
// @Description Возвращает список всех новостей с пагинацией
// @Tags New
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по заголовку или подзаголовку"
// @Param is_show query bool false "Фильтр по отображению" default(true)
// @Param sort_by query string false "Сортировка (created_at, number, title)" default(created_at)
// @Param sort_order query string false "Порядок сортировки (asc, desc)" default(desc)
// @Success 200 {object} response.NewListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /news [get]
func (h *NewHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	isShow, _ := strconv.ParseBool(c.Query("is_show", "true"))
	sortBy := c.Query("sort_by", "created_at")
	sortOrder := c.Query("sort_order", "desc")
	offset := (page - 1) * limit

	query := h.db.Model(&models.New{})

	if search != "" {
		query = query.Where("title ILIKE ? OR subtitle ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if c.Query("is_show") != "" {
		query = query.Where("is_show = ?", isShow)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество новостей"},
		)
	}

	// Валидация и установка сортировки
	orderClause := "created_at DESC"
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	switch sortBy {
	case "number":
		orderClause = "number " + sortOrder
	case "title":
		orderClause = "title " + sortOrder
	default:
		orderClause = "created_at " + sortOrder
	}

	var news []models.New
	if err := query.Order(orderClause).
		Limit(limit).
		Offset(offset).
		Find(&news).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить новости"},
		)
	}

	data := make([]response.NewResponse, 0, len(news))
	for _, new := range news {
		data = append(data, response.NewResponse{
			ID:        new.ID,
			Title:     new.Title,
			Subtitle:  new.Subtitle,
			CoverURL:  new.CoverURL,
			Number:    new.Number,
			IsShow:    new.IsShow,
			CreatedAt: new.CreatedAt,
			UpdatedAt: new.UpdatedAt,
		})
	}

	return c.JSON(response.NewListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetNewByID godoc
// @Summary Получить новость по ID
// @Description Возвращает новость по ID
// @Tags New
// @Param id path int true "New ID"
// @Success 200 {object} response.NewResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /news/{id} [get]
func (h *NewHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var new models.New
	if err := h.db.First(&new, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Новость не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить новость"},
		)
	}

	return c.JSON(response.NewResponse{
		ID:        new.ID,
		Title:     new.Title,
		Subtitle:  new.Subtitle,
		CoverURL:  new.CoverURL,
		Number:    new.Number,
		IsShow:    new.IsShow,
		CreatedAt: new.CreatedAt,
		UpdatedAt: new.UpdatedAt,
	})
}

// CreateNew godoc
// @Summary Создать новость
// @Description Создает новую новость
// @Tags New
// @Accept json
// @Produce json
// @Param data body response.NewCreate true "Данные для создания новости"
// @Success 201 {object} response.NewResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /news [post]
func (h *NewHandler) Create(c *fiber.Ctx) error {
	var req response.NewCreate

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

	new := models.New{
		Title:    req.Title,
		Subtitle: req.Subtitle,
		CoverURL: req.CoverURL,
		Number:   req.Number,
		IsShow:   req.IsShow,
	}

	if err := h.db.Create(&new).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать новость"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(response.NewResponse{
		ID:        new.ID,
		Title:     new.Title,
		Subtitle:  new.Subtitle,
		CoverURL:  new.CoverURL,
		Number:    new.Number,
		IsShow:    new.IsShow,
		CreatedAt: new.CreatedAt,
		UpdatedAt: new.UpdatedAt,
	})
}

// UpdateNew godoc
// @Summary Полностью обновить новость
// @Description Обновляет все поля новости по ID (PUT)
// @Tags New
// @Accept json
// @Produce json
// @Param id path int true "New ID"
// @Param data body response.NewUpdate true "Обновляемые поля новости"
// @Success 200 {object} response.NewResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /news/{id} [put]
func (h *NewHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var new models.New
	if err := h.db.First(&new, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Новость не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить новость"},
		)
	}

	var req response.NewUpdate
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

	new.Title = req.Title
	new.Subtitle = req.Subtitle
	new.CoverURL = req.CoverURL
	new.Number = req.Number
	new.IsShow = req.IsShow

	if err := h.db.Save(&new).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить новость"},
		)
	}

	return c.JSON(response.NewResponse{
		ID:        new.ID,
		Title:     new.Title,
		Subtitle:  new.Subtitle,
		CoverURL:  new.CoverURL,
		Number:    new.Number,
		IsShow:    new.IsShow,
		CreatedAt: new.CreatedAt,
		UpdatedAt: new.UpdatedAt,
	})
}

// PatchNew godoc
// @Summary Частично обновить новость
// @Description Обновляет указанные поля новости по ID (PATCH)
// @Tags New
// @Accept json
// @Produce json
// @Param id path int true "New ID"
// @Param data body response.NewPatch true "Обновляемые поля новости"
// @Success 200 {object} response.NewResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /news/{id} [patch]
func (h *NewHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var new models.New
	if err := h.db.First(&new, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Новость не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить новость"},
		)
	}

	var req response.NewPatch
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
		new.Title = *req.Title
	}
	if req.Subtitle != nil {
		new.Subtitle = *req.Subtitle
	}
	if req.CoverURL != nil {
		new.CoverURL = *req.CoverURL
	}
	if req.Number != nil {
		new.Number = *req.Number
	}
	if req.IsShow != nil {
		new.IsShow = *req.IsShow
	}

	if err := h.db.Save(&new).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить новость"},
		)
	}

	return c.JSON(response.NewResponse{
		ID:        new.ID,
		Title:     new.Title,
		Subtitle:  new.Subtitle,
		CoverURL:  new.CoverURL,
		Number:    new.Number,
		IsShow:    new.IsShow,
		CreatedAt: new.CreatedAt,
		UpdatedAt: new.UpdatedAt,
	})
}

// ToggleNewVisibility godoc
// @Summary Переключить видимость новости
// @Description Переключает видимость новости (is_show)
// @Tags New
// @Param id path int true "New ID"
// @Success 200 {object} response.NewResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /news/{id}/toggle-visibility [patch]
func (h *NewHandler) ToggleVisibility(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var new models.New
	if err := h.db.First(&new, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Новость не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить новость"},
		)
	}

	new.IsShow = !new.IsShow

	if err := h.db.Save(&new).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить видимость новости"},
		)
	}

	return c.JSON(response.NewResponse{
		ID:        new.ID,
		Title:     new.Title,
		Subtitle:  new.Subtitle,
		CoverURL:  new.CoverURL,
		Number:    new.Number,
		IsShow:    new.IsShow,
		CreatedAt: new.CreatedAt,
		UpdatedAt: new.UpdatedAt,
	})
}

// DeleteNew godoc
// @Summary Удалить новость
// @Description Удаляет новость из БД по ID
// @Tags New
// @Param id path int true "New ID"
// @Success 204 "Новость успешно удалена"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /news/{id} [delete]
func (h *NewHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var new models.New
	if err := h.db.First(&new, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Новость не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить новость"},
		)
	}

	if err := h.db.Delete(&new).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить новость"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
