package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	order "app/internal/models/order"
	models "app/internal/models/user"
	"app/internal/response"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// GetAllUsers godoc
// @Summary Получить список пользователей
// @Description Возвращает список всех пользователей
// @Tags User
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Лимит на страницу" default(20)
// @Param search query string false "Поиск по email или username"
// @Param right_id query int false "Фильтр по правам"
// @Success 200 {object} response.UserListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users [get]
func (h *UserHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search")
	rightID, _ := strconv.ParseUint(c.Query("right_id"), 10, 32)
	offset := (page - 1) * limit

	query := h.db.Model(&models.User{}).Preload("Right")

	if search != "" {
		query = query.Where("email ILIKE ? OR username ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if rightID > 0 {
		query = query.Where("right_id = ?", rightID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить количество пользователей"},
		)
	}

	var users []models.User
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить пользователей"},
		)
	}

	data := make([]response.UserResponse, 0, len(users))
	for _, user := range users {
		resp := response.UserResponse{
			ID:               user.ID,
			Username:         user.Username,
			Email:            user.Email,
			FirstName:        user.FirstName,
			LastName:         user.LastName,
			FatherName:       user.FatherName,
			Phone:            user.Phone,
			Address:          user.Address,
			RegistrationDate: user.RegistrationDate,
			RightID:          user.RightID,
			CreatedAt:        user.CreatedAt,
			UpdatedAt:        user.UpdatedAt,
		}

		if user.Right != nil {
			resp.Right = &response.RightResponse{
				ID:    user.Right.ID,
				Title: user.Right.Title,
			}
		}

		data = append(data, resp)
	}

	return c.JSON(response.UserListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetUserByID godoc
// @Summary Получить пользователя по ID
// @Description Возвращает пользователя по ID
// @Tags User
// @Param id path int true "User ID"
// @Success 200 {object} response.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var user models.User
	if err := h.db.Preload("Right").First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Пользователь не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить пользователя"},
		)
	}

	resp := response.UserResponse{
		ID:               user.ID,
		Username:         user.Username,
		Email:            user.Email,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		FatherName:       user.FatherName,
		Phone:            user.Phone,
		Address:          user.Address,
		RegistrationDate: user.RegistrationDate,
		RightID:          user.RightID,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}

	if user.Right != nil {
		resp.Right = &response.RightResponse{
			ID:    user.Right.ID,
			Title: user.Right.Title,
		}
	}

	return c.JSON(resp)
}

// CreateUser godoc
// @Summary Создать пользователя
// @Description Создает нового пользователя
// @Tags User
// @Accept json
// @Produce json
// @Param data body response.UserCreate true "Данные для создания пользователя"
// @Success 201 {object} response.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users [post]
func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req response.UserCreate

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

	var existing models.User
	if err := h.db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Пользователь с таким email или username уже существует"},
		)
	}

	var right models.Right
	if err := h.db.First(&right, req.RightID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Права не найдены"},
		)
	}

	user := models.User{
		Username:   req.Username,
		Email:      req.Email,
		Password:   req.Password,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		FatherName: req.FatherName,
		Phone:      req.Phone,
		Address:    req.Address,
		RightID:    req.RightID,
	}

	if err := h.db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось создать пользователя"},
		)
	}

	h.db.Preload("Right").First(&user, user.ID)

	resp := response.UserResponse{
		ID:               user.ID,
		Username:         user.Username,
		Email:            user.Email,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		FatherName:       user.FatherName,
		Phone:            user.Phone,
		Address:          user.Address,
		RegistrationDate: user.RegistrationDate,
		RightID:          user.RightID,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}

	if user.Right != nil {
		resp.Right = &response.RightResponse{
			ID:    user.Right.ID,
			Title: user.Right.Title,
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateUser godoc
// @Summary Полностью обновить пользователя
// @Description Обновляет все поля пользователя по ID (PUT)
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param data body response.UserUpdate true "Обновляемые поля пользователя"
// @Success 200 {object} response.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/{id} [put]
func (h *UserHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Пользователь не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить пользователя"},
		)
	}

	var req response.UserUpdate
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

	if user.Email != req.Email || user.Username != req.Username {
		var existing models.User
		if err := h.db.Where("(email = ? OR username = ?) AND id != ?", req.Email, req.Username, id).
			First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(
				response.ErrorResponse{Error: "Пользователь с таким email или username уже существует"},
			)
		}
	}

	var right models.Right
	if err := h.db.First(&right, req.RightID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.ErrorResponse{Error: "Права не найдены"},
		)
	}

	user.Username = req.Username
	user.Email = req.Email
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.FatherName = req.FatherName
	user.Phone = req.Phone
	user.Address = req.Address
	user.RightID = req.RightID

	if err := h.db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить пользователя"},
		)
	}

	h.db.Preload("Right").First(&user, user.ID)

	resp := response.UserResponse{
		ID:               user.ID,
		Username:         user.Username,
		Email:            user.Email,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		FatherName:       user.FatherName,
		Phone:            user.Phone,
		Address:          user.Address,
		RegistrationDate: user.RegistrationDate,
		RightID:          user.RightID,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}

	if user.Right != nil {
		resp.Right = &response.RightResponse{
			ID:    user.Right.ID,
			Title: user.Right.Title,
		}
	}

	return c.JSON(resp)
}

// PatchUser godoc
// @Summary Частично обновить пользователя
// @Description Обновляет указанные поля пользователя по ID (PATCH)
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param data body response.UserPatch true "Обновляемые поля пользователя"
// @Success 200 {object} response.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/{id} [patch]
func (h *UserHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Пользователь не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить пользователя"},
		)
	}

	var req response.UserPatch
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

	if req.Email != nil || req.Username != nil {
		newEmail := req.Email
		if newEmail == nil {
			newEmail = &user.Email
		}
		newUsername := req.Username
		if newUsername == nil {
			newUsername = &user.Username
		}

		if *newEmail != user.Email || *newUsername != user.Username {
			var existing models.User
			if err := h.db.Where("(email = ? OR username = ?) AND id != ?", *newEmail, *newUsername, id).
				First(&existing).Error; err == nil {
				return c.Status(fiber.StatusConflict).JSON(
					response.ErrorResponse{Error: "Пользователь с таким email или username уже существует"},
				)
			}
		}
	}

	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.FirstName != nil {
		user.FirstName = req.FirstName
	}
	if req.LastName != nil {
		user.LastName = req.LastName
	}
	if req.FatherName != nil {
		user.FatherName = req.FatherName
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.Address != nil {
		user.Address = req.Address
	}
	if req.RightID != nil {
		var right models.Right
		if err := h.db.First(&right, *req.RightID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Права не найдены"},
			)
		}
		user.RightID = *req.RightID
	}

	if err := h.db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить пользователя"},
		)
	}

	h.db.Preload("Right").First(&user, user.ID)

	resp := response.UserResponse{
		ID:               user.ID,
		Username:         user.Username,
		Email:            user.Email,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		FatherName:       user.FatherName,
		Phone:            user.Phone,
		Address:          user.Address,
		RegistrationDate: user.RegistrationDate,
		RightID:          user.RightID,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}

	if user.Right != nil {
		resp.Right = &response.RightResponse{
			ID:    user.Right.ID,
			Title: user.Right.Title,
		}
	}

	return c.JSON(resp)
}

// UpdateUserPassword godoc
// @Summary Обновить пароль пользователя
// @Description Обновляет пароль пользователя
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param data body response.UserPasswordUpdate true "Данные для обновления пароля"
// @Success 200 {object} response.ErrorResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/{id}/password [put]
func (h *UserHandler) UpdatePassword(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Пользователь не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить пользователя"},
		)
	}

	var req response.UserPasswordUpdate
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

	if !user.CheckPassword(req.OldPassword) {
		return c.Status(fiber.StatusUnauthorized).JSON(
			response.ErrorResponse{Error: "Неверный старый пароль"},
		)
	}

	user.Password = req.NewPassword
	if err := h.db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось обновить пароль"},
		)
	}

	return c.JSON(response.ErrorResponse{Error: "Пароль успешно обновлен"})
}

// DeleteUser godoc
// @Summary Удалить пользователя
// @Description Удаляет пользователя из БД по ID
// @Tags User
// @Param id path int true "User ID"
// @Success 204 "Пользователь успешно удален"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{Error: "Неверный формат ID"},
		)
	}

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				response.ErrorResponse{Error: "Пользователь не найден"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось получить пользователя"},
		)
	}

	var ordersCount int64
	if err := h.db.Model(&order.Order{}).Where("user_id = ?", id).Count(&ordersCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось проверить связанные заказы"},
		)
	}

	if ordersCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(
			response.ErrorResponse{Error: "Нельзя удалить пользователя, так как с ним связаны заказы"},
		)
	}

	if err := h.db.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.ErrorResponse{Error: "Не удалось удалить пользователя"},
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
