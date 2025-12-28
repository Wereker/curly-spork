// internal/handlers/right.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/user"
	"app/internal/response"
)

type RightHandler struct {
	db *gorm.DB
}

func NewRightHandler(db *gorm.DB) *RightHandler {
	return &RightHandler{db: db}
}

// GetAllRights godoc
// @Summary Получить список прав
// @Description Возвращает список всех прав
// @Tags Right
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по названию"
// @Success 200 {object} response.RightListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /rights [get]
func (h *RightHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	offset := (page - 1) * limit

	query := h.db.Model(&models.Right{})

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество прав"},
		)
	}

	var rights []models.Right
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rights).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить права"},
		)
	}

	data := make([]response.RightResponse, 0, len(rights))
	for _, right := range rights {
		data = append(data, response.RightResponse{
			ID:               right.ID,
			Title:            right.Title,
			Description:      right.Description,
			CanViewCatalog:   right.CanViewCatalog,
			CanMakeOrder:     right.CanMakeOrder,
			CanManageOrder:   right.CanManageOrder,
			CanManageProduct: right.CanManageProduct,
			CanManageUser:    right.CanManageUser,
			CreatedAt:        right.CreatedAt,
			UpdatedAt:        right.UpdatedAt,
		})
	}

	return c.JSON(response.RightListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetRightByID godoc
// @Summary Получить право по ID
// @Description Возвращает право по ID
// @Tags Right
// @Param id path int true "Right ID"
// @Success 200 {object} response.RightResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /rights/{id} [get]
func (h *RightHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var right models.Right
	if err := h.db.First(&right, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Право не найдено"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить право"},
		)
	}

	return c.JSON(response.RightResponse{
		ID:               right.ID,
		Title:            right.Title,
		Description:      right.Description,
		CanViewCatalog:   right.CanViewCatalog,
		CanMakeOrder:     right.CanMakeOrder,
		CanManageOrder:   right.CanManageOrder,
		CanManageProduct: right.CanManageProduct,
		CanManageUser:    right.CanManageUser,
		CreatedAt:        right.CreatedAt,
		UpdatedAt:        right.UpdatedAt,
	})
}

// CreateRight godoc
// @Summary Создать право
// @Description Создает новое право
// @Tags Right
// @Accept json
// @Produce json
// @Param data body response.RightCreate true "Данные для создания права"
// @Success 201 {object} response.RightResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /rights [post]
func (h *RightHandler) Create(c *fiber.Ctx) error {
	var req response.RightCreate

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

	var existing models.Right
	if err := h.db.Where("title = ?", req.Title).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Право с таким названием уже существует"},
		)
	}

	right := models.Right{
		Title:            req.Title,
		Description:      req.Description,
		CanViewCatalog:   req.CanViewCatalog,
		CanMakeOrder:     req.CanMakeOrder,
		CanManageOrder:   req.CanManageOrder,
		CanManageProduct: req.CanManageProduct,
		CanManageUser:    req.CanManageUser,
	}

	if err := h.db.Create(&right).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать право"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(response.RightResponse{
		ID:               right.ID,
		Title:            right.Title,
		Description:      right.Description,
		CanViewCatalog:   right.CanViewCatalog,
		CanMakeOrder:     right.CanMakeOrder,
		CanManageOrder:   right.CanManageOrder,
		CanManageProduct: right.CanManageProduct,
		CanManageUser:    right.CanManageUser,
		CreatedAt:        right.CreatedAt,
		UpdatedAt:        right.UpdatedAt,
	})
}

// UpdateRight godoc
// @Summary Полностью обновить право
// @Description Обновляет все поля права по ID (PUT)
// @Tags Right
// @Accept json
// @Produce json
// @Param id path int true "Right ID"
// @Param data body response.RightUpdate true "Обновляемые поля права"
// @Success 200 {object} response.RightResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /rights/{id} [put]
func (h *RightHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var right models.Right
	if err := h.db.First(&right, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Право не найдено"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить право"},
		)
	}

	var req response.RightUpdate
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

	if right.Title != req.Title {
		var existing models.Right
		if err := h.db.Where("title = ? AND id != ?", req.Title, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Право с таким названием уже существует"},
			)
		}
	}

	right.Title = req.Title
	right.Description = req.Description
	right.CanViewCatalog = req.CanViewCatalog
	right.CanMakeOrder = req.CanMakeOrder
	right.CanManageOrder = req.CanManageOrder
	right.CanManageProduct = req.CanManageProduct
	right.CanManageUser = req.CanManageUser

	if err := h.db.Save(&right).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить право"},
		)
	}

	return c.JSON(response.RightResponse{
		ID:               right.ID,
		Title:            right.Title,
		Description:      right.Description,
		CanViewCatalog:   right.CanViewCatalog,
		CanMakeOrder:     right.CanMakeOrder,
		CanManageOrder:   right.CanManageOrder,
		CanManageProduct: right.CanManageProduct,
		CanManageUser:    right.CanManageUser,
		CreatedAt:        right.CreatedAt,
		UpdatedAt:        right.UpdatedAt,
	})
}

// PatchRight godoc
// @Summary Частично обновить право
// @Description Обновляет указанные поля права по ID (PATCH)
// @Tags Right
// @Accept json
// @Produce json
// @Param id path int true "Right ID"
// @Param data body response.RightPatch true "Обновляемые поля права"
// @Success 200 {object} response.RightResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /rights/{id} [patch]
func (h *RightHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var right models.Right
	if err := h.db.First(&right, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Право не найдено"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить право"},
		)
	}

	var req response.RightPatch
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
		if right.Title != *req.Title {
			var existing models.Right
			if err := h.db.Where("title = ? AND id != ?", *req.Title, id).First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Право с таким названием уже существует"},
				)
			}
		}
		right.Title = *req.Title
	}
	if req.Description != nil {
		right.Description = req.Description
	}
	if req.CanViewCatalog != nil {
		right.CanViewCatalog = *req.CanViewCatalog
	}
	if req.CanMakeOrder != nil {
		right.CanMakeOrder = *req.CanMakeOrder
	}
	if req.CanManageOrder != nil {
		right.CanManageOrder = *req.CanManageOrder
	}
	if req.CanManageProduct != nil {
		right.CanManageProduct = *req.CanManageProduct
	}
	if req.CanManageUser != nil {
		right.CanManageUser = *req.CanManageUser
	}

	if err := h.db.Save(&right).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить право"},
		)
	}

	return c.JSON(response.RightResponse{
		ID:               right.ID,
		Title:            right.Title,
		Description:      right.Description,
		CanViewCatalog:   right.CanViewCatalog,
		CanMakeOrder:     right.CanMakeOrder,
		CanManageOrder:   right.CanManageOrder,
		CanManageProduct: right.CanManageProduct,
		CanManageUser:    right.CanManageUser,
		CreatedAt:        right.CreatedAt,
		UpdatedAt:        right.UpdatedAt,
	})
}

// DeleteRight godoc
// @Summary Удалить право
// @Description Удаляет право из БД по ID
// @Tags Right
// @Param id path int true "Right ID"
// @Success 204 "Право успешно удалено"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /rights/{id} [delete]
func (h *RightHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var right models.Right
	if err := h.db.First(&right, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Право не найдено"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить право"},
		)
	}

	var usersCount int64
	if err := h.db.Model(&models.User{}).Where("right_id = ?", id).Count(&usersCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанных пользователей"},
		)
	}

	if usersCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить право, так как с ним связаны пользователи"},
		)
	}

	if err := h.db.Delete(&right).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить право"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
