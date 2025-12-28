package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"github.com/gosimple/slug"
)

type Category struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Slug        string         `json:"slug" gorm:"size:100;uniqueIndex"`
	Title       string         `json:"title" gorm:"size:50;not null"`
	Description string         `json:"description" gorm:"type:text"`
	ParentID    *uint          `json:"parent_id"`
	Parent      *Category      `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
    Children    []Category     `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	CoverURL    string         `json:"cover_url,omitempty" gorm:"size:255"`

	Products    []Product      `json:"products,omitempty" gorm:"foreignKey:CategoryID"`

	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Category) TableName() string {
	return "categories"
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
    if c.Slug == "" {
        c.Slug = slug.Make(c.Title)
    }
    
    var count int64
    baseSlug := c.Slug
    counter := 1
    
    for {
        if counter > 1 {
            c.Slug = fmt.Sprintf("%s-%d", baseSlug, counter)
        }
        
        tx.Model(&Category{}).
            Where("slug = ? AND id != ?", c.Slug, c.ID).
            Count(&count)
            
        if count == 0 {
            break
        }
        counter++
    }
    
    return nil
}

func (c *Category) BeforeSave(tx *gorm.DB) error {
    if c.ParentID != nil {
        if err := c.CheckCircularReference(tx, *c.ParentID); err != nil {
            return err
        }
    }
    return nil
}

func (c *Category) CheckCircularReference(tx *gorm.DB, parentID uint) error {
    if c.ID == parentID {
        return fmt.Errorf("category cannot be parent of itself")
    }
    
    currentParentID := parentID
    for currentParentID != 0 {
        var parent Category
        if err := tx.First(&parent, currentParentID).Error; err != nil {
            return err
        }
        
        if parent.ParentID != nil && *parent.ParentID == c.ID {
            return fmt.Errorf("circular reference detected")
        }
        
        if parent.ParentID == nil {
            break
        }
        currentParentID = *parent.ParentID
    }
    
    return nil
}