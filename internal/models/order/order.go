package models

import (
	models "app/internal/models/user"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID       uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Number   string    `json:"number" gorm:"size:50;uniqueIndex;not null"`
	Date     time.Time `json:"date" gorm:"autoCreateTime"`
	Price    float64   `json:"price" gorm:"type:decimal(12,2);default:0"`
	Discount float64   `json:"discount" gorm:"type:decimal(12,2);default:0"`
	IsUrgent bool      `json:"is_urgent" gorm:"default:false"`

	UserID        uint         `json:"user_id"`
	User          *models.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ShipmentID    uint         `json:"shipment_id"`
	Shipment      *Shipment    `json:"shipment,omitempty" gorm:"foreignKey:ShipmentID"`
	PaymentID     uint         `json:"payment_id"`
	Payment       *Payment     `json:"payment,omitempty" gorm:"foreignKey:PaymentID"`
	OrderStatusID uint         `json:"order_status_id"`
	OrderStatus   *OrderStatus `json:"order_status" gorm:"foreignKey:OrderStatusID"`

	OrderItems []OrderItem `json:"order_items,omitempty" gorm:"foreignKey:OrderID"`
	Request    *Request    `json:"request,omitempty" gorm:"foreignKey:OrderID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Order) TableName() string {
	return "orders"
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.Number == "" {
		o.Number = generateOrderNumber()
	}
	return o.ensureUniqueOrderNumber(tx)
}

func generateOrderNumber() string {
	id := uuid.New()
	uuidStr := strings.ReplaceAll(id.String(), "-", "")
	if len(uuidStr) > 20 {
		uuidStr = uuidStr[:20]
	}
	return fmt.Sprintf("A-%s", strings.ToUpper(uuidStr))
}

func (o *Order) ensureUniqueOrderNumber(tx *gorm.DB) error {
	baseNumber := o.Number
	counter := 1

	for {
		var count int64
		query := tx.Model(&Order{}).Where("number = ?", o.Number)

		if o.ID != 0 {
			query = query.Where("id != ?", o.ID)
		}

		query.Count(&count)

		if count == 0 {
			break
		}

		o.Number = fmt.Sprintf("%s-%d", baseNumber, counter)
		counter++

		if counter > 100 {
			return fmt.Errorf("failed to generate unique order number after 100 attempts")
		}
	}

	return nil
}

func (o *Order) GetTotalAmount() float64 {
	total := o.Price - o.Discount

	if o.Shipment != nil {
		total += o.Shipment.Price
	}

	return total
}
