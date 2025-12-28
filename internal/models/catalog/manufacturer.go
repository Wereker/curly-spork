package models

import (
	"fmt"
	"time"

	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

type Manufacturer struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Slug        string         `json:"slug" gorm:"size:100;uniqueIndex"`
	Title       string         `json:"title" gorm:"size:50;not null"`
	Country     string         `json:"country,omitempty" gorm:"size:50"`
	Description string         `json:"description,omitempty" gorm:"type:text"`

	Products    []Product      `json:"products,omitempty" gorm:"foreignKey:ManufacturerID"`

	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Manufacturer) TableName() string {
	return "manufacturers"
}

func (m *Manufacturer) BeforeCreate(tx *gorm.DB) error {
	if m.Slug == "" {
		m.Slug = slug.Make(m.Title)
	}

	var count int64
	baseSlug := m.Slug
	counter := 1

	for {
		if counter > 1 {
			m.Slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		}

		tx.Model(&Manufacturer{}).
			Where("slug = ?", m.Slug).
			Count(&count)

		if count == 0 {
			break
		}

		counter++
	}
	
	return nil
}