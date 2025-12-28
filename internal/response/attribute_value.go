// internal/response/attribute_value.go
package response

import (
	"time"
)

type AttributeValueResponse struct {
	ID          uint                    `json:"id"`
	Value       string                  `json:"value"`
	ProductID   uint                    `json:"product_id"`
	Product     *ProductShortResponse   `json:"product,omitempty"`
	AttributeID uint                    `json:"attribute_id"`
	Attribute   *AttributeShortResponse `json:"attribute,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

type AttributeValueCreate struct {
	Value       string `json:"value" validate:"required,min=1,max=100"`
	ProductID   uint   `json:"product_id" validate:"required"`
	AttributeID uint   `json:"attribute_id" validate:"required"`
}

type AttributeValueUpdate struct {
	Value       string `json:"value" validate:"required,min=1,max=100"`
	ProductID   uint   `json:"product_id" validate:"required"`
	AttributeID uint   `json:"attribute_id" validate:"required"`
}

type AttributeValuePatch struct {
	Value       *string `json:"value,omitempty" validate:"omitempty,min=1,max=100"`
	ProductID   *uint   `json:"product_id,omitempty"`
	AttributeID *uint   `json:"attribute_id,omitempty"`
}

type AttributeValueListResponse struct {
	Data  []AttributeValueResponse `json:"data"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page,omitempty"`
	Limit int                      `json:"limit,omitempty"`
}
