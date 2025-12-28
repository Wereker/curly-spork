package response

import "time"

type RightResponse struct {
	ID               uint      `json:"id"`
	Title            string    `json:"title"`
	Description      *string   `json:"description,omitempty"`
	CanViewCatalog   bool      `json:"can_view_catalog"`
	CanMakeOrder     bool      `json:"can_make_order"`
	CanManageOrder   bool      `json:"can_manage_order"`
	CanManageProduct bool      `json:"can_manage_product"`
	CanManageUser    bool      `json:"can_manage_user"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type RightCreate struct {
	Title            string  `json:"title" validate:"required,min=2,max=50"`
	Description      *string `json:"description,omitempty"`
	CanViewCatalog   bool    `json:"can_view_catalog"`
	CanMakeOrder     bool    `json:"can_make_order"`
	CanManageOrder   bool    `json:"can_manage_order"`
	CanManageProduct bool    `json:"can_manage_product"`
	CanManageUser    bool    `json:"can_manage_user"`
}

type RightUpdate struct {
	Title            string  `json:"title" validate:"required,min=2,max=50"`
	Description      *string `json:"description,omitempty"`
	CanViewCatalog   bool    `json:"can_view_catalog"`
	CanMakeOrder     bool    `json:"can_make_order"`
	CanManageOrder   bool    `json:"can_manage_order"`
	CanManageProduct bool    `json:"can_manage_product"`
	CanManageUser    bool    `json:"can_manage_user"`
}

type RightPatch struct {
	Title            *string `json:"title,omitempty" validate:"omitempty,min=2,max=50"`
	Description      *string `json:"description,omitempty"`
	CanViewCatalog   *bool   `json:"can_view_catalog,omitempty"`
	CanMakeOrder     *bool   `json:"can_make_order,omitempty"`
	CanManageOrder   *bool   `json:"can_manage_order,omitempty"`
	CanManageProduct *bool   `json:"can_manage_product,omitempty"`
	CanManageUser    *bool   `json:"can_manage_user,omitempty"`
}

type RightListResponse struct {
	Data  []RightResponse `json:"data"`
	Total int64           `json:"total"`
	Page  int             `json:"page,omitempty"`
	Limit int             `json:"limit,omitempty"`
}
