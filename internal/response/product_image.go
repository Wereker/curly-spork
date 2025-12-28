package response

import "time"

type ProductImageResponse struct {
	ID        uint                  `json:"id"`
	ImageURL  string                `json:"image_url"`
	AltText   string                `json:"alt_text,omitempty"`
	IsMain    bool                  `json:"is_main"`
	ProductID uint                  `json:"product_id"`
	Product   *ProductShortResponse `json:"product,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

type ProductImageCreate struct {
	ImageURL  string `json:"image_url" validate:"required,url,max=255"`
	AltText   string `json:"alt_text,omitempty" validate:"max=255"`
	IsMain    bool   `json:"is_main"`
	ProductID uint   `json:"product_id" validate:"required"`
}

type ProductImageUpdate struct {
	ImageURL  string `json:"image_url" validate:"required,url,max=255"`
	AltText   string `json:"alt_text,omitempty" validate:"max=255"`
	IsMain    bool   `json:"is_main"`
	ProductID uint   `json:"product_id" validate:"required"`
}

type ProductImagePatch struct {
	ImageURL  *string `json:"image_url,omitempty" validate:"omitempty,url,max=255"`
	AltText   *string `json:"alt_text,omitempty" validate:"omitempty,max=255"`
	IsMain    *bool   `json:"is_main,omitempty"`
	ProductID *uint   `json:"product_id,omitempty"`
}

type ProductImageListResponse struct {
	Data  []ProductImageResponse `json:"data"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page,omitempty"`
	Limit int                    `json:"limit,omitempty"`
}
