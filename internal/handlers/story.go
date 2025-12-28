// internal/handlers/story.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/sitecontent"
	"app/internal/response"
)

type StoryHandler struct {
	db *gorm.DB
}

func NewStoryHandler(db *gorm.DB) *StoryHandler {
	return &StoryHandler{db: db}
}

// GetAllStories godoc
// @Summary Получить список историй
// @Description Возвращает список всех историй с пагинацией
// @Tags Story
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по названию"
// @Param is_show query bool false "Фильтр по отображению" default(true)
// @Success 200 {object} response.StoryListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /stories [get]
func (h *StoryHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	isShow, _ := strconv.ParseBool(c.Query("is_show", "true"))
	offset := (page - 1) * limit

	query := h.db.Model(&models.Story{})

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}
	if c.Query("is_show") != "" {
		query = query.Where("is_show = ?", isShow)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество историй"},
		)
	}

	var stories []models.Story
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&stories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить истории"},
		)
	}

	data := make([]response.StoryResponse, 0, len(stories))
	for _, story := range stories {
		data = append(data, response.StoryResponse{
			ID:        story.ID,
			Title:     story.Title,
			CoverURL:  story.CoverURL,
			IsShow:    story.IsShow,
			CreatedAt: story.CreatedAt,
			UpdatedAt: story.UpdatedAt,
		})
	}

	return c.JSON(response.StoryListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetStoryByID godoc
// @Summary Получить историю по ID
// @Description Возвращает историю по ID
// @Tags Story
// @Param id path int true "Story ID"
// @Success 200 {object} response.StoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /stories/{id} [get]
func (h *StoryHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var story models.Story
	if err := h.db.First(&story, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "История не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить историю"},
		)
	}

	return c.JSON(response.StoryResponse{
		ID:        story.ID,
		Title:     story.Title,
		CoverURL:  story.CoverURL,
		IsShow:    story.IsShow,
		CreatedAt: story.CreatedAt,
		UpdatedAt: story.UpdatedAt,
	})
}

// CreateStory godoc
// @Summary Создать историю
// @Description Создает новую историю
// @Tags Story
// @Accept json
// @Produce json
// @Param data body response.StoryCreate true "Данные для создания истории"
// @Success 201 {object} response.StoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /stories [post]
func (h *StoryHandler) Create(c *fiber.Ctx) error {
	var req response.StoryCreate

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

	story := models.Story{
		Title:    req.Title,
		CoverURL: req.CoverURL,
		IsShow:   req.IsShow,
	}

	if err := h.db.Create(&story).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать историю"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(response.StoryResponse{
		ID:        story.ID,
		Title:     story.Title,
		CoverURL:  story.CoverURL,
		IsShow:    story.IsShow,
		CreatedAt: story.CreatedAt,
		UpdatedAt: story.UpdatedAt,
	})
}

// UpdateStory godoc
// @Summary Полностью обновить историю
// @Description Обновляет все поля истории по ID (PUT)
// @Tags Story
// @Accept json
// @Produce json
// @Param id path int true "Story ID"
// @Param data body response.StoryUpdate true "Обновляемые поля истории"
// @Success 200 {object} response.StoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /stories/{id} [put]
func (h *StoryHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var story models.Story
	if err := h.db.First(&story, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "История не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить историю"},
		)
	}

	var req response.StoryUpdate
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

	story.Title = req.Title
	story.CoverURL = req.CoverURL
	story.IsShow = req.IsShow

	if err := h.db.Save(&story).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить историю"},
		)
	}

	return c.JSON(response.StoryResponse{
		ID:        story.ID,
		Title:     story.Title,
		CoverURL:  story.CoverURL,
		IsShow:    story.IsShow,
		CreatedAt: story.CreatedAt,
		UpdatedAt: story.UpdatedAt,
	})
}

// PatchStory godoc
// @Summary Частично обновить историю
// @Description Обновляет указанные поля истории по ID (PATCH)
// @Tags Story
// @Accept json
// @Produce json
// @Param id path int true "Story ID"
// @Param data body response.StoryPatch true "Обновляемые поля истории"
// @Success 200 {object} response.StoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /stories/{id} [patch]
func (h *StoryHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var story models.Story
	if err := h.db.First(&story, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "История не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить историю"},
		)
	}

	var req response.StoryPatch
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
		story.Title = *req.Title
	}
	if req.CoverURL != nil {
		story.CoverURL = *req.CoverURL
	}
	if req.IsShow != nil {
		story.IsShow = *req.IsShow
	}

	if err := h.db.Save(&story).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить историю"},
		)
	}

	return c.JSON(response.StoryResponse{
		ID:        story.ID,
		Title:     story.Title,
		CoverURL:  story.CoverURL,
		IsShow:    story.IsShow,
		CreatedAt: story.CreatedAt,
		UpdatedAt: story.UpdatedAt,
	})
}

// ToggleStoryVisibility godoc
// @Summary Переключить видимость истории
// @Description Переключает видимость истории (is_show)
// @Tags Story
// @Param id path int true "Story ID"
// @Success 200 {object} response.StoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /stories/{id}/toggle-visibility [patch]
func (h *StoryHandler) ToggleVisibility(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var story models.Story
	if err := h.db.First(&story, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "История не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить историю"},
		)
	}

	story.IsShow = !story.IsShow

	if err := h.db.Save(&story).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить видимость истории"},
		)
	}

	return c.JSON(response.StoryResponse{
		ID:        story.ID,
		Title:     story.Title,
		CoverURL:  story.CoverURL,
		IsShow:    story.IsShow,
		CreatedAt: story.CreatedAt,
		UpdatedAt: story.UpdatedAt,
	})
}

// DeleteStory godoc
// @Summary Удалить историю
// @Description Удаляет историю из БД по ID
// @Tags Story
// @Param id path int true "Story ID"
// @Success 204 "История успешно удалена"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /stories/{id} [delete]
func (h *StoryHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var story models.Story
	if err := h.db.First(&story, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "История не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить историю"},
		)
	}

	if err := h.db.Delete(&story).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить историю"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
