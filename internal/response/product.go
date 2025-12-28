package response

import (
	models "app/internal/models/catalog"
	"time"
)

type ProductResponse struct {
	ID               uint                       `json:"id"`
	Slug             string                     `json:"slug"`
	Title            string                     `json:"title"`
	ShortDescription string                     `json:"short_description"`
	Description      *string                    `json:"description,omitempty"`
	Price            float64                    `json:"price"`
	Discount         *uint                      `json:"discount,omitempty"`
	DiscountedPrice  float64                    `json:"discounted_price"`
	CategoryID       uint                       `json:"category_id"`
	Category         *CategoryShortResponse     `json:"category,omitempty"`
	ManufacturerID   uint                       `json:"manufacturer_id"`
	Manufacturer     *ManufacturerShortResponse `json:"manufacturer,omitempty"`
	Attributes       []ProductAttributeResponse `json:"attributes,omitempty"`
	Images           []ProductImageResponse     `json:"images,omitempty"`
	Warehouse        *ProductWarehouseResponse  `json:"warehouse,omitempty"`
	CreatedAt        time.Time                  `json:"created_at"`
	UpdatedAt        time.Time                  `json:"updated_at"`
}

type ProductShortResponse struct {
	ID    uint    `json:"id"`
	Title string  `json:"title"`
	Slug  string  `json:"slug"`
	Price float64 `json:"price"`
}

type ProductAttributeResponse struct {
	AttributeID    uint                 `json:"attribute_id"`
	AttributeSlug  string               `json:"attribute_slug"`
	AttributeTitle string               `json:"attribute_title"`
	AttributeType  models.AttributeType `json:"attribute_type"`
	Value          string               `json:"value"`
}

type ProductCreate struct {
	Slug             string  `json:"slug" validate:"required,slug,min=2,max=100"`
	Title            string  `json:"title" validate:"required,min=2,max=100"`
	ShortDescription string  `json:"short_description" validate:"required,min=2,max=100"`
	Description      *string `json:"description,omitempty"`
	Price            float64 `json:"price" validate:"required,gt=0"`
	Discount         *uint   `json:"discount,omitempty" validate:"omitempty,lte=100"`
	CategoryID       uint    `json:"category_id" validate:"required"`
	ManufacturerID   uint    `json:"manufacturer_id" validate:"required"`
}

type ProductUpdate struct {
	Slug             string  `json:"slug" validate:"required,slug,min=2,max=100"`
	Title            string  `json:"title" validate:"required,min=2,max=100"`
	ShortDescription string  `json:"short_description" validate:"required,min=2,max=100"`
	Description      *string `json:"description,omitempty"`
	Price            float64 `json:"price" validate:"required,gt=0"`
	Discount         *uint   `json:"discount,omitempty" validate:"omitempty,lte=100"`
	CategoryID       uint    `json:"category_id" validate:"required"`
	ManufacturerID   uint    `json:"manufacturer_id" validate:"required"`
}

type ProductPatch struct {
	Slug             *string  `json:"slug,omitempty" validate:"omitempty,slug,min=2,max=100"`
	Title            *string  `json:"title,omitempty" validate:"omitempty,min=2,max=100"`
	ShortDescription *string  `json:"short_description,omitempty" validate:"omitempty,min=2,max=100"`
	Description      *string  `json:"description,omitempty"`
	Price            *float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
	Discount         *uint    `json:"discount,omitempty" validate:"omitempty,lte=100"`
	CategoryID       *uint    `json:"category_id,omitempty"`
	ManufacturerID   *uint    `json:"manufacturer_id,omitempty"`
}

type ProductListResponse struct {
	Data  []ProductResponse `json:"data"`
	Total int64             `json:"total"`
	Page  int               `json:"page,omitempty"`
	Limit int               `json:"limit,omitempty"`
}
