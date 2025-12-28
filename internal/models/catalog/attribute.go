package models

import (
	"fmt"
	"time"

	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

type AttributeType string

const (
    AttributeTypeText        AttributeType = "text"
    AttributeTypeNumber      AttributeType = "number"
    AttributeTypeSelect      AttributeType = "select"
    AttributeTypeMultiSelect AttributeType = "multi_select"
)

type Attribute struct {
	ID            uint           `json:"id" gorm:"primaryKey,autoIncrement"`
	Slug          string         `json:"slug" gorm:"size:100;uniqueIndex"`
	Title         string         `json:"title" gorm:"size:100"`
	AttributeType AttributeType  `json:"attribute_type" gorm:"type:varchar(20);default:'text'"`

	AttributeValues []AttributeValue `json:"attribute_values,omitempty" gorm:"foreignKey:AttributeID"`

	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Attribute) TableName() string {
	return "attributes"
}

func (a *Attribute) BeforeCreate(tx *gorm.DB) error {
	if a.Slug == "" {
		a.Slug = slug.Make(a.Title)
	}

	var count int64
	baseSlug := a.Slug
	counter := 1

	for {
		if counter > 1 {
			a.Slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		}

		tx.Model(&Attribute{}).
			Where("slug = ?", a.Slug).
			Count(&count)

		if count == 0 {
			break
		}

		counter++
	}
	
	return nil
}