package response

import (
	"time"
)

// ShipmentStatusResponse - схема для ответа с данными статуса доставки
type ShipmentStatusResponse struct {
	ID              uint      `json:"id"`
	Title           string    `json:"title"`
	BackgroundColor string    `json:"background_color"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ShipmentStatusCreate - схема для создания статуса доставки
type ShipmentStatusCreate struct {
	Title           string `json:"title" validate:"required,min=2,max=50"`
	BackgroundColor string `json:"background_color" validate:"required,min=2,max=50"`
}

// ShipmentStatusUpdate - схема для обновления статуса доставки (PUT - полное обновление)
type ShipmentStatusUpdate struct {
	Title           string `json:"title" validate:"required,min=2,max=50"`
	BackgroundColor string `json:"background_color" validate:"required,min=2,max=50"`
}

// ShipmentStatusPatch - схема для частичного обновления (PATCH)
type ShipmentStatusPatch struct {
	Title           *string `json:"title,omitempty" validate:"omitempty,min=2,max=50"`
	BackgroundColor *string `json:"background_color,omitempty" validate:"omitempty,min=2,max=50"`
}

// ShipmentStatusListResponse - схема для ответа со списком статусов
type ShipmentStatusListResponse struct {
	Data  []ShipmentStatusResponse `json:"data"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page,omitempty"`
	Limit int                      `json:"limit,omitempty"`
}
