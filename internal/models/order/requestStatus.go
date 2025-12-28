package models

import (
	"time"

	"gorm.io/gorm"
)

type RequestStatus struct {
	ID              uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Title           string `json:"title" gorm:"size:50"`
	BackgroundColor string `json:"background_color" gorm:"size:50"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (RequestStatus) TableName() string {
	return "request_statuses"
}