package models

import (
	"time"

	"gorm.io/gorm"
)

type PaymentMethod struct {
	ID          uint    `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string  `json:"title" gorm:"size:50; not null"`
	Description *string `json:"description,omitempty" gorm:"type:text"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (PaymentMethod) TableName() string {
	return "payment_methods"
}