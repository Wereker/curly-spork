package models

import (
	"time"

	"gorm.io/gorm"
)

type PaymentStatus struct {
	ID              uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Title           string `json:"title" gorm:"size:50; not null"`
	BackgroundColor string `json:"background_color" gorm:"size:50; not null"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (PaymentStatus) TableName() string {
	return "payment_statuses"
}