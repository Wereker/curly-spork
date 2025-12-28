package models

import (
	"time"

	"gorm.io/gorm"
)

type ProductImage struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	ImageURL  string         `json:"image_url" gorm:"size:255;not null"`
	AltText   string         `json:"alt_text,omitempty" gorm:"size:255"`
	IsMain    bool           `json:"is_main" gorm:"default:false"`

	ProductID uint           `json:"product_id"`
	Product   *Product       `json:"product,omitempty" gorm:"foreignKey:ProductID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (ProductImage) TableName() string {
	return "product_images"
}