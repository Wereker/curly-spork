package response

import "time"

type ManufacturerResponse struct {
	ID            uint      `json:"id"`
	Slug          string    `json:"slug"`
	Title         string    `json:"title"`
	Country       string    `json:"country,omitempty"`
	Description   string    `json:"description,omitempty"`
	ProductsCount int64     `json:"products_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ManufacturerShortResponse struct {
	ID    uint   `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type ManufacturerCreate struct {
	Slug        string `json:"slug,omitempty" validate:"omitempty,slug,min=2,max=100"`
	Title       string `json:"title" validate:"required,min=2,max=50"`
	Country     string `json:"country,omitempty" validate:"omitempty,max=50"`
	Description string `json:"description,omitempty"`
}

type ManufacturerUpdate struct {
	Slug        string `json:"slug" validate:"slug,min=2,max=100"`
	Title       string `json:"title" validate:"required,min=2,max=50"`
	Country     string `json:"country,omitempty" validate:"omitempty,max=50"`
	Description string `json:"description,omitempty"`
}

type ManufacturerPatch struct {
	Slug        *string `json:"slug,omitempty" validate:"omitempty,slug,min=2,max=100"`
	Title       *string `json:"title,omitempty" validate:"omitempty,min=2,max=50"`
	Country     *string `json:"country,omitempty" validate:"omitempty,max=50"`
	Description *string `json:"description,omitempty"`
}

type ManufacturerListResponse struct {
	Data  []ManufacturerResponse `json:"data"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page,omitempty"`
	Limit int                    `json:"limit,omitempty"`
}
