package response

import "time"

type RequestStatusResponse struct {
	ID              uint      `json:"id"`
	Title           string    `json:"title"`
	BackgroundColor string    `json:"background_color"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type RequestStatusCreate struct {
	Title           string `json:"title" validate:"required,min=2,max=50"`
	BackgroundColor string `json:"background_color" validate:"required,min=2,max=50"`
}

type RequestStatusUpdate struct {
	Title           string `json:"title" validate:"required,min=2,max=50"`
	BackgroundColor string `json:"background_color" validate:"required,min=2,max=50"`
}

type RequestStatusPatch struct {
	Title           *string `json:"title,omitempty" validate:"omitempty,min=2,max=50"`
	BackgroundColor *string `json:"background_color,omitempty" validate:"omitempty,min=2,max=50"`
}

type RequestStatusListResponse struct {
	Data  []RequestStatusResponse `json:"data"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page,omitempty"`
	Limit int                     `json:"limit,omitempty"`
}
