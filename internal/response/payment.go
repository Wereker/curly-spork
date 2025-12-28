package response

import "time"

type PaymentResponse struct {
	ID              uint                   `json:"id"`
	PaymentDate     time.Time              `json:"payment_date"`
	PaymentMethodID uint                   `json:"payment_method_id"`
	PaymentMethod   *PaymentMethodResponse `json:"payment_method,omitempty"`
	PaymentStatusID uint                   `json:"payment_status_id"`
	PaymentStatus   *PaymentStatusResponse `json:"payment_status,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type PaymentCreate struct {
	PaymentMethodID uint `json:"payment_method_id" validate:"required"`
	PaymentStatusID uint `json:"payment_status_id" validate:"required"`
}

type PaymentUpdate struct {
	PaymentMethodID uint `json:"payment_method_id" validate:"required"`
	PaymentStatusID uint `json:"payment_status_id" validate:"required"`
}

type PaymentPatch struct {
	PaymentMethodID *uint `json:"payment_method_id,omitempty"`
	PaymentStatusID *uint `json:"payment_status_id,omitempty"`
}

type PaymentListResponse struct {
	Data  []PaymentResponse `json:"data"`
	Total int64             `json:"total"`
	Page  int               `json:"page,omitempty"`
	Limit int               `json:"limit,omitempty"`
}
