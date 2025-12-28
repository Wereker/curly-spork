package response

import "time"

type RequestResponse struct {
	ID              uint                   `json:"id"`
	Number          string                 `json:"number"`
	UserComment     string                 `json:"user_comment,omitempty"`
	ManagerComment  string                 `json:"manager_comment,omitempty"`
	OrderID         uint                   `json:"order_id"`
	Order           *OrderShortResponse    `json:"order,omitempty"`
	UserID          uint                   `json:"user_id"`
	User            *UserShortResponse     `json:"user,omitempty"`
	RequestStatusID uint                   `json:"request_status_id"`
	RequestStatus   *RequestStatusResponse `json:"request_status,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type OrderShortResponse struct {
	ID     uint    `json:"id"`
	Number string  `json:"number"`
	Price  float64 `json:"price"`
}

type UserShortResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type RequestCreate struct {
	UserComment     string `json:"user_comment,omitempty" validate:"max=128"`
	ManagerComment  string `json:"manager_comment,omitempty" validate:"max=128"`
	OrderID         uint   `json:"order_id" validate:"required"`
	UserID          uint   `json:"user_id" validate:"required"`
	RequestStatusID uint   `json:"request_status_id" validate:"required"`
}

type RequestUpdate struct {
	UserComment     string `json:"user_comment,omitempty" validate:"max=128"`
	ManagerComment  string `json:"manager_comment,omitempty" validate:"max=128"`
	OrderID         uint   `json:"order_id" validate:"required"`
	UserID          uint   `json:"user_id" validate:"required"`
	RequestStatusID uint   `json:"request_status_id" validate:"required"`
}

type RequestPatch struct {
	UserComment     *string `json:"user_comment,omitempty" validate:"omitempty,max=128"`
	ManagerComment  *string `json:"manager_comment,omitempty" validate:"omitempty,max=128"`
	OrderID         *uint   `json:"order_id,omitempty"`
	UserID          *uint   `json:"user_id,omitempty"`
	RequestStatusID *uint   `json:"request_status_id,omitempty"`
}

type RequestListResponse struct {
	Data  []RequestResponse `json:"data"`
	Total int64             `json:"total"`
	Page  int               `json:"page,omitempty"`
	Limit int               `json:"limit,omitempty"`
}
