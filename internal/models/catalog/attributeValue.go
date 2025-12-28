package models

import (
	"time"

	"gorm.io/gorm"
)

type AttributeValue struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Value       string         `json:"value" gorm:"size:100;not null"`

	ProductID   uint           `json:"product_id" gorm:"uniqueIndex:idx_product_attribute"`
	Product     *Product       `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	AttributeID uint           `json:"attribute_id" gorm:"uniqueIndex:idx_product_attribute"`
	Attribute   *Attribute     `json:"attribute,omitempty" gorm:"foreignKey:AttributeID"`

	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (AttributeValue) TableName() string {
	return "attribute_values"
}