package models

import (
	"time"

	"gorm.io/gorm"
)

type ProductWarehouse struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`
	Count uint `json:"count" gorm:"default:10"`

	ProductID uint     `json:"product_id" gorm:"uniqueIndex"`
	Product  *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (ProductWarehouse) TableName() string {
	return "product_warehouses"
}