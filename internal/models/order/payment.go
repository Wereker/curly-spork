package models

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	PaymentDate time.Time `json:"payment_date" gorm:"autoCreateTime"`

	PaymentMethodID uint           `json:"payment_method_id"`
	PaymentMethod   *PaymentMethod `json:"payment_method" gorm:"foreignKey:PaymentMethodID"`
	PaymentStatusID uint           `json:"payment_status_id"`
	PaymentStatus   *PaymentStatus `json:"payment_status" gorm:"foreignKey:PaymentStatusID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Payment) TableName() string {
	return "payments"
}