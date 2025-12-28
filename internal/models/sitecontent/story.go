package models

import (
	"time"

	"gorm.io/gorm"
)

type Story struct {
	ID       uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Title    string `json:"title" gorm:"size:256"`
	CoverURL string `json:"cover_url,omitempty" gorm:"size:128"`
	IsShow   bool   `json:"is_show" gorm:"default:true"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Story) TableName() string {
	return "stories"
}
