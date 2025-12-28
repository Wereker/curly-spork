package models

import (
	"time"

	"gorm.io/gorm"
)

type New struct {
	ID       uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Title    string `json:"title" gorm:"size:128;not null"`
	Subtitle string `json:"subtitle" gorm:"size:256;not null"`
	CoverURL string `json:"cover_url,omitempty" gorm:"size:128"`
	Number   uint   `json:"number" gorm:"default:1"`
	IsShow   bool   `json:"is_show" gorm:"default:true"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (New) TableName() string {
	return "news"
}
