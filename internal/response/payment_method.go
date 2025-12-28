package response

import "time"

type PaymentMethodResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PaymentMethodCreate struct {
	Title       string  `json:"title" validate:"required,min=2,max=50"`
	Description *string `json:"description,omitempty"`
}

type PaymentMethodUpdate struct {
	Title       string  `json:"title" validate:"required,min=2,max=50"`
	Description *string `json:"description,omitempty"`
}

type PaymentMethodPatch struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,min=2,max=50"`
	Description *string `json:"description,omitempty"`
}

type PaymentMethodListResponse struct {
	Data  []PaymentMethodResponse `json:"data"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page,omitempty"`
	Limit int                     `json:"limit,omitempty"`
}
