package models

import (
	"fmt"
	"time"

	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

type Product struct {
	ID               uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Slug             string         `json:"slug" gorm:"size:100;uniqueIndex"`
	Title            string         `json:"title" gorm:"size:100;not null"`
	ShortDescription string         `json:"short_description" gorm:"size:100;not null"`
	Description      *string        `json:"description,omitempty" gorm:"type:text"`
	Price            float64        `json:"price" gorm:"type:decimal(10,2);not null"`
	Discount         *uint          `json:"discount,omitempty" gorm:"default:0"`

	CategoryID       uint           `json:"category_id"`
	Category         *Category      `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	ManufacturerID   uint           `json:"manufacturer_id"`
	Manufacturer     *Manufacturer  `json:"manufacturer,omitempty" gorm:"foreignKey:ManufacturerID"`

	AttributeValues   []AttributeValue   `json:"attribute_values,omitempty" gorm:"foreignKey:ProductID"`
	ProductImages     []ProductImage     `json:"product_images,omitempty" gorm:"foreignKey:ProductID"`
	ProductWarehouse  *ProductWarehouse  `json:"product_warehouse,omitempty" gorm:"foreignKey:ProductID"`

	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Product) TableName() string {
	return "products"
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.Slug == "" {
		p.Slug = slug.Make(p.Title)
	}

	var count int64
	baseSlug := p.Slug
	counter := 1

	for {
		if counter > 1 {
			p.Slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		}

		tx.Model(&Product{}).
			Where("slug = ?", p.Slug).
			Count(&count)

		if count == 0 {
			break
		}

		counter++
	}
	
	return nil
}

func (p *Product) GetDiscountedPrice() float64 {
	if p.Discount != nil && *p.Discount > 0 {
		discountPercent := float64(*p.Discount) / 100.0
		return p.Price * (1.0 - discountPercent)
	}
	return p.Price
}

func (p *Product) GetDiscountAmount() float64 {
	return p.Price - p.GetDiscountedPrice()
}