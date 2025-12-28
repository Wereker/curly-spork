package models

import (
	"time"
	"app/internal/models/catalog"
	"gorm.io/gorm"
)

type OrderItem struct {
	ID        uint `json:"id" gorm:"primaryKey;autoIncrement"`
 	Quantity  uint `json:"quantity" gorm:"not null"`
	UnitPrice float64 `json:"unit_price" gorm:"type:decimal(10,2);not null"`

	OrderID uint `json:"order_id"`
	Order *Order `json:"order" gorm:"foreignKey:OrderID"`
	ProductID uint `json:"product_id"`
	Product *models.Product `json:"product" gorm:"foreignKey:ProductID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (OrderItem) TableName() string {
	return  "order_items"
}

func (ot *OrderItem) GetTotalPrice() float64 {
	return float64(ot.Quantity) * ot.UnitPrice
}