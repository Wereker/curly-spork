package response

import "time"

type StoryResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	CoverURL  string    `json:"cover_url,omitempty"`
	IsShow    bool      `json:"is_show"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StoryCreate struct {
	Title    string `json:"title" validate:"required,min=2,max=256"`
	CoverURL string `json:"cover_url,omitempty" validate:"omitempty,url,max=128"`
	IsShow   bool   `json:"is_show"`
}

type StoryUpdate struct {
	Title    string `json:"title" validate:"required,min=2,max=256"`
	CoverURL string `json:"cover_url,omitempty" validate:"omitempty,url,max=128"`
	IsShow   bool   `json:"is_show"`
}

type StoryPatch struct {
	Title    *string `json:"title,omitempty" validate:"omitempty,min=2,max=256"`
	CoverURL *string `json:"cover_url,omitempty" validate:"omitempty,url,max=128"`
	IsShow   *bool   `json:"is_show,omitempty"`
}

type StoryListResponse struct {
	Data  []StoryResponse `json:"data"`
	Total int64           `json:"total"`
	Page  int             `json:"page,omitempty"`
	Limit int             `json:"limit,omitempty"`
}
