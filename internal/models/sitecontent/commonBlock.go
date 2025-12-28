package models

import (
	catalog "app/internal/models/catalog"
	"errors"
	"time"

	"gorm.io/gorm"
)

type CommonBlock struct {
	ID       uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Title    string `json:"title" gorm:"size:256"`
	CoverURL string `json:"cover_url,omitempty"`
	IsShow   bool   `json:"is_show" gorm:"default:true"`

	CategoryID     *uint                 `json:"category_id,omitempty"`
	Category       *catalog.Category     `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	ManufacturerID *uint                 `json:"manufacturer_id,omitempty"`
	Manufacturer   *catalog.Manufacturer `json:"manufacturer,omitempty" gorm:"foreignKey:ManufacturerID"`
	AttributeID    *uint                 `json:"attribute_id,omitempty"`
	Attribute      *catalog.Attribute    `json:"attribute,omitempty" gorm:"foreignKey:AttributeID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (CommonBlock) TableName() string {
	return "common_blocks"
}

func (cb *CommonBlock) BeforeSave(tx *gorm.DB) error {
	return cb.Validate()
}

func (cb *CommonBlock) Validate() error {
	filled := 0
	fields := []*uint{cb.CategoryID, cb.ManufacturerID, cb.AttributeID}

	for _, field := range fields {
		if field != nil {
			filled++
		}
	}

	if filled == 0 {
		return errors.New("укажите хотя бы одну связь: категорию, производителя или атрибут")
	}

	if filled > 1 {
		return errors.New("можно выбрать только одну связь: либо категорию, либо производителя, либо атрибут")
	}

	return nil
}

func (cb *CommonBlock) GetFilteredProducts(db *gorm.DB) ([]catalog.Product, error) {
	var products []catalog.Product
	query := db.Model(&catalog.Product{}).Limit(4)

	switch {
	case cb.CategoryID != nil:
		query = query.Where("category_id = ?", *cb.CategoryID)
	case cb.ManufacturerID != nil:
		query = query.Where("manufacturer_id = ?", *cb.ManufacturerID)
	case cb.AttributeID != nil:
		query = query.Joins(
			"JOIN attribute_values ON products.id = attribute_values.product_id",
		).Where("attribute_values.attribute_id = ?", *cb.AttributeID).Distinct()
	default:
		return []catalog.Product{}, nil
	}

	err := query.Find(&products).Error
	return products, err
}
