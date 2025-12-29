package response

import "time"

type CategoryResponse struct {
	ID            uint                    `json:"id"`
	Slug          string                  `json:"slug"`
	Title         string                  `json:"title"`
	Description   string                  `json:"description,omitempty"`
	ParentID      *uint                   `json:"parent_id,omitempty"`
	Parent        *CategoryShortResponse  `json:"parent,omitempty"`
	Children      []CategoryShortResponse `json:"children,omitempty"`
	CoverURL      string                  `json:"cover_url,omitempty"`
	ProductsCount int64                   `json:"products_count"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
}

type CategoryShortResponse struct {
	ID       uint   `json:"id"`
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	CoverURL string `json:"cover_url,omitempty"`
}

type CategoryCreate struct {
	Slug        string `json:"slug" validate:"omitempty,slug,min=2,max=100"`
	Title       string `json:"title" validate:"required,min=2,max=50"`
	Description string `json:"description,omitempty"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	CoverURL    string `json:"cover_url,omitempty" validate:"omitempty,url,max=255"`
}

type CategoryUpdate struct {
	Slug        string `json:"slug" validate:"omitempty,slug,min=2,max=100"`
	Title       string `json:"title" validate:"required,min=2,max=50"`
	Description string `json:"description,omitempty"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	CoverURL    string `json:"cover_url,omitempty" validate:"omitempty,url,max=255"`
}

type CategoryPatch struct {
	Slug        *string `json:"slug,omitempty" validate:"omitempty,slug,min=2,max=100"`
	Title       *string `json:"title,omitempty" validate:"omitempty,min=2,max=50"`
	Description *string `json:"description,omitempty"`
	ParentID    *uint   `json:"parent_id,omitempty"`
	CoverURL    *string `json:"cover_url,omitempty" validate:"omitempty,url,max=255"`
}

type CategoryListResponse struct {
	Data  []CategoryResponse `json:"data"`
	Total int64              `json:"total"`
	Page  int                `json:"page,omitempty"`
	Limit int                `json:"limit,omitempty"`
}

type CategoryTreeResponse struct {
	ID          uint                   `json:"id"`
	Slug        string                 `json:"slug"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	CoverURL    string                 `json:"cover_url,omitempty"`
	Children    []CategoryTreeResponse `json:"children,omitempty"`
}
