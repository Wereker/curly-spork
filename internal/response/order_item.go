package response

import "time"

type OrderItemResponse struct {
	ID         uint                  `json:"id"`
	Quantity   uint                  `json:"quantity"`
	UnitPrice  float64               `json:"unit_price"`
	TotalPrice float64               `json:"total_price"`
	OrderID    uint                  `json:"order_id"`
	ProductID  uint                  `json:"product_id"`
	Product    *ProductShortResponse `json:"product,omitempty"`
	CreatedAt  time.Time             `json:"created_at"`
	UpdatedAt  time.Time             `json:"updated_at"`
}

type OrderItemCreate struct {
	Quantity  uint    `json:"quantity" validate:"required,gt=0"`
	UnitPrice float64 `json:"unit_price" validate:"required,gt=0"`
	OrderID   uint    `json:"order_id" validate:"required"`
	ProductID uint    `json:"product_id" validate:"required"`
}

type OrderItemUpdate struct {
	Quantity  uint    `json:"quantity" validate:"required,gt=0"`
	UnitPrice float64 `json:"unit_price" validate:"required,gt=0"`
	OrderID   uint    `json:"order_id" validate:"required"`
	ProductID uint    `json:"product_id" validate:"required"`
}

type OrderItemPatch struct {
	Quantity  *uint    `json:"quantity,omitempty" validate:"omitempty,gt=0"`
	UnitPrice *float64 `json:"unit_price,omitempty" validate:"omitempty,gt=0"`
	OrderID   *uint    `json:"order_id,omitempty"`
	ProductID *uint    `json:"product_id,omitempty"`
}

type OrderItemListResponse struct {
	Data  []OrderItemResponse `json:"data"`
	Total int64               `json:"total"`
	Page  int                 `json:"page,omitempty"`
	Limit int                 `json:"limit,omitempty"`
}
