package response

import "time"

type ProductWarehouseResponse struct {
	ID        uint                  `json:"id"`
	Count     uint                  `json:"count"`
	ProductID uint                  `json:"product_id"`
	Product   *ProductShortResponse `json:"product,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

type ProductWarehouseCreate struct {
	Count     uint `json:"count" validate:"omitempty,gte=0"`
	ProductID uint `json:"product_id" validate:"required"`
}

type ProductWarehouseUpdate struct {
	Count     uint `json:"count" validate:"omitempty,gte=0"`
	ProductID uint `json:"product_id" validate:"required"`
}

type ProductWarehousePatch struct {
	Count     *uint `json:"count,omitempty" validate:"omitempty,gte=0"`
	ProductID *uint `json:"product_id,omitempty"`
}

type ProductWarehouseListResponse struct {
	Data  []ProductWarehouseResponse `json:"data"`
	Total int64                      `json:"total"`
	Page  int                        `json:"page,omitempty"`
	Limit int                        `json:"limit,omitempty"`
}
