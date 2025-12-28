package response

import (
	"time"
)

type OrderResponse struct {
	ID            uint                 `json:"id"`
	Number        string               `json:"number"`
	Date          time.Time            `json:"date"`
	Price         float64              `json:"price"`
	Discount      float64              `json:"discount"`
	IsUrgent      bool                 `json:"is_urgent"`
	TotalAmount   float64              `json:"total_amount"`
	UserID        uint                 `json:"user_id"`
	User          *UserResponse        `json:"user,omitempty"`
	ShipmentID    uint                 `json:"shipment_id"`
	Shipment      *ShipmentResponse    `json:"shipment,omitempty"`
	PaymentID     uint                 `json:"payment_id"`
	Payment       *PaymentResponse     `json:"payment,omitempty"`
	OrderStatusID uint                 `json:"order_status_id"`
	OrderStatus   *OrderStatusResponse `json:"order_status,omitempty"`
	OrderItems    []OrderItemResponse  `json:"order_items,omitempty"`
	Request       *RequestResponse     `json:"request,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

type OrderCreate struct {
	Price         float64 `json:"price" validate:"required,gte=0"`
	Discount      float64 `json:"discount" validate:"gte=0"`
	IsUrgent      bool    `json:"is_urgent"`
	UserID        uint    `json:"user_id" validate:"required"`
	ShipmentID    uint    `json:"shipment_id" validate:"required"`
	PaymentID     uint    `json:"payment_id" validate:"required"`
	OrderStatusID uint    `json:"order_status_id" validate:"required"`
}

type OrderUpdate struct {
	Price         float64 `json:"price" validate:"required,gte=0"`
	Discount      float64 `json:"discount" validate:"gte=0"`
	IsUrgent      bool    `json:"is_urgent"`
	UserID        uint    `json:"user_id" validate:"required"`
	ShipmentID    uint    `json:"shipment_id" validate:"required"`
	PaymentID     uint    `json:"payment_id" validate:"required"`
	OrderStatusID uint    `json:"order_status_id" validate:"required"`
}

type OrderPatch struct {
	Price         *float64 `json:"price,omitempty" validate:"omitempty,gte=0"`
	Discount      *float64 `json:"discount,omitempty" validate:"omitempty,gte=0"`
	IsUrgent      *bool    `json:"is_urgent,omitempty"`
	UserID        *uint    `json:"user_id,omitempty"`
	ShipmentID    *uint    `json:"shipment_id,omitempty"`
	PaymentID     *uint    `json:"payment_id,omitempty"`
	OrderStatusID *uint    `json:"order_status_id,omitempty"`
}

type OrderListResponse struct {
	Data  []OrderResponse `json:"data"`
	Total int64           `json:"total"`
	Page  int             `json:"page,omitempty"`
	Limit int             `json:"limit,omitempty"`
}
