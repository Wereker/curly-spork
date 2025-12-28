package response

import (
	"time"
)

type CommonBlockResponse struct {
	ID             uint                       `json:"id"`
	Title          string                     `json:"title"`
	CoverURL       string                     `json:"cover_url,omitempty"`
	IsShow         bool                       `json:"is_show"`
	CategoryID     *uint                      `json:"category_id,omitempty"`
	Category       *CategoryShortResponse     `json:"category,omitempty"`
	ManufacturerID *uint                      `json:"manufacturer_id,omitempty"`
	Manufacturer   *ManufacturerShortResponse `json:"manufacturer,omitempty"`
	AttributeID    *uint                      `json:"attribute_id,omitempty"`
	Attribute      *AttributeShortResponse    `json:"attribute,omitempty"`
	CreatedAt      time.Time                  `json:"created_at"`
	UpdatedAt      time.Time                  `json:"updated_at"`
}

type CommonBlockCreate struct {
	Title          string `json:"title" validate:"required,min=2,max=256"`
	CoverURL       string `json:"cover_url,omitempty" validate:"omitempty,url"`
	IsShow         bool   `json:"is_show"`
	CategoryID     *uint  `json:"category_id,omitempty"`
	ManufacturerID *uint  `json:"manufacturer_id,omitempty"`
	AttributeID    *uint  `json:"attribute_id,omitempty"`
}

type CommonBlockUpdate struct {
	Title          string `json:"title" validate:"required,min=2,max=256"`
	CoverURL       string `json:"cover_url,omitempty" validate:"omitempty,url"`
	IsShow         bool   `json:"is_show"`
	CategoryID     *uint  `json:"category_id,omitempty"`
	ManufacturerID *uint  `json:"manufacturer_id,omitempty"`
	AttributeID    *uint  `json:"attribute_id,omitempty"`
}

type CommonBlockPatch struct {
	Title          *string `json:"title,omitempty" validate:"omitempty,min=2,max=256"`
	CoverURL       *string `json:"cover_url,omitempty" validate:"omitempty,url"`
	IsShow         *bool   `json:"is_show,omitempty"`
	CategoryID     *uint   `json:"category_id,omitempty"`
	ManufacturerID *uint   `json:"manufacturer_id,omitempty"`
	AttributeID    *uint   `json:"attribute_id,omitempty"`
}

type CommonBlockListResponse struct {
	Data  []CommonBlockResponse `json:"data"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page,omitempty"`
	Limit int                   `json:"limit,omitempty"`
}
