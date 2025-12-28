// internal/handlers/request_status.go
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "app/internal/models/order"
	"app/internal/response"
)

type RequestStatusHandler struct {
	db *gorm.DB
}

func NewRequestStatusHandler(db *gorm.DB) *RequestStatusHandler {
	return &RequestStatusHandler{db: db}
}

// GetAllRequestStatuses godoc
// @Summary Получить список статусов заявок
// @Description Возвращает список всех статусов заявок
// @Tags RequestStatus
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по названию"
// @Success 200 {object} response.RequestStatusListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /request-statuses [get]
func (h *RequestStatusHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	offset := (page - 1) * limit

	query := h.db.Model(&models.RequestStatus{})

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество статусов заявок"},
		)
	}

	var statuses []models.RequestStatus
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&statuses).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статусы заявок"},
		)
	}

	data := make([]response.RequestStatusResponse, 0, len(statuses))
	for _, status := range statuses {
		data = append(data, response.RequestStatusResponse{
			ID:              status.ID,
			Title:           status.Title,
			BackgroundColor: status.BackgroundColor,
			CreatedAt:       status.CreatedAt,
			UpdatedAt:       status.UpdatedAt,
		})
	}

	return c.JSON(response.RequestStatusListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetRequestStatusByID godoc
// @Summary Получить статус заявки по ID
// @Description Возвращает статус заявки по ID
// @Tags RequestStatus
// @Param id path int true "RequestStatus ID"
// @Success 200 {object} response.RequestStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /request-statuses/{id} [get]
func (h *RequestStatusHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.RequestStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус заявки не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус заявки"},
		)
	}

	return c.JSON(response.RequestStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// CreateRequestStatus godoc
// @Summary Создать статус заявки
// @Description Создает новый статус заявки
// @Tags RequestStatus
// @Accept json
// @Produce json
// @Param data body response.RequestStatusCreate true "Данные для создания статуса заявки"
// @Success 201 {object} response.RequestStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /request-statuses [post]
func (h *RequestStatusHandler) Create(c *fiber.Ctx) error {
	var req response.RequestStatusCreate

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

	var existing models.RequestStatus
	if err := h.db.Where("title = ?", req.Title).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Статус заявки с таким названием уже существует"},
		)
	}

	status := models.RequestStatus{
		Title:           req.Title,
		BackgroundColor: req.BackgroundColor,
	}

	if err := h.db.Create(&status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать статус заявки"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(response.RequestStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// UpdateRequestStatus godoc
// @Summary Полностью обновить статус заявки
// @Description Обновляет все поля статуса заявки по ID (PUT)
// @Tags RequestStatus
// @Accept json
// @Produce json
// @Param id path int true "RequestStatus ID"
// @Param data body response.RequestStatusUpdate true "Обновляемые поля статуса заявки"
// @Success 200 {object} response.RequestStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /request-statuses/{id} [put]
func (h *RequestStatusHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.RequestStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус заявки не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус заявки"},
		)
	}

	var req response.RequestStatusUpdate
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

	if status.Title != req.Title {
		var existing models.RequestStatus
		if err := h.db.Where("title = ? AND id != ?", req.Title, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Статус заявки с таким названием уже существует"},
			)
		}
	}

	status.Title = req.Title
	status.BackgroundColor = req.BackgroundColor

	if err := h.db.Save(&status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить статус заявки"},
		)
	}

	return c.JSON(response.RequestStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// PatchRequestStatus godoc
// @Summary Частично обновить статус заявки
// @Description Обновляет указанные поля статуса заявки по ID (PATCH)
// @Tags RequestStatus
// @Accept json
// @Produce json
// @Param id path int true "RequestStatus ID"
// @Param data body response.RequestStatusPatch true "Обновляемые поля статуса заявки"
// @Success 200 {object} response.RequestStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /request-statuses/{id} [patch]
func (h *RequestStatusHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.RequestStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус заявки не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус заявки"},
		)
	}

	var req response.RequestStatusPatch
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
		if status.Title != *req.Title {
			var existing models.RequestStatus
			if err := h.db.Where("title = ? AND id != ?", *req.Title, id).First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Статус заявки с таким названием уже существует"},
				)
			}
		}
		status.Title = *req.Title
	}

	if req.BackgroundColor != nil {
		status.BackgroundColor = *req.BackgroundColor
	}

	if err := h.db.Save(&status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить статус заявки"},
		)
	}

	return c.JSON(response.RequestStatusResponse{
		ID:              status.ID,
		Title:           status.Title,
		BackgroundColor: status.BackgroundColor,
		CreatedAt:       status.CreatedAt,
		UpdatedAt:       status.UpdatedAt,
	})
}

// DeleteRequestStatus godoc
// @Summary Удалить статус заявки
// @Description Удаляет статус заявки из БД по ID
// @Tags RequestStatus
// @Param id path int true "RequestStatus ID"
// @Success 204 "Статус заявки успешно удалён"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /request-statuses/{id} [delete]
func (h *RequestStatusHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var status models.RequestStatus
	if err := h.db.First(&status, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Статус заявки не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить статус заявки"},
		)
	}

	var requestsCount int64
	if err := h.db.Model(&models.Request{}).Where("request_status_id = ?", id).Count(&requestsCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные заявки"},
		)
	}

	if requestsCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить статус заявки, так как с ним связаны заявки"},
		)
	}

	if err := h.db.Delete(&status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить статус заявки"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
