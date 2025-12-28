package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Shipment struct {
	ID      uint    `json:"id" gorm:"primaryKey;autoIncrement"`
	Number  string  `json:"number" gorm:"size:50;uniqueIndex;not null"`
	Price   float64 `json:"price" gorm:"type:decimal(12,2);default:0"`
	Address string  `json:"address,omitempty" gorm:"size:256;default:''"`

	ShipmentMethodID uint           `json:"shipment_method_id"`
	ShipmentMethod   *ShipmentMethod `json:"shipment_method" gorm:"foreignKey:ShipmentMethodID"`
	ShipmentStatusID uint           `json:"shipment_status_id"`
	ShipmentStatus   *ShipmentStatus `json:"shipment_status" gorm:"foreignKey:ShipmentStatusID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Shipment) TableName() string {
	return "shipments"
}

func (s *Shipment) BeforeCreate(tx *gorm.DB) error {
    if s.Number == "" {
        s.Number = generateTrackingNumber()
    }
    
    return s.ensureUniqueTrackingNumber(tx)
}

func generateTrackingNumber() string {
    id := uuid.New()
    uuidStr := strings.ReplaceAll(id.String(), "-", "")
    if len(uuidStr) > 20 {
        uuidStr = uuidStr[:20]
    }
    return fmt.Sprintf("S-%s", strings.ToUpper(uuidStr))
}

func (s *Shipment) ensureUniqueTrackingNumber(tx *gorm.DB) error {
    baseNumber := s.Number
    counter := 1
    
    for {
        var count int64
        query := tx.Model(&Shipment{}).Where("number = ?", s.Number)
        
        if s.ID != 0 {
            query = query.Where("id != ?", s.ID)
        }
        
        query.Count(&count)
        
        if count == 0 {
            break
        }
        
        s.Number = fmt.Sprintf("%s-%d", baseNumber, counter)
        counter++
        
        if counter > 100 {
            return fmt.Errorf("failed to generate unique tracking number after 100 attempts")
        }
    }
    
    return nil
}