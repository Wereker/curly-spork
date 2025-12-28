// internal/handlers/category.go
package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/catalog"
	"app/internal/response"
)

type CategoryHandler struct {
	db *gorm.DB
}

func NewCategoryHandler(db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

// GetAllCategories godoc
// @Summary Получить список категорий
// @Description Возвращает список всех категорий с пагинацией
// @Tags Category
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по названию или slug"
// @Param parent_id query int false "Фильтр по родительской категории"
// @Param only_root query bool false "Только корневые категории" default(false)
// @Success 200 {object} response.CategoryListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /categories [get]
func (h *CategoryHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	parentID, _ := strconv.ParseUint(c.Query("parent_id"), 10, 32)
	onlyRoot, _ := strconv.ParseBool(c.Query("only_root", "false"))
	offset := (page - 1) * limit

	query := h.db.Model(&models.Category{}).
		Preload("Parent").
		Preload("Children").
		Select("categories.*, (SELECT COUNT(*) FROM products WHERE products.category_id = categories.id) as products_count")

	if search != "" {
		query = query.Where("title ILIKE ? OR slug ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if onlyRoot {
		query = query.Where("parent_id IS NULL")
	} else if parentID > 0 {
		query = query.Where("parent_id = ?", parentID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество категорий"},
		)
	}

	var categories []models.Category
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&categories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить категории"},
		)
	}

	data := make([]response.CategoryResponse, 0, len(categories))
	for _, category := range categories {
		resp := response.CategoryResponse{
			ID:          category.ID,
			Slug:        category.Slug,
			Title:       category.Title,
			Description: category.Description,
			ParentID:    category.ParentID,
			CoverURL:    category.CoverURL,
			CreatedAt:   category.CreatedAt,
			UpdatedAt:   category.UpdatedAt,
		}

		if category.Parent != nil {
			resp.Parent = &response.CategoryShortResponse{
				ID:       category.Parent.ID,
				Slug:     category.Parent.Slug,
				Title:    category.Parent.Title,
				CoverURL: category.Parent.CoverURL,
			}
		}

		if len(category.Children) > 0 {
			resp.Children = make([]response.CategoryShortResponse, 0, len(category.Children))
			for _, child := range category.Children {
				resp.Children = append(resp.Children, response.CategoryShortResponse{
					ID:       child.ID,
					Slug:     child.Slug,
					Title:    child.Title,
					CoverURL: child.CoverURL,
				})
			}
		}

		data = append(data, resp)
	}

	return c.JSON(response.CategoryListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetCategoryTree godoc
// @Summary Получить дерево категорий
// @Description Возвращает дерево всех категорий
// @Tags Category
// @Success 200 {array} response.CategoryTreeResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /categories/tree [get]
func (h *CategoryHandler) GetTree(c *fiber.Ctx) error {
	var categories []models.Category
	if err := h.db.Where("parent_id IS NULL").
		Preload("Children").
		Preload("Children.Children").
		Order("title ASC").
		Find(&categories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить дерево категорий"},
		)
	}

	var tree []response.CategoryTreeResponse
	for _, category := range categories {
		tree = append(tree, h.buildCategoryTree(category))
	}

	return c.JSON(tree)
}

func (h *CategoryHandler) buildCategoryTree(category models.Category) response.CategoryTreeResponse {
	resp := response.CategoryTreeResponse{
		ID:          category.ID,
		Slug:        category.Slug,
		Title:       category.Title,
		Description: category.Description,
		CoverURL:    category.CoverURL,
	}

	if len(category.Children) > 0 {
		resp.Children = make([]response.CategoryTreeResponse, 0, len(category.Children))
		for _, child := range category.Children {
			resp.Children = append(resp.Children, h.buildCategoryTree(child))
		}
	}

	return resp
}

// GetCategoryByID godoc
// @Summary Получить категорию по ID
// @Description Возвращает категорию по ID
// @Tags Category
// @Param id path int true "Category ID"
// @Success 200 {object} response.CategoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /categories/{id} [get]
func (h *CategoryHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var category models.Category
	if err := h.db.Preload("Parent").
		Preload("Children").
		Preload("Products").
		First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Категория не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить категорию"},
		)
	}

	var productsCount int64
	h.db.Model(&models.Product{}).Where("category_id = ?", id).Count(&productsCount)

	resp := response.CategoryResponse{
		ID:            category.ID,
		Slug:          category.Slug,
		Title:         category.Title,
		Description:   category.Description,
		ParentID:      category.ParentID,
		CoverURL:      category.CoverURL,
		ProductsCount: productsCount,
		CreatedAt:     category.CreatedAt,
		UpdatedAt:     category.UpdatedAt,
	}

	if category.Parent != nil {
		resp.Parent = &response.CategoryShortResponse{
			ID:       category.Parent.ID,
			Slug:     category.Parent.Slug,
			Title:    category.Parent.Title,
			CoverURL: category.Parent.CoverURL,
		}
	}

	if len(category.Children) > 0 {
		resp.Children = make([]response.CategoryShortResponse, 0, len(category.Children))
		for _, child := range category.Children {
			resp.Children = append(resp.Children, response.CategoryShortResponse{
				ID:       child.ID,
				Slug:     child.Slug,
				Title:    child.Title,
				CoverURL: child.CoverURL,
			})
		}
	}

	return c.JSON(resp)
}

// GetCategoryBySlug godoc
// @Summary Получить категорию по slug
// @Description Возвращает категорию по slug
// @Tags Category
// @Param slug path string true "Category Slug"
// @Success 200 {object} response.CategoryResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /categories/slug/{slug} [get]
func (h *CategoryHandler) GetBySlug(c *fiber.Ctx) error {
	slug := strings.TrimSpace(c.Params("slug"))
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Slug не может быть пустым"},
		)
	}

	var category models.Category
	if err := h.db.Preload("Parent").
		Preload("Children").
		Where("slug = ?", slug).
		First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Категория не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить категорию"},
		)
	}

	var productsCount int64
	h.db.Model(&models.Product{}).Where("category_id = ?", category.ID).Count(&productsCount)

	resp := response.CategoryResponse{
		ID:            category.ID,
		Slug:          category.Slug,
		Title:         category.Title,
		Description:   category.Description,
		ParentID:      category.ParentID,
		CoverURL:      category.CoverURL,
		ProductsCount: productsCount,
		CreatedAt:     category.CreatedAt,
		UpdatedAt:     category.UpdatedAt,
	}

	if category.Parent != nil {
		resp.Parent = &response.CategoryShortResponse{
			ID:       category.Parent.ID,
			Slug:     category.Parent.Slug,
			Title:    category.Parent.Title,
			CoverURL: category.Parent.CoverURL,
		}
	}

	if len(category.Children) > 0 {
		resp.Children = make([]response.CategoryShortResponse, 0, len(category.Children))
		for _, child := range category.Children {
			resp.Children = append(resp.Children, response.CategoryShortResponse{
				ID:       child.ID,
				Slug:     child.Slug,
				Title:    child.Title,
				CoverURL: child.CoverURL,
			})
		}
	}

	return c.JSON(resp)
}

// CreateCategory godoc
// @Summary Создать категорию
// @Description Создает новую категорию
// @Tags Category
// @Accept json
// @Produce json
// @Param data body response.CategoryCreate true "Данные для создания категории"
// @Success 201 {object} response.CategoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /categories [post]
func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	var req response.CategoryCreate

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

	var existing models.Category
	if err := h.db.Where("slug = ?", req.Slug).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Категория с таким slug уже существует"},
		)
	}

	if req.ParentID != nil && *req.ParentID > 0 {
		var parent models.Category
		if err := h.db.First(&parent, *req.ParentID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Родительская категория не найдена"},
			)
		}
	}

	category := models.Category{
		Slug:        req.Slug,
		Title:       req.Title,
		Description: req.Description,
		ParentID:    req.ParentID,
		CoverURL:    req.CoverURL,
	}

	if err := h.db.Create(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать категорию"},
		)
	}

	h.db.Preload("Parent").Preload("Children").First(&category, category.ID)

	resp := response.CategoryResponse{
		ID:          category.ID,
		Slug:        category.Slug,
		Title:       category.Title,
		Description: category.Description,
		ParentID:    category.ParentID,
		CoverURL:    category.CoverURL,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}

	if category.Parent != nil {
		resp.Parent = &response.CategoryShortResponse{
			ID:       category.Parent.ID,
			Slug:     category.Parent.Slug,
			Title:    category.Parent.Title,
			CoverURL: category.Parent.CoverURL,
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateCategory godoc
// @Summary Полностью обновить категорию
// @Description Обновляет все поля категории по ID (PUT)
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param data body response.CategoryUpdate true "Обновляемые поля категории"
// @Success 200 {object} response.CategoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /categories/{id} [put]
func (h *CategoryHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var category models.Category
	if err := h.db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Категория не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить категорию"},
		)
	}

	var req response.CategoryUpdate
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

	if category.Slug != req.Slug {
		var existing models.Category
		if err := h.db.Where("slug = ? AND id != ?", req.Slug, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Категория с таким slug уже существует"},
			)
		}
	}

	if req.ParentID != nil && *req.ParentID > 0 {
		if *req.ParentID == uint(id) {
			return c.Status(fiber.StatusBadRequest).JSON(
				response.ErrorResponse{Error: "Категория не может быть родителем самой себя"},
			)
		}

		var parent models.Category
		if err := h.db.First(&parent, *req.ParentID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Родительская категория не найдена"},
			)
		}

		if err := h.checkCircularReference(uint(id), *req.ParentID); err != nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: err.Error()},
			)
		}
	}

	category.Slug = req.Slug
	category.Title = req.Title
	category.Description = req.Description
	category.ParentID = req.ParentID
	category.CoverURL = req.CoverURL

	if err := h.db.Save(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить категорию"},
		)
	}

	h.db.Preload("Parent").Preload("Children").First(&category, category.ID)

	resp := response.CategoryResponse{
		ID:          category.ID,
		Slug:        category.Slug,
		Title:       category.Title,
		Description: category.Description,
		ParentID:    category.ParentID,
		CoverURL:    category.CoverURL,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}

	if category.Parent != nil {
		resp.Parent = &response.CategoryShortResponse{
			ID:       category.Parent.ID,
			Slug:     category.Parent.Slug,
			Title:    category.Parent.Title,
			CoverURL: category.Parent.CoverURL,
		}
	}

	return c.JSON(resp)
}

// PatchCategory godoc
// @Summary Частично обновить категорию
// @Description Обновляет указанные поля категории по ID (PATCH)
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param data body response.CategoryPatch true "Обновляемые поля категории"
// @Success 200 {object} response.CategoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /categories/{id} [patch]
func (h *CategoryHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var category models.Category
	if err := h.db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Категория не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить категорию"},
		)
	}

	var req response.CategoryPatch
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
		if category.Slug != *req.Slug {
			var existing models.Category
			if err := h.db.Where("slug = ? AND id != ?", *req.Slug, id).First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Категория с таким slug уже существует"},
				)
			}
		}
		category.Slug = *req.Slug
	}

	if req.Title != nil {
		category.Title = *req.Title
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.CoverURL != nil {
		category.CoverURL = *req.CoverURL
	}

	if req.ParentID != nil {
		if *req.ParentID == 0 {
			category.ParentID = nil
		} else {
			if *req.ParentID == uint(id) {
				return c.Status(fiber.StatusBadRequest).JSON(
					response.ErrorResponse{Error: "Категория не может быть родителем самой себя"},
				)
			}

			var parent models.Category
			if err := h.db.First(&parent, *req.ParentID).Error; err != nil {
				return c.Status(fiber.StatusNotFound).JSON(
					response.ErrorResponse{Error: "Родительская категория не найдена"},
				)
			}

			if err := h.checkCircularReference(uint(id), *req.ParentID); err != nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: err.Error()},
				)
			}
			category.ParentID = req.ParentID
		}
	}

	if err := h.db.Save(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить категорию"},
		)
	}

	h.db.Preload("Parent").Preload("Children").First(&category, category.ID)

	resp := response.CategoryResponse{
		ID:          category.ID,
		Slug:        category.Slug,
		Title:       category.Title,
		Description: category.Description,
		ParentID:    category.ParentID,
		CoverURL:    category.CoverURL,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}

	if category.Parent != nil {
		resp.Parent = &response.CategoryShortResponse{
			ID:       category.Parent.ID,
			Slug:     category.Parent.Slug,
			Title:    category.Parent.Title,
			CoverURL: category.Parent.CoverURL,
		}
	}

	return c.JSON(resp)
}

// DeleteCategory godoc
// @Summary Удалить категорию
// @Description Удаляет категорию из БД по ID
// @Tags Category
// @Param id path int true "Category ID"
// @Success 204 "Категория успешно удалена"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /categories/{id} [delete]
func (h *CategoryHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var category models.Category
	if err := h.db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Категория не найдена"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить категорию"},
		)
	}

	var childrenCount int64
	if err := h.db.Model(&models.Category{}).Where("parent_id = ?", id).Count(&childrenCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить дочерние категории"},
		)
	}

	if childrenCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить категорию, так как у нее есть дочерние категории"},
		)
	}

	var productsCount int64
	if err := h.db.Model(&models.Product{}).Where("category_id = ?", id).Count(&productsCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные товары"},
		)
	}

	if productsCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить категорию, так как с ней связаны товары"},
		)
	}

	if err := h.db.Delete(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить категорию"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CategoryHandler) checkCircularReference(categoryID, parentID uint) error {
	currentID := parentID
	for {
		var parent models.Category
		if err := h.db.Select("id, parent_id").First(&parent, currentID).Error; err != nil {
			return nil
		}

		if parent.ParentID == nil {
			return nil
		}

		if *parent.ParentID == categoryID {
			return fiber.NewError(fiber.StatusConflict, "Обнаружена циклическая ссылка в иерархии категорий")
		}

		currentID = *parent.ParentID
	}
}
