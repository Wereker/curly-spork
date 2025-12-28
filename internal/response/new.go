package response

import "time"

type NewResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	CoverURL  string    `json:"cover_url,omitempty"`
	Number    uint      `json:"number"`
	IsShow    bool      `json:"is_show"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NewCreate struct {
	Title    string `json:"title" validate:"required,min=2,max=128"`
	Subtitle string `json:"subtitle" validate:"required,min=2,max=256"`
	CoverURL string `json:"cover_url,omitempty" validate:"omitempty,url,max=128"`
	Number   uint   `json:"number" validate:"gte=1"`
	IsShow   bool   `json:"is_show"`
}

type NewUpdate struct {
	Title    string `json:"title" validate:"required,min=2,max=128"`
	Subtitle string `json:"subtitle" validate:"required,min=2,max=256"`
	CoverURL string `json:"cover_url,omitempty" validate:"omitempty,url,max=128"`
	Number   uint   `json:"number" validate:"gte=1"`
	IsShow   bool   `json:"is_show"`
}

type NewPatch struct {
	Title    *string `json:"title,omitempty" validate:"omitempty,min=2,max=128"`
	Subtitle *string `json:"subtitle,omitempty" validate:"omitempty,min=2,max=256"`
	CoverURL *string `json:"cover_url,omitempty" validate:"omitempty,url,max=128"`
	Number   *uint   `json:"number,omitempty" validate:"omitempty,gte=1"`
	IsShow   *bool   `json:"is_show,omitempty"`
}

type NewListResponse struct {
	Data  []NewResponse `json:"data"`
	Total int64         `json:"total"`
	Page  int           `json:"page,omitempty"`
	Limit int           `json:"limit,omitempty"`
}
