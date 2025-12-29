package response

import (
	models "app/internal/models/catalog"
	"time"
)

type AttributeResponse struct {
	ID            uint                          `json:"id"`
	Slug          string                        `json:"slug"`
	Title         string                        `json:"title"`
	AttributeType models.AttributeType          `json:"attribute_type"`
	Values        []AttributeValueShortResponse `json:"values,omitempty"`
	CreatedAt     time.Time                     `json:"created_at"`
	UpdatedAt     time.Time                     `json:"updated_at"`
}

type AttributeShortResponse struct {
	ID            uint                 `json:"id"`
	Slug          string               `json:"slug"`
	Title         string               `json:"title"`
	AttributeType models.AttributeType `json:"attribute_type"`
}

type AttributeValueShortResponse struct {
	ID    uint   `json:"id"`
	Value string `json:"value"`
}

type AttributeCreate struct {
	Slug          string               `json:"slug" validate:"omitempty,slug,min=2,max=100"`
	Title         string               `json:"title" validate:"required,min=2,max=100"`
	AttributeType models.AttributeType `json:"attribute_type" validate:"required,oneof=text number select multi_select"`
}

type AttributeUpdate struct {
	Slug          string               `json:"slug" validate:"omitempty,slug,min=2,max=100"`
	Title         string               `json:"title" validate:"required,min=2,max=100"`
	AttributeType models.AttributeType `json:"attribute_type" validate:"required,oneof=text number select multi_select"`
}

type AttributePatch struct {
	Slug          *string               `json:"slug,omitempty" validate:"omitempty,slug,min=2,max=100"`
	Title         *string               `json:"title,omitempty" validate:"omitempty,min=2,max=100"`
	AttributeType *models.AttributeType `json:"attribute_type,omitempty" validate:"omitempty,oneof=text number select multi_select"`
}

type AttributeListResponse struct {
	Data  []AttributeResponse `json:"data"`
	Total int64               `json:"total"`
	Page  int                 `json:"page,omitempty"`
	Limit int                 `json:"limit,omitempty"`
}
